// cmd/keymap.go
// TUI 키입력 의미 계층 - 물리 키를 액션으로 변환

package cmd

import tea "github.com/charmbracelet/bubbletea"

// Action은 키 입력의 의미 단위
type Action int

const (
	ActionNone Action = iota

	// 전역
	ActionQuit // ctrl+c: 프로그램 즉시 종료

	// 화면 탐색
	ActionBack   // esc: 현재 컨텍스트 한 단계 뒤로
	ActionExit   // q:   현재 기능 전체 종료, 메인 메뉴 복귀
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
// 단, ActionBack/ActionExit/ActionConfirm/ActionDelete/ActionEdit/ActionRefresh는 직접 처리합니다.
func ListAction(msg tea.KeyMsg) Action {
	switch msg.String() {
	case "esc":
		return ActionBack
	case "q":
		return ActionExit
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
	case "q":
		return ActionExit
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
// esc = 한 단계 이전 필드, q = 입력 전체 취소 후 메뉴 복귀
func FormAction(msg tea.KeyMsg) Action {
	switch msg.String() {
	case "esc":
		return ActionBack
	case "q":
		return ActionExit
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
// enter/esc = 직전 안전 상태 복귀, q = 메뉴 복귀
func ErrorAction(msg tea.KeyMsg) Action {
	switch msg.String() {
	case "enter", "esc":
		return ActionBack
	case "q":
		return ActionExit
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
