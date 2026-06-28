// cmd/section_sub.go
// 섹션 선택 - bubbletea 서브 모델

package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"whoo-cli/api"
	"whoo-cli/config"
)

// sectionSelectedMsg는 섹션 선택 완료 메시지
type sectionSelectedMsg struct {
	cfg *config.Config
}

// ─── 목록 항목 및 Delegate ────────────────────────────────────

type sectionListItem struct {
	section   api.Section
	isCurrent bool
}

func (i sectionListItem) Title() string       { return i.section.Title }
func (i sectionListItem) Description() string { return "" }
func (i sectionListItem) FilterValue() string { return i.section.Title }

type sectionDelegate struct{}

func (d sectionDelegate) Height() int                               { return 1 }
func (d sectionDelegate) Spacing() int                              { return 0 }
func (d sectionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd  { return nil }
func (d sectionDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	si := item.(sectionListItem)
	currentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))

	num := fmt.Sprintf("%d.", index+1)
	info := fmt.Sprintf(" %s (%s)", si.section.Title, si.section.Currency)

	if si.isCurrent {
		info += " [현재]"
		fmt.Fprint(w, currentStyle.Render("> "+num+info))
	} else if index == m.Index() {
		fmt.Fprint(w, selectedStyle.Render("> "+num+info))
	} else {
		fmt.Fprint(w, "  "+num+info)
	}
}

// ─── 서브 모델 ────────────────────────────────────────────────

type sectionSubModel struct {
	state       int
	sections    []api.Section // 원본 데이터 (selectSection에서 참조)
	sectionList list.Model
	cfg         *config.Config
	client      *api.WhooingClient
	errMsg      string
}

const (
	sectionStateLoading = iota
	sectionStateSelecting
	sectionStateSuccess
	sectionStateError
)

func newSectionSubModel(cfg *config.Config) *sectionSubModel {
	return &sectionSubModel{
		state:  sectionStateLoading,
		cfg:    cfg,
		client: NewClient(cfg),
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

func (m *sectionSubModel) buildList(sections []api.Section) list.Model {
	items := make([]list.Item, len(sections))
	for i, s := range sections {
		items[i] = sectionListItem{
			section:   s,
			isCurrent: s.SectionID == m.cfg.SectionID,
		}
	}
	l := newCompactListWith(items, sectionDelegate{}, 60, len(items)+2)
	// 현재 섹션으로 커서 초기화
	for i, s := range sections {
		if s.SectionID == m.cfg.SectionID {
			l.Select(i)
			break
		}
	}
	return l
}

func (m *sectionSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sectionsLoadedMsg:
		if len(msg.sections) == 0 {
			m.state = sectionStateError
			m.errMsg = "접근 가능한 섹션이 없습니다"
		} else {
			m.sections = msg.sections
			m.sectionList = m.buildList(msg.sections)
			m.state = sectionStateSelecting
		}

	case sectionErrMsg:
		m.state = sectionStateError
		m.errMsg = msg.err.Error()

	case tea.KeyMsg:
		if GlobalAction(msg) == ActionQuit {
			return m, tea.Quit
		}
		// 에러/로딩 상태: enter/esc/q 모두 메뉴 복귀
		if m.state != sectionStateSelecting {
			switch ErrorAction(msg) {
			case ActionBack, ActionExit:
				return m, func() tea.Msg { return backToMenuMsg{} }
			}
			return m, nil
		}
		// 선택 상태: 목록 키맵 적용
		switch ListAction(msg) {
		case ActionBack, ActionExit:
			return m, func() tea.Msg { return backToMenuMsg{} }
		case ActionConfirm:
			return m, m.selectSection()
		}
		if idx, ok := NumberAction(msg); ok && idx < len(m.sections) {
			m.sectionList.Select(idx)
			return m, m.selectSection()
		}
		var cmd tea.Cmd
		m.sectionList, cmd = m.sectionList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *sectionSubModel) selectSection() tea.Cmd {
	selectedSection := m.sections[m.sectionList.Index()]
	m.cfg.SectionID = selectedSection.SectionID
	if err := m.cfg.Save(); err != nil {
		m.state = sectionStateError
		m.errMsg = fmt.Sprintf("섹션 저장 실패: %v", err)
		return nil
	}
	cfg, _ := config.Load()
	return func() tea.Msg { return sectionSelectedMsg{cfg: cfg} }
}

func (m *sectionSubModel) View() string {
	var content strings.Builder
	content.WriteString(titleStyle.Render("섹션 선택") + "\n\n")

	switch m.state {
	case sectionStateLoading:
		content.WriteString("섹션 목록을 불러오는 중...\n")

	case sectionStateSelecting:
		content.WriteString(headerStyle.Render("사용 가능한 섹션:") + "\n\n")
		content.WriteString(m.sectionList.View() + "\n")
		content.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [1-9] 번호 선택  [Enter] 확인  [Esc/q] 취소") + "\n")

	case sectionStateError:
		content.WriteString(errorStyle.Render("[오류] "+m.errMsg) + "\n\n")
		content.WriteString(helpStyle.Render("[Esc/q] 취소") + "\n")
	}

	return content.String()
}
