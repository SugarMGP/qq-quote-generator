package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

type SVGBuilder struct{}

func (SVGBuilder) Build(card CardLayout) ([]byte, error) {
	var out bytes.Buffer
	fmt.Fprintf(&out, `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s">`, px(card.Width), px(card.Height), px(card.Width), px(card.Height))
	fmt.Fprintf(&out, `<rect width="%s" height="%s" rx="12" fill="%s"/>`, px(card.Width), px(card.Height), card.Theme.CardBG)
	out.WriteString(`<defs>`)
	for index, row := range card.Rows {
		fmt.Fprintf(&out, `<clipPath id="avatar-%d" clipPathUnits="userSpaceOnUse"><circle cx="%s" cy="%s" r="21"/></clipPath>`, index, px(row.Avatar.X+21), px(row.Avatar.Y+21))
	}
	out.WriteString(`</defs>`)
	for index, row := range card.Rows {
		writeRow(&out, index, row, card.Theme)
	}
	out.WriteString(`</svg>`)
	return out.Bytes(), nil
}

func writeRow(out *bytes.Buffer, index int, row RowLayout, theme Theme) {
	fmt.Fprintf(out, `<circle cx="%s" cy="%s" r="21" fill="%s"/>`, px(row.Avatar.X+21), px(row.Avatar.Y+21), theme.AvatarBG)
	if strings.HasPrefix(row.AvatarDataURI, "data:image/") {
		fmt.Fprintf(out, `<image x="%s" y="%s" width="42" height="42" href="%s" preserveAspectRatio="xMidYMid slice" clip-path="url(#avatar-%d)"/>`, px(row.Avatar.X), px(row.Avatar.Y), row.AvatarDataURI, index)
	}
	writeText(out, row.Nickname.Rect.X, row.Nickname.Rect.Y+row.Nickname.FontSize, row.Nickname.FontSize, theme.NameColor, row.Nickname.Lines[0].Text)
	writeBubble(out, row.Bubble.Rect, theme.BubbleBG)
	for _, segment := range row.Segments {
		if segment.Type == "text" {
			for lineIndex, line := range segment.Lines {
				writeText(out, segment.Rect.X, segment.Rect.Y+textSize+float64(lineIndex)*textLineHeight, textSize, theme.TextColor, line.Text)
			}
		} else if segment.Type == "image" && strings.HasPrefix(segment.DataURI, "data:image/") && segment.Rect.W > 0 && segment.Rect.H > 0 {
			radius := "6"
			if segment.Kind == "emoji" || segment.Kind == "sticker" {
				radius = "0"
			}
			fmt.Fprintf(out, `<image x="%s" y="%s" width="%s" height="%s" href="%s" preserveAspectRatio="xMidYMid meet" rx="%s"/>`, px(segment.Rect.X), px(segment.Rect.Y), px(segment.Rect.W), px(segment.Rect.H), segment.DataURI, radius)
		}
	}
}

func writeBubble(out *bytes.Buffer, rect Rect, fill string) {
	x, y, right, bottom := rect.X, rect.Y, rect.X+rect.W, rect.Y+rect.H
	fmt.Fprintf(out, `<path d="M %s %s Q %s %s %s %s H %s Q %s %s %s %s V %s Q %s %s %s %s H %s Q %s %s %s %s Z" fill="%s"/>`,
		px(x+4), px(y), px(x), px(y), px(x), px(y+4), px(right-12), px(right), px(y), px(right), px(y+12), px(bottom-12), px(right), px(bottom), px(right-12), px(bottom), px(x+12), px(x), px(bottom), px(x), px(bottom-12), fill)
}

func writeText(out *bytes.Buffer, x, y, size float64, color, value string) {
	fmt.Fprintf(out, `<text x="%s" y="%s" font-family="Noto Sans CJK SC" font-size="%s" fill="%s">`, px(x), px(y), px(size), color)
	_ = xml.EscapeText(out, []byte(value))
	out.WriteString(`</text>`)
}

func px(value float64) string { return strconv.FormatFloat(value, 'f', -1, 64) }
