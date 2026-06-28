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
func newMenuSubModel(cfg *config.Config) *menuSubModel {
	items := []list.Item{
		simpleItem{"거래내역 조회"},
		simpleItem{"거래 입력"},
		simpleItem{"자산/부채 현황"},
		simpleItem{"섹션 변경"},
		simpleItem{"사용자 정보"},
		simpleItem{"섹션 관리"},
		simpleItem{"항목 관리"},
		simpleItem{"흐름 분석"},
		simpleItem{"카드 관리"},
		simpleItem{"예산·목표"},
	}
	return &menuSubModel{
		list: newCompactList(items, 40, len(items)+2),
		cfg:  cfg,
	}
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
		// 메인 메뉴는 최상위: esc/q 모두 프로그램 종료
		switch ListAction(msg) {
		case ActionBack, ActionExit:
			return m, tea.Quit
		case ActionConfirm:
			return m, func() tea.Msg { return menuSelectionMsg{selection: m.list.Index()} }
		}
		if idx, ok := NumberAction(msg); ok && idx < len(m.list.Items()) {
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
	content += helpStyle.Render(fmt.Sprintf("후잉 CLI v%s", Version)) + "\n\n"
	content += headerStyle.Render("메인 메뉴") + "\n\n"
	content += m.list.View() + "\n"
	content += "\n" + helpStyle.Render("[↑/↓/j/k] 이동  [1-9] 번호 선택  [Enter] 확인  [Esc/q] 종료")
	return content
}
