// cmd/listutil.go
// bubbles/list 공통 헬퍼 — 콤팩트 목록 팩토리 및 공유 타입

package cmd

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// ─── 공통 항목 타입 ────────────────────────────────────────────

// simpleItem은 단순 텍스트 목록 항목 (list.DefaultItem 구현)
type simpleItem struct {
	title string
}

func (i simpleItem) Title() string       { return i.title }
func (i simpleItem) Description() string { return "" }
func (i simpleItem) FilterValue() string { return i.title }

// ─── Delegate ─────────────────────────────────────────────────

// numberedDelegate는 "N. label" 형식 한 줄 렌더 delegate
// list.DefaultItem 인터페이스를 구현하는 모든 항목에 적용
type numberedDelegate struct{}

func (d numberedDelegate) Height() int                               { return 1 }
func (d numberedDelegate) Spacing() int                              { return 0 }
func (d numberedDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd  { return nil }
func (d numberedDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	title := ""
	if di, ok := item.(list.DefaultItem); ok {
		title = di.Title()
	}
	line := fmt.Sprintf(" %d. %s", index+1, title)
	if index == m.Index() {
		fmt.Fprint(w, selectedStyle.Render(">"+line))
	} else {
		fmt.Fprint(w, "  "+line)
	}
}

// plainDelegate는 번호 없이 커서만 표시하는 한 줄 렌더 delegate
type plainDelegate struct{}

func (d plainDelegate) Height() int                               { return 1 }
func (d plainDelegate) Spacing() int                              { return 0 }
func (d plainDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd  { return nil }
func (d plainDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	title := ""
	if di, ok := item.(list.DefaultItem); ok {
		title = di.Title()
	}
	if index == m.Index() {
		fmt.Fprint(w, selectedStyle.Render("> "+title))
	} else {
		fmt.Fprint(w, "  "+title)
	}
}

// ─── 팩토리 ───────────────────────────────────────────────────

// newCompactListWith는 커스텀 delegate를 사용하는 콤팩트 list.Model 생성
func newCompactListWith(items []list.Item, delegate list.ItemDelegate, width, height int) list.Model {
	l := list.New(items, delegate, width, height)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.DisableQuitKeybindings()
	return l
}

// newCompactList는 번호 있는 콤팩트 목록 생성 (메뉴/짧은 목록용)
func newCompactList(items []list.Item, width, height int) list.Model {
	return newCompactListWith(items, numberedDelegate{}, width, height)
}

// newPlainList는 번호 없는 콤팩트 목록 생성 (선택 없이 탐색만 필요한 목록용)
func newPlainList(items []list.Item, width, height int) list.Model {
	return newCompactListWith(items, plainDelegate{}, width, height)
}
