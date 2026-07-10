package style

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	maxImageBytes     int64  = 32 << 20
	maxImageDimension        = 16_384
	maxImagePixels    uint64 = 32 * 1024 * 1024
	maxImageRedirects        = 5
	maxImageDecoders         = 2
)

var (
	errBlockedImageAddress = errors.New("style: image URL resolves to a non-public address")
	blockedImagePrefixes   = [...]netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/8"),
		netip.MustParsePrefix("100.64.0.0/10"),
		netip.MustParsePrefix("192.0.0.0/24"),
		netip.MustParsePrefix("192.0.2.0/24"),
		netip.MustParsePrefix("198.18.0.0/15"),
		netip.MustParsePrefix("198.51.100.0/24"),
		netip.MustParsePrefix("203.0.113.0/24"),
		netip.MustParsePrefix("192.88.99.0/24"),
		netip.MustParsePrefix("168.63.129.16/32"),
		netip.MustParsePrefix("240.0.0.0/4"),
		netip.MustParsePrefix("64:ff9b::/96"),
		netip.MustParsePrefix("64:ff9b:1::/48"),
		netip.MustParsePrefix("100::/64"),
		netip.MustParsePrefix("2001::/23"),
		netip.MustParsePrefix("2001:db8::/32"),
		netip.MustParsePrefix("2002::/16"),
		netip.MustParsePrefix("2620:4f:8000::/48"),
		netip.MustParsePrefix("3fff::/20"),
		netip.MustParsePrefix("5f00::/16"),
		netip.MustParsePrefix("fec0::/10"),
	}
	imageDecodeSlots = make(chan struct{}, maxImageDecoders)
)

type imageResolver interface {
	LookupNetIP(ctx context.Context, network, host string) ([]netip.Addr, error)
}

type imageDialContext func(ctx context.Context, network, address string) (net.Conn, error)

var defaultImageHTTPClient = newImageHTTPClient(net.DefaultResolver, (&net.Dialer{
	Timeout:   10 * time.Second,
	KeepAlive: 30 * time.Second,
}).DialContext)

// ImageFillFit 定义图片背景的缩放模式。
type ImageFillFit int

const (
	ImageFillContain ImageFillFit = iota // 1
	ImageFillCover                       // 2
	ImageFillFill                        // 3
	ImageFillNone                        // 4
)

// ImageFill 描述作为背景使用的图片及其缩放模式。
// Src 为已解码的 image.Image；Fit 决定图片如何填满容器区域。
type ImageFill struct {
	Src image.Image
	Fit ImageFillFit
}

// DecodeImageURL 从 HTTP(S) URL 下载并解码图片。
func DecodeImageURL(rawURL string) (image.Image, error) {
	return decodeImageURL(rawURL, defaultImageHTTPClient)
}

func decodeImageURL(rawURL string, client *http.Client) (image.Image, error) {
	validatedURL, err := validateImageURL(rawURL)
	if err != nil {
		return nil, err
	}

	resp, err := client.Get(validatedURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("style: image request failed: HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength > maxImageBytes {
		return nil, fmt.Errorf("style: image exceeds %d byte limit", maxImageBytes)
	}

	return decodeImageFromReader(resp.Body)
}

// DecodeImageFile 从本地文件路径解码图片。
func DecodeImageFile(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return decodeImageFromReader(f)
}

// LoadImage 自动检测 URL 或文件路径，解码图片。
// 以 "http://" 或 "https://" 开头视为 URL，否则视为本地文件。
func LoadImage(src string) (image.Image, error) {
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		return DecodeImageURL(src)
	}
	return DecodeImageFile(src)
}

