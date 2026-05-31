package style

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeImageURLRejectsHTTPErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "missing", http.StatusNotFound)
	}))
	defer server.Close()

	img, err := DecodeImageURL(server.URL)
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
