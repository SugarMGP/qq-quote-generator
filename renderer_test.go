package main

import (
	"strings"
	"testing"
)

func TestTemplateUsesCircularAvatarAndContentWidth(t *testing.T) {
	if !strings.Contains(quoteHTML, "border-radius: 50%;") {
		t.Fatal("avatar should render as a circle")
	}
	if strings.Contains(quoteHTML, "min-width: 300px;") {
		t.Fatal("short quotes should not keep a fixed 300px minimum width")
	}
}
