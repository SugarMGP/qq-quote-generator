package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func fixtureDataURI(t *testing.T, width, height int) string {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 17), G: uint8(y * 23), B: 120, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

func TestResourceLoaderLoadsDataURI(t *testing.T) {
	got := NewResourceLoader(&http.Client{Timeout: time.Second}, 1<<20).Load(context.Background(), fixtureDataURI(t, 8, 4))
	if got.Missing || got.Width != 8 || got.Height != 4 {
		t.Fatalf("loaded image = %#v", got)
	}
}

func TestResourceLoaderMarksFailedRemoteImageMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "no", http.StatusNotFound)
	}))
	defer server.Close()

	got := NewResourceLoader(server.Client(), 1<<20).Load(context.Background(), server.URL)
	if !got.Missing || got.Err == nil {
		t.Fatalf("result = %#v", got)
	}
}

func TestResourceLoaderRejectsOversizedImage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(bytes.Repeat([]byte{1}, 33))
	}))
	defer server.Close()

	got := NewResourceLoader(server.Client(), 32).Load(context.Background(), server.URL)
	if !got.Missing || got.Err == nil {
		t.Fatalf("result = %#v", got)
	}
}
