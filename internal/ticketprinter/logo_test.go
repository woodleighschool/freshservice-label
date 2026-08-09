package ticketprinter

import (
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestNewRendererFromURLFetchesLogoOnce(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "image/png")
		logo := image.NewRGBA(image.Rect(0, 0, 100, 20))
		draw.Draw(logo, logo.Bounds(), image.NewUniform(color.Black), image.Point{}, draw.Src)
		if err := png.Encode(w, logo); err != nil {
			t.Errorf("encode logo: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	renderer, err := NewRendererFromURL(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("create renderer: %v", err)
	}
	for range 2 {
		if _, err := renderer.renderLabel(exampleLabel()); err != nil {
			t.Fatalf("render label: %v", err)
		}
	}
	if got := requests.Load(); got != 1 {
		t.Fatalf("logo requests = %d, want 1", got)
	}
}

func TestNewRendererFromURLRejectsInvalidLogo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("not a PNG"))
	}))
	t.Cleanup(server.Close)

	_, err := NewRendererFromURL(context.Background(), server.URL)
	if err == nil || !strings.Contains(err.Error(), "decode logo PNG") {
		t.Fatalf("NewRendererFromURL error = %v, want PNG decode error", err)
	}
}
