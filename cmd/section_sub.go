// cmd/section_sub.go
// 섹션 선택 - bubbletea 서브 모델

package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"whooing-cli/api"
	"whooing-cli/config"
)

// sectionSelectedMsg는 섹션 선택 완료 메시지
type sectionSelectedMsg struct {
	cfg *config.Config
}

// sectionSubModel은 섹션 선택 서브 모델
type sectionSubModel struct {
	state    int
	sections []api.Section
	cursor   int
	selected int
	cfg      *config.Config
	client   *api.WhooingClient
	errMsg   string
}

const (
	sectionStateLoading = iota
	sectionStateSelecting
	sectionStateSuccess
	sectionStateError
)

// newSectionSubModel은 새로운 섹션 선택 모델을 생성
func newSectionSubModel(cfg *config.Config) *sectionSubModel {
	return &sectionSubModel{
		state:    sectionStateLoading,
		cursor:   0,
		selected: -1,
		cfg:      cfg,
		client:   NewClient(cfg),
	}
}

func (m *sectionSubModel) Init() tea.Cmd {
	return m.fetchSections()
}

func (m *sectionSubModel) fetchSections() tea.Cmd {
	return func() tea.Msg {
		sections, err := m.client.GetSections()
		if err != nil {
			return sectionErrMsg{err: err}
		}
		return sectionsLoadedMsg{sections: sections}
	}
}

type sectionsLoadedMsg struct {
	sections []api.Section
}

type sectionErrMsg struct {
	err error
}

func (m *sectionSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sectionsLoadedMsg:
		if len(msg.sections) == 0 {
			m.state = sectionStateError
			m.errMsg = "접근 가능한 섹션이 없습니다"
		} else {
			m.sections = msg.sections
			m.state = sectionStateSelecting
			for i, s := range m.sections {
				if s.SectionID == m.cfg.SectionID {
					m.cursor = i
					break
				}
			}
		}

	case sectionErrMsg:
		m.state = sectionStateError
		m.errMsg = msg.err.Error()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEscape:
			return m, func() tea.Msg { return backToMenuMsg{} }

		case tea.KeyUp:
			if m.state == sectionStateSelecting && m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown:
			if m.state == sectionStateSelecting && m.cursor < len(m.sections)-1 {
				m.cursor++
			}

		case tea.KeyRunes:
			if m.state == sectionStateSelecting && len(msg.Runes) == 1 {
				r := msg.Runes[0]
				switch r {
				case 'k':
					if m.cursor > 0 {
						m.cursor--
					}
				case 'j':
					if m.cursor < len(m.sections)-1 {
						m.cursor++
					}
				case 'q':
					return m, func() tea.Msg { return backToMenuMsg{} }
				default:
					if r >= '1' && r <= '9' {
						idx := int(r - '1')
						if idx < len(m.sections) {
							m.cursor = idx
							return m, m.selectSection()
						}
					}
				}
			}

		case tea.KeyEnter:
			if m.state == sectionStateSelecting && len(m.sections) > 0 {
				return m, m.selectSection()
			}
		}
	}

	return m, nil
}

func (m *sectionSubModel) selectSection() tea.Cmd {
	selectedSection := m.sections[m.cursor]
	m.cfg.SectionID = selectedSection.SectionID
	if err := m.cfg.Save(); err != nil {
		m.state = sectionStateError
		m.errMsg = fmt.Sprintf("섹션 저장 실패: %v", err)
		return nil
	}
	// 설정 다시 로드
	cfg, _ := config.Load()
	return func() tea.Msg { return sectionSelectedMsg{cfg: cfg} }
}

func (m *sectionSubModel) View() string {
	currentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

	var content strings.Builder
	content.WriteString(titleStyle.Render("섹션 선택") + "\n\n")

	switch m.state {
	case sectionStateLoading:
		content.WriteString("섹션 목록을 불러오는 중...\n")

	case sectionStateSelecting:
		content.WriteString(headerStyle.Render("사용 가능한 섹션:") + "\n\n")

		for i, section := range m.sections {
			num := fmt.Sprintf("%d.", i+1)
			sectionInfo := fmt.Sprintf(" %s (%s)", section.Title, section.Currency)

			var line string
			if section.SectionID == m.cfg.SectionID {
				sectionInfo += " [현재]"
				line = currentStyle.Render("> " + num + sectionInfo)
			} else if i == m.cursor {
				line = selectedStyle.Render("> " + num + sectionInfo)
			} else {
				line = "  " + num + sectionInfo
			}

			content.WriteString(line + "\n")
		}

		content.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [1-9] 번호 선택  [Enter] 확인  [Esc/q] 취소") + "\n")

	case sectionStateError:
		content.WriteString(errorStyle.Render("[오류] "+m.errMsg) + "\n\n")
		content.WriteString(helpStyle.Render("[Esc/q] 취소") + "\n")
	}

	return content.String()
}
