// cmd/keymap_test.go

package cmd

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// makeKey는 테스트용 tea.KeyMsg를 생성합니다.
func makeKey(s string) tea.KeyMsg {
	switch s {
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	default:
		runes := []rune(s)
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: runes}
	}
}

func TestGlobalAction(t *testing.T) {
	tests := []struct {
		key  string
		want Action
	}{
		{"ctrl+c", ActionQuit},
		{"q", ActionNone},
		{"esc", ActionNone},
		{"enter", ActionNone},
	}
	for _, tt := range tests {
		got := GlobalAction(makeKey(tt.key))
		if got != tt.want {
			t.Errorf("GlobalAction(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestListAction(t *testing.T) {
	tests := []struct {
		key  string
		want Action
	}{
		{"esc", ActionBack},
		{"q", ActionExit},
		{"enter", ActionConfirm},
		{"d", ActionDelete},
		{"e", ActionEdit},
		{"r", ActionRefresh},
		{"j", ActionNone},
		{"k", ActionNone},
		{"up", ActionNone},
		{"down", ActionNone},
	}
	for _, tt := range tests {
		got := ListAction(makeKey(tt.key))
		if got != tt.want {
			t.Errorf("ListAction(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestHorizontalSelectAction(t *testing.T) {
	tests := []struct {
		key  string
		want Action
	}{
		{"esc", ActionBack},
		{"q", ActionExit},
		{"enter", ActionConfirm},
		{"left", ActionMoveLeft},
		{"h", ActionMoveLeft},
		{"right", ActionMoveRight},
		{"l", ActionMoveRight},
		{"up", ActionMoveUp},
		{"k", ActionMoveUp},
		{"down", ActionMoveDown},
		{"j", ActionMoveDown},
	}
	for _, tt := range tests {
		got := HorizontalSelectAction(makeKey(tt.key))
		if got != tt.want {
			t.Errorf("HorizontalSelectAction(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestConfirmAction(t *testing.T) {
	tests := []struct {
		key  string
		want Action
	}{
		{"y", ActionConfirm},
		{"Y", ActionConfirm},
		{"enter", ActionNone},
		{"n", ActionBack},
		{"N", ActionBack},
		{"esc", ActionBack},
		{"q", ActionNone},
	}
	for _, tt := range tests {
		got := ConfirmAction(makeKey(tt.key))
		if got != tt.want {
			t.Errorf("ConfirmAction(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestErrorAction(t *testing.T) {
	tests := []struct {
		key  string
		want Action
	}{
		{"enter", ActionBack},
		{"esc", ActionBack},
		{"q", ActionExit},
		{"ctrl+c", ActionNone},
	}
	for _, tt := range tests {
		got := ErrorAction(makeKey(tt.key))
		if got != tt.want {
			t.Errorf("ErrorAction(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestNumberAction(t *testing.T) {
	tests := []struct {
		key      string
		wantIdx  int
		wantOk   bool
	}{
		{"1", 0, true},
		{"5", 4, true},
		{"9", 8, true},
		{"0", -1, false},
		{"a", -1, false},
		{"enter", -1, false},
	}
	for _, tt := range tests {
		idx, ok := NumberAction(makeKey(tt.key))
		if ok != tt.wantOk || idx != tt.wantIdx {
			t.Errorf("NumberAction(%q) = (%d, %v), want (%d, %v)", tt.key, idx, ok, tt.wantIdx, tt.wantOk)
		}
	}
}
