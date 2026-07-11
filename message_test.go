package main

import (
	"reflect"
	"testing"
)

func TestParseMessageFieldPreservesFormats(t *testing.T) {
	tests := []struct {
		name string
		raw  interface{}
		want []MessageSegment
	}{
		{name: "string", raw: "你好", want: []MessageSegment{{Type: "text", Text: "你好"}}},
		{name: "nil", raw: nil, want: nil},
		{name: "number", raw: float64(42), want: []MessageSegment{{Type: "text", Text: "42"}}},
		{name: "segments", raw: []interface{}{map[string]interface{}{"type": "image", "kind": "emoji", "url": "data:image/png;base64,AA=="}}, want: []MessageSegment{{Type: "image", Kind: "emoji", URL: "data:image/png;base64,AA=="}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMessageField(tt.raw)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestSafeImageURLKeepsOnlySupportedSources(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{raw: " https://example.com/a.png ", want: "https://example.com/a.png"},
		{raw: "data:image/png;base64,AA==", want: "data:image/png;base64,AA=="},
		{raw: "javascript:alert(1)", want: ""},
	}

	for _, tt := range tests {
		if got := safeImageURL(tt.raw); string(got) != tt.want {
			t.Fatalf("safeImageURL(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

func TestThemeForHourReturnsPalette(t *testing.T) {
	if got := themeForHour(12); got != lightTheme {
		t.Fatalf("day theme = %#v, want %#v", got, lightTheme)
	}
	if got := themeForHour(23); got != darkTheme {
		t.Fatalf("night theme = %#v, want %#v", got, darkTheme)
	}
}
