// cmd/keymap.go
// TUI 키입력 의미 계층 - 물리 키를 액션으로 변환

package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Action은 키 입력의 의미 단위
type Action int

const (
	ActionNone Action = iota

	// 전역
	ActionQuit // ctrl+c: 프로그램 즉시 종료

	// 화면 탐색
	ActionBack    // esc: 한 단계 뒤로 / 기능 종료(메뉴 복귀). 기능 화면은 esc만 사용
	ActionExit    // 메인 메뉴 전용: q로 앱 종료 (ListAction 등에서는 매핑하지 않음)
	ActionConfirm // enter: 현재 포커스 항목 확인/적용

	// 커서 이동
	ActionMoveUp    // ↑ / k
	ActionMoveDown  // ↓ / j
	ActionMoveLeft  // ← / h
	ActionMoveRight // → / l

	// 도메인 액션
	ActionEdit    // e
	ActionDelete  // d
	ActionRefresh // r
)

// GlobalAction은 어떤 화면에서든 최우선으로 처리해야 할 전역 액션을 반환합니다.
// ActionQuit이 반환되면 호출자는 즉시 tea.Quit을 반환해야 합니다.
func GlobalAction(msg tea.KeyMsg) Action {
	if msg.String() == "ctrl+c" {
		return ActionQuit
	}
	return ActionNone
}

// ListAction은 목록 화면의 키 입력을 액션으로 변환합니다.
// bubbles/list가 이미 처리하는 ↑/↓/j/k는 ActionNone을 반환합니다
// (list.Update에 그대로 위임).
// q는 기능 화면에서 쓰지 않는다(메인 메뉴만 직접 처리). 뒤로/종료는 esc(ActionBack).
func ListAction(msg tea.KeyMsg) Action {
	switch msg.String() {
	case "esc":
		return ActionBack
	case "enter":
		return ActionConfirm
	case "d":
		return ActionDelete
	case "e":
		return ActionEdit
	case "r":
		return ActionRefresh
	}
	return ActionNone
}

// HorizontalSelectAction은 가로 탭/슬롯 선택 화면의 키 입력을 변환합니다.
func HorizontalSelectAction(msg tea.KeyMsg) Action {
	switch msg.String() {
	case "esc":
		return ActionBack
	case "enter":
		return ActionConfirm
	case "left", "h":
		return ActionMoveLeft
	case "right", "l":
		return ActionMoveRight
	case "up", "k":
		return ActionMoveUp
	case "down", "j":
		return ActionMoveDown
	}
	return ActionNone
}

// FormAction은 폼 입력 화면의 키 입력을 변환합니다.
// esc = 한 단계 이전 / 취소. q는 문자 입력에 쓸 수 있도록 매핑하지 않는다.
func FormAction(msg tea.KeyMsg) Action {
	switch msg.String() {
	case "esc":
		return ActionBack
	case "enter":
		return ActionConfirm
	}
	return ActionNone
}

// ConfirmAction은 확인 다이얼로그의 키 입력을 변환합니다.
// y/Y=승인, n/N/esc=취소. enter는 ActionNone — 파괴적 작업에서 의도치 않은
// 확인을 막기 위해 호출자가 enter를 명시적으로 처리해야 합니다.
func ConfirmAction(msg tea.KeyMsg) Action {
	switch msg.String() {
	case "y", "Y":
		return ActionConfirm
	case "n", "N", "esc":
		return ActionBack
	}
	return ActionNone
}

// ErrorAction은 에러 화면의 키 입력을 변환합니다.
// enter/esc = 직전 안전 상태 복귀 (q 미사용)
func ErrorAction(msg tea.KeyMsg) Action {
	switch msg.String() {
	case "enter", "esc":
		return ActionBack
	}
	return ActionNone
}

// NumberAction은 '1'~'9' 키를 0-based 인덱스로 변환합니다.
// 해당 없으면 (-1, false) 반환.
func NumberAction(msg tea.KeyMsg) (int, bool) {
	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 {
		return -1, false
	}
	r := msg.Runes[0]
	if r >= '1' && r <= '9' {
		return int(r - '1'), true
	}
	return -1, false
}

// ItemShortcutAction은 목록 바로가기를 0-based 인덱스로 변환합니다.
// 1–9 → 0–8, a–z / A–Z → 9–34. 해당 없으면 (-1, false).
// 표시 라벨은 itemShortcutLabel과 짝을 이룬다.
func ItemShortcutAction(msg tea.KeyMsg) (int, bool) {
	if idx, ok := NumberAction(msg); ok {
		return idx, true
	}
	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 {
		return -1, false
	}
	r := msg.Runes[0]
	switch {
	case r >= 'a' && r <= 'z':
		return 9 + int(r-'a'), true
	case r >= 'A' && r <= 'Z':
		return 9 + int(r-'A'), true
	}
	return -1, false
}

// itemShortcutLabel은 0-based 인덱스의 표시 라벨을 반환한다.
// 0–8 → "1"–"9", 9–34 → "a"–"z", 그 외는 숫자 문자열.
func itemShortcutLabel(index int) string {
	if index < 0 {
		return "?"
	}
	if index < 9 {
		return string(rune('1' + index))
	}
	letter := index - 9
	if letter < 26 {
		return string(rune('a' + letter))
	}
	return fmt.Sprintf("%d", index+1)
}
