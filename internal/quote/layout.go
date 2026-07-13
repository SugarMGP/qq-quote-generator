package quote

import (
	"math"
	"strings"
	"unicode"
)

const (
	cardMaxWidth   = 600.0
	cardPadX       = 12.0
	cardPadY       = 16.0
	avatarSize     = 42.0
	rowGap         = 10.0
	rowMargin      = 14.0
	nicknameSize   = 12.0
	nicknameHeight = 16.0
	nicknameMargin = 4.0
	bubblePadX     = 12.0
	bubblePadY     = 8.0
	textSize       = 15.0
	textLineHeight = 24.0
	segmentMargin  = 4.0
)

type Rect struct{ X, Y, W, H float64 }

type PreparedMessage struct {
	Nickname string
	Avatar   LoadedImage
	Segments []PreparedSegment
}

type PreparedSegment struct {
	Type, Kind, Text string
	Image            LoadedImage
}

type TextLine struct {
	Text     string
	Width    float64
	Baseline float64
}
type TextLayout struct {
	Rect     Rect
	FontSize float64
	Lines    []TextLine
}
type SegmentLayout struct {
	Type, Kind, DataURI string
	Rect                Rect
	Lines               []TextLine
}
type BubbleLayout struct {
	Rect Rect
}
type RowLayout struct {
	Height        float64
	Avatar        Rect
	AvatarDataURI string
	Nickname      TextLayout
	Bubble        BubbleLayout
	Segments      []SegmentLayout
}
type CardLayout struct {
	Width, Height float64
	Theme         Theme
	FontFamily    string
	Rows          []RowLayout
}

type LayoutEngine struct{ fonts *FontManager }

func NewLayoutEngine(fonts *FontManager) *LayoutEngine { return &LayoutEngine{fonts: fonts} }

func containSize(width, height, maxWidth, maxHeight float64) (float64, float64) {
	if width <= 0 || height <= 0 {
		return 0, 0
	}
	scale := math.Min(1, math.Min(maxWidth/width, maxHeight/height))
	return width * scale, height * scale
}

func (e *LayoutEngine) Layout(messages []PreparedMessage, theme Theme) CardLayout {
	card := CardLayout{Width: cardPadX * 2, Height: cardPadY * 2, Theme: theme, FontFamily: e.fonts.Family()}
	availableContent := cardMaxWidth - cardPadX*2 - avatarSize - rowGap
	maxRowWidth := 0.0
	rows := make([]RowLayout, 0, len(messages))
	y := cardPadY
	for _, message := range messages {
		row := e.layoutRow(message, theme, availableContent, y)
		if e.messageRequiresMaxWidth(message, availableContent-bubblePadX*2) {
			row.Bubble.Rect.W = availableContent
		}
		rows = append(rows, row)
		rowWidth := avatarSize + rowGap + math.Max(row.Nickname.Rect.W, row.Bubble.Rect.W)
		maxRowWidth = math.Max(maxRowWidth, rowWidth)
		y += row.Height + rowMargin
	}
	if len(rows) > 0 {
		card.Height = y - rowMargin + cardPadY
		card.Width = math.Min(cardMaxWidth, cardPadX*2+maxRowWidth)
	}
	card.Rows = rows
	return card
}

func (e *LayoutEngine) messageRequiresMaxWidth(message PreparedMessage, maxInnerWidth float64) bool {
	for _, segment := range message.Segments {
		if segment.Type != "text" {
			continue
		}
		for _, paragraph := range strings.Split(segment.Text, "\n") {
			if e.fonts.Measure(paragraph, textSize) > maxInnerWidth {
				return true
			}
		}
	}
	return false
}

