// cmd/section_hub_sub.go
// 섹션 메뉴 허브 — 변경 / 관리 1depth 선택

package cmd

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"whoo-cli/config"
)

// sectionHubSubModel은 메인 메뉴 "섹션" 진입 후 하위 선택
type sectionHubSubModel struct {
	list list.Model
	cfg  *config.Config
}

func newSectionHubSubModel(cfg *config.Config, selected int) *sectionHubSubModel {
	items := []list.Item{
		simpleItem{"섹션 변경"},
		simpleItem{"섹션 관리"},
	}
	m := &sectionHubSubModel{
		list: newCompactList(items, 40, len(items)+2),
		cfg:  cfg,
	}
	selectListIndex(&m.list, selected)
	return m
}

func (m *sectionHubSubModel) cursorIndex() int {
	return m.list.Index()
}

func (m *sectionHubSubModel) Init() tea.Cmd {
	return nil
}

func (m *sectionHubSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if GlobalAction(msg) == ActionQuit {
			return m, tea.Quit
		}
		switch ListAction(msg) {
		case ActionBack:
			return m, func() tea.Msg { return backToMenuMsg{} }
		case ActionConfirm:
			return m, m.selectCurrent()
		}
		if idx, confirm, ok := handleListNumberJump(msg, len(m.list.Items()), true); ok {
			m.list.Select(idx)
			if confirm {
				return m, m.selectCurrent()
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *sectionHubSubModel) selectCurrent() tea.Cmd {
	idx := m.list.Index()
	return func() tea.Msg {
		switch idx {
		case 0: // 변경
			return stateTransitionMsg{newState: stateSection}
		case 1: // 관리
			return stateTransitionMsg{newState: stateSectionManage}
		default:
			return backToMenuMsg{}
		}
	}
}

func (m *sectionHubSubModel) View() string {
	var content string
	content += titleStyle.Render("섹션") + "\n\n"
	content += headerStyle.Render("작업 선택") + "\n\n"
	content += m.list.View() + "\n"
	content += "\n" + helpStyle.Render(numberedMenuHelp(len(m.list.Items())))
	return content
}
