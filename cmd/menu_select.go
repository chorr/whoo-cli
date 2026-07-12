// cmd/menu_select.go
// 세로 번호 메뉴 공통 렌더/키 처리 (기능 진입 1뎁스 UX 통일)

package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// renderNumberedMenuLines는 "1. label" 형식 세로 메뉴를 렌더한다.
// 10번째부터는 a, b, c… (itemShortcutLabel).
func renderNumberedMenuLines(labels []string, cursor int) string {
	var b strings.Builder
	for i, label := range labels {
		line := fmt.Sprintf(" %s. %s", itemShortcutLabel(i), label)
		if i == cursor {
			b.WriteString(selectedStyle.Render(">"+line) + "\n")
		} else {
			b.WriteString("  " + line + "\n")
		}
	}
	return b.String()
}

// numberedMenuHelp는 항목 수에 맞는 도움말 문자열
func numberedMenuHelp(count int) string {
	if count <= 0 {
		return "[Esc] 뒤로"
	}
	if count <= 9 {
		return fmt.Sprintf("[↑/↓/j/k] 이동  [1-%d] 번호 선택  [Enter] 확인  [Esc] 뒤로", count)
	}
	last := itemShortcutLabel(count - 1)
	return fmt.Sprintf("[↑/↓/j/k] 이동  [1-9/a-%s] 바로선택  [Enter] 확인  [Esc] 뒤로", last)
}

// verticalMenuResult는 세로 번호 메뉴 키 처리 결과
type verticalMenuResult struct {
	Cursor   int
	Confirm  bool // Enter 또는 번호 바로선택
	Back     bool // Esc
	Handled  bool // 이 헬퍼가 처리함 (list.Update 생략용)
}

// handleVerticalMenuKey는 세로 번호 메뉴의 공통 키 처리.
// 번호/알파벳 바로선택 시 Confirm=true (메인 메뉴와 동일하게 즉시 선택).
func handleVerticalMenuKey(msg tea.KeyMsg, cursor, count int) verticalMenuResult {
	if count <= 0 {
		if ListAction(msg) == ActionBack {
			return verticalMenuResult{Cursor: 0, Back: true, Handled: true}
		}
		return verticalMenuResult{Cursor: 0}
	}
	cursor = clampIndex(cursor, count)

	switch ListAction(msg) {
	case ActionBack:
		return verticalMenuResult{Cursor: cursor, Back: true, Handled: true}
	case ActionConfirm:
		return verticalMenuResult{Cursor: cursor, Confirm: true, Handled: true}
	}

	switch msg.String() {
	case "up", "k":
		if cursor > 0 {
			cursor--
		}
		return verticalMenuResult{Cursor: cursor, Handled: true}
	case "down", "j":
		if cursor < count-1 {
			cursor++
		}
		return verticalMenuResult{Cursor: cursor, Handled: true}
	}

	// 1-9 / a-z 바로선택 → 즉시 확정
	if idx, ok := ItemShortcutAction(msg); ok && idx < count {
		return verticalMenuResult{Cursor: idx, Confirm: true, Handled: true}
	}

	return verticalMenuResult{Cursor: cursor, Handled: false}
}

// handleListNumberJump는 bubbles/list에서 번호 키로 커서 이동(또는 즉시 확정)을 처리한다.
// confirmOnNumber=true 이면 번호 선택 즉시 확정(진입 메뉴), false면 커서만 이동.
func handleListNumberJump(msg tea.KeyMsg, count int, confirmOnNumber bool) (idx int, confirm bool, ok bool) {
	if count <= 0 {
		return -1, false, false
	}
	i, yes := ItemShortcutAction(msg)
	if !yes || i >= count {
		return -1, false, false
	}
	return i, confirmOnNumber, true
}
