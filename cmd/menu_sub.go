// cmd/menu_sub.go
// 메인 메뉴 - bubbletea 서브 모델

package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"whoo-cli/config"
)

// menuSelectionMsg는 메뉴 선택 메시지
type menuSelectionMsg struct {
	selection int
}

// menuSubModel은 메인 메뉴 서브 모델
type menuSubModel struct {
	list list.Model
	cfg  *config.Config
}

// newMenuSubModel은 새로운 메뉴 모델을 생성
// selected는 복원할 커서(0-based). 범위 밖이면 보정된다.
func newMenuSubModel(cfg *config.Config, selected int) *menuSubModel {
	items := []list.Item{
		simpleItem{"거래내역 조회"},
		simpleItem{"거래 입력"},
		simpleItem{"자산/부채 현황"},
		simpleItem{"자주입력"},
		simpleItem{"월별입력"},
		simpleItem{"항목 관리"},
		simpleItem{"흐름 분석"},
		simpleItem{"카드 관리"},
		simpleItem{"예산/목표"},
		simpleItem{"섹션"}, // 변경/관리 허브
		simpleItem{"사용자 정보"},
	}
	m := &menuSubModel{
		list: newCompactList(items, 40, len(items)+2),
		cfg:  cfg,
	}
	selectListIndex(&m.list, selected)
	return m
}

// cursorIndex는 현재 선택 커서
func (m *menuSubModel) cursorIndex() int {
	return m.list.Index()
}

func (m *menuSubModel) Init() tea.Cmd {
	return nil
}

func (m *menuSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if GlobalAction(msg) == ActionQuit {
			return m, tea.Quit
		}
		// 메인 메뉴만 q 지원 (기능 화면은 esc만). esc/q 모두 종료.
		if msg.String() == "q" {
			return m, tea.Quit
		}
		switch ListAction(msg) {
		case ActionBack:
			return m, tea.Quit
		case ActionConfirm:
			return m, func() tea.Msg { return menuSelectionMsg{selection: m.list.Index()} }
		}
		// 1–9 + a/b/c… (10번째 이후). 유효 인덱스면 즉시 선택.
		if idx, ok := ItemShortcutAction(msg); ok && idx < len(m.list.Items()) {
			m.list.Select(idx) // 커서도 맞춰 복원 시 동일 위치
			return m, func() tea.Msg { return menuSelectionMsg{selection: idx} }
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *menuSubModel) View() string {
	var content string
	content += titleStyle.Render(banner) + "\n"
	content += helpStyle.Render(fmt.Sprintf("v%s", Version)) + "\n\n"
	content += m.list.View() + "\n"
	n := len(m.list.Items())
	shortcutHelp := "[1-9] 번호 선택"
	if n > 9 {
		last := itemShortcutLabel(n - 1)
		shortcutHelp = fmt.Sprintf("[1-9/a-%s] 바로선택", last)
	}
	content += "\n" + helpStyle.Render(fmt.Sprintf("[↑/↓/j/k] 이동  %s  [Enter] 확인  [Esc/q] 종료", shortcutHelp))
	return content
}