func decodeImageFromReader(r io.Reader) (image.Image, error) {
	imageDecodeSlots <- struct{}{}
	defer func() { <-imageDecodeSlots }()

	data, err := readImageBytes(r, maxImageBytes)
	if err != nil {
		return nil, err
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if err := validateImageDimensions(config.Width, config.Height); err != nil {
		return nil, err
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	bounds := img.Bounds()
	if err := validateImageDimensions(bounds.Dx(), bounds.Dy()); err != nil {
		return nil, err
	}
	return img, nil
}

func readImageBytes(r io.Reader, limit int64) ([]byte, error) {
	limited := &io.LimitedReader{R: r, N: limit + 1}
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("style: image exceeds %d byte limit", limit)
	}
	return data, nil
}

func validateImageDimensions(width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("style: image has invalid dimensions %dx%d", width, height)
	}
	if width > maxImageDimension || height > maxImageDimension {
		return fmt.Errorf("style: image dimensions %dx%d exceed %d pixel side limit", width, height, maxImageDimension)
	}
	if uint64(width)*uint64(height) > maxImagePixels {
		return fmt.Errorf("style: image dimensions %dx%d exceed %d pixel limit", width, height, maxImagePixels)
	}
	return nil
}

func newImageHTTPClient(resolver imageResolver, dial imageDialContext) *http.Client {
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           validatedImageDialContext(resolver, dial),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          4,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
	return &http.Client{
		Transport:     transport,
		Timeout:       15 * time.Second,
		CheckRedirect: checkImageRedirect,
	}
}

func checkImageRedirect(req *http.Request, via []*http.Request) error {
	if len(via) > maxImageRedirects {
		return fmt.Errorf("style: image request exceeded %d redirects", maxImageRedirects)
	}
	_, err := validateImageURL(req.URL.String())
	return err
}

func validateImageURL(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("style: invalid image URL: %w", err)
	}
	if parsed.Opaque != "" || parsed.Host == "" || parsed.Hostname() == "" {
		return nil, errors.New("style: image URL must be an absolute HTTP(S) URL")
	}
	switch strings.ToLower(parsed.Scheme) {
	case "http", "https":
	default:
		return nil, errors.New("style: image URL must use HTTP or HTTPS")
	}

	host := strings.TrimSuffix(strings.ToLower(parsed.Hostname()), ".")
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return nil, fmt.Errorf("%w: %s", errBlockedImageAddress, parsed.Hostname())
	}
	if addr, err := netip.ParseAddr(host); err == nil && isBlockedImageAddress(addr) {
		return nil, fmt.Errorf("%w: %s", errBlockedImageAddress, addr)
	}
	return parsed, nil
}

func validatedImageDialContext(resolver imageResolver, dial imageDialContext) imageDialContext {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("style: invalid image server address: %w", err)
		}

		addresses, err := resolveImageAddresses(ctx, resolver, host)
		if err != nil {
			return nil, err
		}
		for _, addr := range addresses {
			if isBlockedImageAddress(addr) {
				return nil, fmt.Errorf("%w: %s", errBlockedImageAddress, addr)
			}
		}

		var dialErrors []error
		for _, addr := range addresses {
			conn, err := dial(ctx, network, net.JoinHostPort(addr.String(), port))
			if err == nil {
				return conn, nil
			}
			dialErrors = append(dialErrors, err)
		}
		return nil, fmt.Errorf("style: image server connection failed: %w", errors.Join(dialErrors...))
	}
}

func resolveImageAddresses(ctx context.Context, resolver imageResolver, host string) ([]netip.Addr, error) {
	if addr, err := netip.ParseAddr(host); err == nil {
		return []netip.Addr{addr.Unmap()}, nil
	}

	addresses, err := resolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("style: resolve image server %q: %w", host, err)
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("style: resolve image server %q: no addresses", host)
	}
	for index := range addresses {
		addresses[index] = addresses[index].Unmap()
	}
	return addresses, nil
}

func isBlockedImageAddress(addr netip.Addr) bool {
	addr = addr.Unmap()
	if !addr.IsValid() || !addr.IsGlobalUnicast() || addr.IsPrivate() || addr.IsLoopback() ||
		addr.IsUnspecified() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsMulticast() {
		return true
	}
	for _, prefix := range blockedImagePrefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}
