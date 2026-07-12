package cmd

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestRenderNumberedMenuLines(t *testing.T) {
	s := renderNumberedMenuLines([]string{"자산", "부채", "자본"}, 1)
	if !strings.Contains(s, "1. 자산") || !strings.Contains(s, "2. 부채") {
		t.Fatal(s)
	}
	if !strings.Contains(s, ">") {
		t.Fatal("cursor marker missing")
	}
}

func TestHandleVerticalMenuKeyNumberSelect(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	r := handleVerticalMenuKey(msg, 0, 5)
	if !r.Confirm || r.Cursor != 1 {
		t.Fatalf("%+v", r)
	}
}

func TestHandleVerticalMenuKeyNav(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}}
	r := handleVerticalMenuKey(msg, 0, 3)
	if r.Cursor != 1 || r.Confirm {
		t.Fatalf("%+v", r)
	}
}
