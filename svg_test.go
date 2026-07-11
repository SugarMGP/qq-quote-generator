package main

import (
	"bytes"
	"context"
	"testing"
)

func layoutWithImage(t *testing.T, text string) CardLayout {
	t.Helper()
	image := NewResourceLoader(nil, 1<<20).Load(context.Background(), fixtureDataURI(t, 8, 4))
	return testLayoutEngine(t).Layout([]PreparedMessage{{Nickname: "张三", Avatar: image, Segments: []PreparedSegment{{Type: "text", Text: text}, {Type: "image", Image: image}}}}, darkTheme)
}

func TestSVGBuilderProducesSelfContainedDocument(t *testing.T) {
	svg, err := (SVGBuilder{}).Build(layoutWithImage(t, "你好"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`<svg xmlns="http://www.w3.org/2000/svg"`, `<clipPath`, `data:image/png;base64,`} {
		if !bytes.Contains(svg, []byte(want)) {
			t.Fatalf("missing %q in %s", want, svg)
		}
	}
}

func TestSVGBuilderEscapesUserText(t *testing.T) {
	svg, err := (SVGBuilder{}).Build(layoutWithImage(t, `<script>&`))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(svg, []byte("<script>")) || !bytes.Contains(svg, []byte("&lt;script&gt;&amp;")) {
		t.Fatalf("unsafe SVG: %s", svg)
	}
}

func TestSVGBuilderUsesThemeColors(t *testing.T) {
	for _, theme := range []Theme{lightTheme, darkTheme} {
		card := testLayoutEngine(t).Layout([]PreparedMessage{{Segments: []PreparedSegment{{Type: "text", Text: "x"}}}}, theme)
		svg, err := (SVGBuilder{}).Build(card)
		if err != nil {
			t.Fatal(err)
		}
		for _, color := range []string{theme.CardBG, theme.AvatarBG, theme.NameColor, theme.BubbleBG, theme.TextColor} {
			if !bytes.Contains(svg, []byte(color)) {
				t.Fatalf("missing %s", color)
			}
		}
	}
}
