package cmd

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
)

func TestClampIndex(t *testing.T) {
	if clampIndex(-1, 5) != 0 {
		t.Fatal("negative")
	}
	if clampIndex(0, 5) != 0 {
		t.Fatal("zero")
	}
	if clampIndex(4, 5) != 4 {
		t.Fatal("last")
	}
	if clampIndex(99, 5) != 4 {
		t.Fatal("overflow")
	}
	if clampIndex(1, 0) != 0 {
		t.Fatal("empty")
	}
}

func TestSelectListIndex(t *testing.T) {
	items := []list.Item{simpleItem{"a"}, simpleItem{"b"}, simpleItem{"c"}}
	l := newCompactList(items, 20, 5)
	selectListIndex(&l, 2)
	if l.Index() != 2 {
		t.Fatalf("index=%d", l.Index())
	}
	selectListIndex(&l, 99)
	if l.Index() != 2 {
		t.Fatalf("clamped index=%d", l.Index())
	}
}

func TestNewMenuSubModelRestoresCursor(t *testing.T) {
	m := newMenuSubModel(nil, 5)
	if m.cursorIndex() != 5 {
		t.Fatalf("menu cursor=%d", m.cursorIndex())
	}
}
