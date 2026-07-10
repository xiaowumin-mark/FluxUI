package style

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"image"
	"image/png"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

func TestDecodeImageURLDecodesBoundedImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		if err := png.Encode(w, image.NewNRGBA(image.Rect(0, 0, 2, 3))); err != nil {
			t.Errorf("encode PNG: %v", err)
		}
	}))
	defer server.Close()

	img, err := decodeImageURL("http://image.test/image.png", newTestImageHTTPClient(t, server, publicImageResolver()))
	if err != nil {
		t.Fatalf("DecodeImageURL returned an error: %v", err)
	}
	if got := img.Bounds().Size(); got.X != 2 || got.Y != 3 {
		t.Fatalf("decoded bounds = %v, want 2x3", got)
	}
}

func TestDecodeImageURLRejectsHTTPErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer server.Close()

	img, err := decodeImageURL("http://image.test/missing.png", newTestImageHTTPClient(t, server, publicImageResolver()))
	if err == nil {
		t.Fatal("expected HTTP error status to return an error")
	}
	if img != nil {
		t.Fatal("expected no image on HTTP error status")
	}
	if !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}

func TestDecodeImageURLRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", strconv.FormatInt(maxImageBytes+1, 10))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	img, err := decodeImageURL("http://image.test/large.png", newTestImageHTTPClient(t, server, publicImageResolver()))
	if err == nil || !strings.Contains(err.Error(), "byte limit") {
		t.Fatalf("expected response byte limit error, got image=%v error=%v", img, err)
	}
}

func TestReadImageBytesRejectsUnknownLengthBodyOverLimit(t *testing.T) {
	_, err := readImageBytes(strings.NewReader(strings.Repeat("x", 65)), 64)
	if err == nil || !strings.Contains(err.Error(), "byte limit") {
		t.Fatalf("expected body byte limit error, got %v", err)
	}
}

func TestDecodeImageFromReaderRejectsOversizedDimensions(t *testing.T) {
	tests := []struct {
		name   string
		width  uint32
		height uint32
		match  string
	}{
		{name: "side", width: maxImageDimension + 1, height: 1, match: "side limit"},
		{name: "pixels", width: 8193, height: 8193, match: "pixel limit"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			img, err := decodeImageFromReader(bytes.NewReader(pngHeader(test.width, test.height)))
			if err == nil || !strings.Contains(err.Error(), test.match) {
				t.Fatalf("expected %s error, got image=%v error=%v", test.match, img, err)
			}
		})
	}
}

func TestDecodeImageURLRejectsNonPublicLiteralAddresses(t *testing.T) {
	urls := []string{
		"http://127.0.0.1/image.png",
		"http://10.0.0.1/image.png",
		"http://169.254.169.254/latest/meta-data",
		"http://100.100.100.200/latest/meta-data",
		"http://168.63.129.16/machine/?comp=goalstate",
		"http://[::1]/image.png",
		"http://[fd00::1]/image.png",
		"http://[fe80::1]/image.png",
		"http://[64:ff9b::a00:1]/image.png",
		"http://[2002:0a00:0001::1]/image.png",
		"http://[2001::1]/image.png",
	}

	for _, rawURL := range urls {
		t.Run(rawURL, func(t *testing.T) {
			img, err := DecodeImageURL(rawURL)
			if !errors.Is(err, errBlockedImageAddress) {
				t.Fatalf("expected blocked address error, got image=%v error=%v", img, err)
			}
		})
	}
}

func TestDecodeImageURLRejectsPrivateDNSResultBeforeDial(t *testing.T) {
	resolver := staticImageResolver{
		"internal.test": {netip.MustParseAddr("192.168.1.20")},
	}
	var dialed atomic.Bool
	client := newImageHTTPClient(resolver, func(context.Context, string, string) (net.Conn, error) {
		dialed.Store(true)
		return nil, errors.New("unexpected dial")
	})
	t.Cleanup(client.CloseIdleConnections)

	img, err := decodeImageURL("http://internal.test/image.png", client)
	if !errors.Is(err, errBlockedImageAddress) {
		t.Fatalf("expected blocked DNS address error, got image=%v error=%v", img, err)
	}
	if dialed.Load() {
		t.Fatal("network dial occurred before the resolved address was rejected")
	}
}

func TestDecodeImageURLRejectsRedirectToPrivateAddress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://169.254.169.254/latest/meta-data", http.StatusFound)
	}))
	defer server.Close()

	img, err := decodeImageURL("http://image.test/redirect", newTestImageHTTPClient(t, server, publicImageResolver()))
	if !errors.Is(err, errBlockedImageAddress) {
		t.Fatalf("expected blocked redirect error, got image=%v error=%v", img, err)
	}
}

func TestCheckImageRedirectRejectsTooManyRedirects(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/image.png", nil)
	via := make([]*http.Request, maxImageRedirects+1)
	if err := checkImageRedirect(req, via); err == nil || !strings.Contains(err.Error(), "redirects") {
		t.Fatalf("expected redirect limit error, got %v", err)
	}
}

func TestLoadImageDecodesLocalFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "image.png")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := png.Encode(file, image.NewNRGBA(image.Rect(0, 0, 4, 5))); err != nil {
		file.Close()
		t.Fatalf("encode image: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close image: %v", err)
	}

	img, err := LoadImage(path)
	if err != nil {
		t.Fatalf("LoadImage returned an error: %v", err)
	}
	if got := img.Bounds().Size(); got.X != 4 || got.Y != 5 {
		t.Fatalf("decoded bounds = %v, want 4x5", got)
	}
}

type staticImageResolver map[string][]netip.Addr

func (r staticImageResolver) LookupNetIP(_ context.Context, _, host string) ([]netip.Addr, error) {
	addresses, ok := r[host]
	if !ok {
		return nil, fmt.Errorf("no test address for %s", host)
	}
	return append([]netip.Addr(nil), addresses...), nil
}

func publicImageResolver() staticImageResolver {
	return staticImageResolver{
		"image.test": {netip.MustParseAddr("93.184.216.34")},
	}
}

func newTestImageHTTPClient(t *testing.T, server *httptest.Server, resolver imageResolver) *http.Client {
	t.Helper()
	target := server.Listener.Addr().String()
	dialer := &net.Dialer{}
	client := newImageHTTPClient(resolver, func(ctx context.Context, network, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, target)
	})
	t.Cleanup(client.CloseIdleConnections)
	return client
}

func pngHeader(width, height uint32) []byte {
	var result bytes.Buffer
	result.Write([]byte("\x89PNG\r\n\x1a\n"))

	data := make([]byte, 13)
	binary.BigEndian.PutUint32(data[0:4], width)
	binary.BigEndian.PutUint32(data[4:8], height)
	data[8] = 8
	data[9] = 2
	writePNGChunk(&result, "IHDR", data)
	writePNGChunk(&result, "IEND", nil)
	return result.Bytes()
}

func writePNGChunk(dst *bytes.Buffer, name string, data []byte) {
	_ = binary.Write(dst, binary.BigEndian, uint32(len(data)))
	dst.WriteString(name)
	dst.Write(data)
	checksum := crc32.NewIEEE()
	_, _ = checksum.Write([]byte(name))
	_, _ = checksum.Write(data)
	_ = binary.Write(dst, binary.BigEndian, checksum.Sum32())
}