func (e *LayoutEngine) layoutRow(message PreparedMessage, theme Theme, maxContentWidth, y float64) RowLayout {
	contentX := cardPadX + avatarSize + rowGap
	nicknameLine := e.textLine(message.Nickname, nicknameSize, nicknameHeight)
	nicknameWidth := nicknameLine.Width
	maxInnerWidth := maxContentWidth - bubblePadX*2
	segments := make([]SegmentLayout, 0, len(message.Segments))
	innerWidth, innerHeight := 0.0, 0.0
	for index, segment := range message.Segments {
		if index > 0 {
			innerHeight += segmentMargin
		}
		layout := SegmentLayout{Type: segment.Type, Kind: segment.Kind, DataURI: segment.Image.DataURI}
		if segment.Type == "text" {
			layout.Lines = e.wrapText(segment.Text, maxInnerWidth, textSize, textLineHeight)
			for _, line := range layout.Lines {
				innerWidth = math.Max(innerWidth, line.Width)
			}
			layout.Rect = Rect{W: innerWidth, H: math.Max(1, float64(len(layout.Lines))) * textLineHeight}
		} else if segment.Type == "image" {
			maxW, maxH := 200.0, 200.0
			if segment.Kind == "emoji" {
				maxW, maxH = 42, 42
			}
			if segment.Kind == "sticker" {
				maxW, maxH = 96, 96
			}
			w, h := containSize(float64(segment.Image.Width), float64(segment.Image.Height), maxW, maxH)
			if segment.Kind == "emoji" {
				w, h = 42, 42
			}
			layout.Rect.W, layout.Rect.H = w, h
			innerWidth = math.Max(innerWidth, w)
		}
		layout.Rect.X = contentX + bubblePadX
		layout.Rect.Y = y + nicknameHeight + nicknameMargin + bubblePadY + innerHeight
		innerHeight += layout.Rect.H
		segments = append(segments, layout)
	}
	bubbleWidth := innerWidth + bubblePadX*2
	bubbleHeight := innerHeight + bubblePadY*2
	bubbleY := y + nicknameHeight + nicknameMargin
	rowHeight := math.Max(avatarSize, nicknameHeight+nicknameMargin+bubbleHeight)
	return RowLayout{
		Height: rowHeight,
		Avatar: Rect{X: cardPadX, Y: y, W: avatarSize, H: avatarSize}, AvatarDataURI: message.Avatar.DataURI,
		Nickname: TextLayout{Rect: Rect{X: contentX, Y: y, W: nicknameWidth, H: nicknameHeight}, FontSize: nicknameSize, Lines: []TextLine{nicknameLine}},
		Bubble:   BubbleLayout{Rect: Rect{X: contentX, Y: bubbleY, W: bubbleWidth, H: bubbleHeight}},
		Segments: segments,
	}
}

func (e *LayoutEngine) wrapText(text string, maxWidth, size, lineHeight float64) []TextLine {
	paragraphs := strings.Split(text, "\n")
	lines := make([]TextLine, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		if paragraph == "" {
			lines = append(lines, e.textLine("", size, lineHeight))
			continue
		}
		current := ""
		for _, token := range lineTokens(paragraph) {
			if current != "" && e.fonts.Measure(current+token, size) > maxWidth {
				lines = append(lines, e.textLine(current, size, lineHeight))
				current = ""
			}
			if e.fonts.Measure(token, size) <= maxWidth {
				current += token
				continue
			}
			for _, char := range token {
				if current != "" && e.fonts.Measure(current+string(char), size) > maxWidth {
					lines = append(lines, e.textLine(current, size, lineHeight))
					current = ""
				}
				current += string(char)
			}
		}
		lines = append(lines, e.textLine(current, size, lineHeight))
	}
	return lines
}

func (e *LayoutEngine) textLine(text string, size, lineHeight float64) TextLine {
	metrics := e.fonts.measureLine(text, size)
	return TextLine{
		Text:     text,
		Width:    metrics.Width,
		Baseline: (lineHeight-(metrics.Ascent-metrics.Descent))/2 + metrics.Ascent,
	}
}

func lineTokens(text string) []string {
	var tokens []string
	var latin []rune
	flush := func() {
		if len(latin) > 0 {
			tokens = append(tokens, string(latin))
			latin = nil
		}
	}
	for _, char := range text {
		if char <= unicode.MaxASCII && (unicode.IsLetter(char) || unicode.IsDigit(char)) {
			latin = append(latin, char)
			continue
		}
		flush()
		tokens = append(tokens, string(char))
	}
	flush()
	return tokens
}
