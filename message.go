package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"
)

type Theme struct {
	Class     string
	CardBG    string
	AvatarBG  string
	NameColor string
	BubbleBG  string
	TextColor string
}

var (
	lightTheme = Theme{"theme-light", "#f7f8fb", "#d9dee8", "#667085", "#ffffff", "#242937"}
	darkTheme  = Theme{"theme-dark", "#1e1e2e", "#333333", "#7c7f93", "#313244", "#cdd6f4"}
)

func themeForTime(t time.Time) Theme { return themeForHour(t.Hour()) }

func themeForHour(hour int) Theme {
	if hour >= 6 && hour < 18 {
		return lightTheme
	}
	return darkTheme
}

func processMessageSegments(segments []MessageSegment) []processedMessageSegment {
	result := make([]processedMessageSegment, 0, len(segments))
	for _, segment := range segments {
		result = append(result, processedMessageSegment{Type: segment.Type, Kind: segment.Kind, Text: segment.Text, URL: safeImageURL(segment.URL), ImageClass: imageClassForKind(segment.Kind)})
	}
	return result
}

func imageClassForKind(kind string) string {
	switch kind {
	case "emoji":
		return "bubble-img bubble-img-emoji"
	case "sticker":
		return "bubble-img bubble-img-sticker"
	default:
		return "bubble-img"
	}
}

func safeImageURL(raw string) template.URL {
	raw = strings.TrimSpace(raw)
	lower := strings.ToLower(raw)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "data:image/") {
		return template.URL(raw)
	}
	return ""
}

func resolveAvatar(msg Message) string {
	if msg.Avatar != "" {
		return msg.Avatar
	}
	if msg.UserID > 0 {
		return fmt.Sprintf("https://q1.qlogo.cn/g?b=qq&nk=%d&s=100", msg.UserID)
	}
	return ""
}

func parseMessageField(raw interface{}) ([]MessageSegment, error) {
	if raw == nil {
		return nil, nil
	}
	switch v := raw.(type) {
	case string:
		return []MessageSegment{{Type: "text", Text: v}}, nil
	case []interface{}:
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		var segments []MessageSegment
		if err := json.Unmarshal(data, &segments); err != nil {
			return nil, err
		}
		return segments, nil
	default:
		return []MessageSegment{{Type: "text", Text: fmt.Sprintf("%v", v)}}, nil
	}
}
