// cmd/listutil.go
// bubbles/list 공통 헬퍼 — 콤팩트 목록 팩토리 및 공유 타입

package cmd

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// backspaceRunes는 UTF-8 안전 backspace (룬 단위 삭제)
func backspaceRunes(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	return string(runes[:len(runes)-1])
}

// addMonthsYYYYMM는 YYYYMM 정수에 months를 더한 값을 반환한다
func addMonthsYYYYMM(ym, months int) int {
	y := ym / 100
	m := ym % 100
	m += months
	for m > 12 {
		m -= 12
		y++
	}
	for m < 1 {
		m += 12
		y--
	}
	return y*100 + m
}

// firstOfMonth는 해당 시각이 속한 달의 1일을 반환
func firstOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

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
	// 1–9 이후 항목은 a, b, c… 표기 (숫자 키 한계)
	line := fmt.Sprintf(" %s. %s", itemShortcutLabel(index), title)
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
	// 페이지네이션 UI를 끄지 않으면 높이 1줄이 페이지 계산에서 빠져
	// 항목 5개·높이 5여도 PerPage=4 → 불필요한 페이지 스크롤이 생긴다.
	l.SetShowPagination(false)
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

// listViewportHeight는 터미널 높이와 고정 크롬(제목·도움말)을 감안한
// bubbles/list 뷰포트 높이를 계산한다.
// height를 항목 수와 같게 잡으면 화면을 넘기고 스크롤이 깨지므로,
// 뷰포트는 항상 화면 안으로 제한한다.
func listViewportHeight(termHeight, chrome, itemCount int) int {
	if termHeight <= 0 {
		termHeight = 24
	}
	if chrome < 0 {
		chrome = 0
	}
	h := termHeight - chrome
	if h < 5 {
		h = 5
	}
	if itemCount <= 0 {
		return 1
	}
	if itemCount < h {
		return itemCount
	}
	return h
}

// listViewportWidth는 목록 너비 (좌우 여백 고려)
func listViewportWidth(termWidth, margin int) int {
	if termWidth <= 0 {
		return 60
	}
	if margin < 0 {
		margin = 0
	}
	w := termWidth - margin
	if w < 20 {
		w = 20
	}
	return w
}
