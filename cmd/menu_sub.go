// cmd/menu_sub.go
// 메인 메뉴 - bubbletea 서브 모델

package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"whooing-cli/config"
)

// menuSelectionMsg는 메뉴 선택 메시지
type menuSelectionMsg struct {
	selection int
}

// menuSubModel은 메인 메뉴 서브 모델
type menuSubModel struct {
	cursor int
	items  []menuItem
	cfg    *config.Config
}

// menuItem은 메뉴 항목
type menuItem struct {
	id    int
	label string
}

// newMenuSubModel은 새로운 메뉴 모델을 생성
func newMenuSubModel(cfg *config.Config) *menuSubModel {
	items := []menuItem{
		{id: 1, label: "거래내역 조회"},
		{id: 2, label: "거래 입력"},
		{id: 3, label: "자산/부채 현황"},
		{id: 4, label: "섹션 변경"},
		{id: 5, label: "사용자 정보"},
	}

	return &menuSubModel{
		cursor: 0,
		items:  items,
		cfg:    cfg,
	}
}

func (m *menuSubModel) Init() tea.Cmd {
	return nil
}

func (m *menuSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown:
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case tea.KeyRunes:
			if len(msg.Runes) == 1 {
				r := msg.Runes[0]
				switch r {
				case 'k':
					if m.cursor > 0 {
						m.cursor--
					}
				case 'j':
					if m.cursor < len(m.items)-1 {
						m.cursor++
					}
				case 'q':
					return m, tea.Quit
				default:
					if r >= '1' && r <= '9' {
						idx := int(r - '1')
						if idx < len(m.items) {
							return m, func() tea.Msg { return menuSelectionMsg{selection: idx} }
						}
					}
				}
			}

		case tea.KeyEnter:
			return m, func() tea.Msg { return menuSelectionMsg{selection: m.cursor} }

		case tea.KeyEscape:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *menuSubModel) View() string {
	var content string

	// 제목
	content += titleStyle.Render(fmt.Sprintf("후잉 CLI v%s", Version)) + "\n\n"

	// 메인 메뉴
	content += headerStyle.Render("메인 메뉴") + "\n\n"

	// 메뉴 항목
	for i, item := range m.items {
		num := fmt.Sprintf("%d.", item.id)
		line := fmt.Sprintf(" %s %s", num, item.label)

		if i == m.cursor {
			content += selectedStyle.Render("> "+line) + "\n"
		} else {
			content += "  " + line + "\n"
		}
	}

	// 도움말
	content += "\n" + helpStyle.Render("[\u2191/\u2193/j/k] 이동  [1-5] 번호 선택  [Enter] 확인  [Esc/q] 종료")

	return content
}
