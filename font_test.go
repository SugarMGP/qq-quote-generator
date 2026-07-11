package main

import "testing"

func TestFontManagerMeasuresDeterministically(t *testing.T) {
	manager, err := NewFontManager(embeddedFont)
	if err != nil {
		t.Fatal(err)
	}
	if manager.Measure("引用测试 ABC", 15) <= manager.Measure("引用", 15) {
		t.Fatal("longer text should have a larger advance")
	}
}

func TestFontManagerRejectsInvalidFont(t *testing.T) {
	if _, err := NewFontManager([]byte("bad")); err == nil {
		t.Fatal("expected invalid font error")
	}
}
