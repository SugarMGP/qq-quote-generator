package main

import (
	_ "embed"
	"fmt"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

//go:embed assets/fonts/NotoSansCJKsc-Regular.otf
var embeddedFont []byte

type FontManager struct {
	font *sfnt.Font
}

func NewFontManager(data []byte) (*FontManager, error) {
	parsed, err := sfnt.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parse font: %w", err)
	}
	return &FontManager{font: parsed}, nil
}

func (m *FontManager) Measure(text string, size float64) float64 {
	var buffer sfnt.Buffer
	ppem := fixed.Int26_6(size * 64)
	var total fixed.Int26_6
	var previous sfnt.GlyphIndex
	hasPrevious := false
	for _, char := range text {
		glyph, err := m.font.GlyphIndex(&buffer, char)
		if err != nil {
			continue
		}
		if hasPrevious {
			kern, err := m.font.Kern(&buffer, previous, glyph, ppem, font.HintingNone)
			if err == nil {
				total += kern
			}
		}
		advance, err := m.font.GlyphAdvance(&buffer, glyph, ppem, font.HintingNone)
		if err == nil {
			total += advance
		}
		previous = glyph
		hasPrevious = true
	}
	return float64(total) / 64
}
