// cmd/section_manage_sub.go
// 섹션 관리 - bubbletea 서브 모델 (생성/수정/삭제/정렬)

package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"whoo-cli/api"
	"whoo-cli/config"
)

// sectionManageMode는 섹션 관리 서브 모드
type sectionManageMode int

const (
	sectionManageModeList sectionManageMode = iota
	sectionManageModeAdd
	sectionManageModeEdit
	sectionManageModeConfirmDelete
	sectionManageModeLoading
	sectionManageModeError
)

// ─── 목록 항목 및 Delegate ────────────────────────────────────

type sectionManageListItem struct {
	section   api.Section
	isCurrent bool
}

func (i sectionManageListItem) Title() string       { return i.section.Title }
func (i sectionManageListItem) Description() string { return "" }
func (i sectionManageListItem) FilterValue() string { return i.section.Title }

type sectionManageDelegate struct{}

func (d sectionManageDelegate) Height() int                               { return 1 }
func (d sectionManageDelegate) Spacing() int                              { return 0 }
func (d sectionManageDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd  { return nil }
func (d sectionManageDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	si := item.(sectionManageListItem)
	mark := "  "
	if si.isCurrent {
		mark = "* "
	}
	line := fmt.Sprintf("%s%s (%s)", mark, si.section.Title, si.section.Currency)
	if index == m.Index() {
		fmt.Fprint(w, selectedStyle.Render("> "+line))
	} else {
		fmt.Fprint(w, "  "+line)
	}
}

// ─── 서브 모델 ────────────────────────────────────────────────

type sectionManageSubModel struct {
	cfg         *config.Config
	client      *api.WhooingClient
	sections    []api.Section // 원본 데이터
	sectionList list.Model
	mode        sectionManageMode
	errMsg      string
	infoMsg     string

	// 폼 입력 필드 (add/edit)
	formStep     int
	formTitle    string
	formCurrency string
	formMemo     string
	textInput    string
	editingID    string
}

const (
	sectionFormStepTitle = iota
	sectionFormStepCurrency
	sectionFormStepMemo
	sectionFormStepConfirm
)

func newSectionManageSubModel(cfg *config.Config) *sectionManageSubModel {
	return &sectionManageSubModel{
		cfg:    cfg,
		client: NewClient(cfg),
		mode:   sectionManageModeLoading,
	}
}

func (m *sectionManageSubModel) Init() tea.Cmd {
	return m.loadSections()
}

func (m *sectionManageSubModel) loadSections() tea.Cmd {
	return func() tea.Msg {
		sections, err := m.client.GetSections()
		if err != nil {
			return sectionManageErrMsg{err: err}
		}
		return sectionManageLoadedMsg{sections: sections}
	}
}

type sectionManageLoadedMsg struct{ sections []api.Section }
type sectionManageErrMsg struct{ err error }
type sectionManageDoneMsg struct{}

func (m *sectionManageSubModel) buildSectionList(sections []api.Section) list.Model {
	items := make([]list.Item, len(sections))
	for i, s := range sections {
		items[i] = sectionManageListItem{
			section:   s,
			isCurrent: s.SectionID == m.cfg.SectionID,
		}
	}
	return newCompactListWith(items, sectionManageDelegate{}, 60, len(items)+2)
}

func (m *sectionManageSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sectionManageLoadedMsg:
		m.sections = msg.sections
		m.sectionList = m.buildSectionList(msg.sections)
		m.mode = sectionManageModeList
		m.infoMsg = ""

	case sectionManageErrMsg:
		m.mode = sectionManageModeError
		m.errMsg = msg.err.Error()

	case tea.KeyMsg:
		if GlobalAction(msg) == ActionQuit {
			return m, tea.Quit
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *sectionManageSubModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case sectionManageModeList:
		return m.handleListKey(msg)
	case sectionManageModeAdd, sectionManageModeEdit:
		return m.handleFormKey(msg)
	case sectionManageModeConfirmDelete:
		return m.handleDeleteKey(msg)
	case sectionManageModeError:
		// enter는 재시도, esc/q는 메뉴 복귀 (이 화면의 특수 에러 정책)
		if msg.Type == tea.KeyEnter {
			m.mode = sectionManageModeLoading
			return m, m.loadSections()
		}
		switch ErrorAction(msg) {
		case ActionBack, ActionExit:
			return m, func() tea.Msg { return backToMenuMsg{} }
		}
	}
	return m, nil
}

func (m *sectionManageSubModel) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch ListAction(msg) {
	case ActionBack, ActionExit:
		return m, func() tea.Msg { return backToMenuMsg{} }
	case ActionEdit:
		if len(m.sections) > 0 {
			s := m.sections[m.sectionList.Index()]
			m.mode = sectionManageModeEdit
			m.editingID = s.SectionID
			m.formStep = sectionFormStepTitle
			m.formTitle = s.Title
			m.formCurrency = s.Currency
			m.formMemo = ""
			m.textInput = s.Title
		}
		return m, nil
	case ActionDelete:
		if len(m.sections) > 0 {
			m.mode = sectionManageModeConfirmDelete
		}
		return m, nil
	}
	// 도메인 전용: a = 추가
	if msg.String() == "a" {
		m.mode = sectionManageModeAdd
		m.formStep = sectionFormStepTitle
		m.formTitle = ""
		m.formCurrency = "KRW"
		m.formMemo = ""
		m.textInput = ""
		return m, nil
	}
	var cmd tea.Cmd
	m.sectionList, cmd = m.sectionList.Update(msg)
	return m, cmd
}

func (m *sectionManageSubModel) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEscape:
		m.mode = sectionManageModeList
		m.textInput = ""
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.textInput) > 0 {
			m.textInput = m.textInput[:len(m.textInput)-1]
		}
	case tea.KeyEnter:
		return m.advanceFormStep()
	case tea.KeyRunes:
		m.textInput += string(msg.Runes)
	}
	return m, nil
}

func (m *sectionManageSubModel) advanceFormStep() (tea.Model, tea.Cmd) {
	switch m.formStep {
	case sectionFormStepTitle:
		if strings.TrimSpace(m.textInput) == "" {
			return m, nil
		}
		m.formTitle = strings.TrimSpace(m.textInput)
		m.formStep = sectionFormStepCurrency
		if m.mode == sectionManageModeEdit {
			m.textInput = m.formCurrency
		} else {
			m.textInput = "KRW"
		}
	case sectionFormStepCurrency:
		if strings.TrimSpace(m.textInput) == "" {
			m.textInput = "KRW"
		}
		m.formCurrency = strings.ToUpper(strings.TrimSpace(m.textInput))
		m.formStep = sectionFormStepMemo
		m.textInput = m.formMemo
	case sectionFormStepMemo:
		m.formMemo = strings.TrimSpace(m.textInput)
		m.formStep = sectionFormStepConfirm
		m.textInput = ""
	case sectionFormStepConfirm:
		return m, m.submitForm()
	}
	return m, nil
}

func (m *sectionManageSubModel) submitForm() tea.Cmd {
	if m.mode == sectionManageModeAdd {
		return func() tea.Msg {
			_, err := m.client.CreateSection(api.SectionCreateParams{
				Title:    m.formTitle,
				Currency: m.formCurrency,
				Memo:     m.formMemo,
			})
			if err != nil {
				return sectionManageErrMsg{err: err}
			}
			return sectionManageLoadedMsg{}
		}
	}
	editID := m.editingID
	return func() tea.Msg {
		_, err := m.client.UpdateSection(editID, api.SectionUpdateParams{
			Title:    m.formTitle,
			Currency: m.formCurrency,
			Memo:     m.formMemo,
		})
		if err != nil {
			return sectionManageErrMsg{err: err}
		}
		sections, err := m.client.GetSections()
		if err != nil {
			return sectionManageErrMsg{err: err}
		}
		return sectionManageLoadedMsg{sections: sections}
	}
}

func (m *sectionManageSubModel) handleDeleteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch ConfirmAction(msg) {
	case ActionConfirm:
		return m, m.deleteSection()
	case ActionBack:
		m.mode = sectionManageModeList
	}
	return m, nil
}

func (m *sectionManageSubModel) deleteSection() tea.Cmd {
	if len(m.sections) == 0 {
		return nil
	}
	id := m.sections[m.sectionList.Index()].SectionID
	return func() tea.Msg {
		_, err := m.client.DeleteSections([]string{id})
		if err != nil {
			return sectionManageErrMsg{err: err}
		}
		sections, err := m.client.GetSections()
		if err != nil {
			return sectionManageErrMsg{err: err}
		}
		return sectionManageLoadedMsg{sections: sections}
	}
}

func (m *sectionManageSubModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("섹션 관리") + "\n\n")

	switch m.mode {
	case sectionManageModeLoading:
		b.WriteString("섹션 목록을 불러오는 중...\n")

	case sectionManageModeList:
		if m.infoMsg != "" {
			b.WriteString(successStyle.Render(m.infoMsg) + "\n\n")
		}
		b.WriteString(headerStyle.Render("섹션 목록") + "\n\n")
		b.WriteString(m.sectionList.View() + "\n")
		b.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [a] 추가  [e] 수정  [d] 삭제  [Esc/q] 뒤로") + "\n")

	case sectionManageModeAdd:
		b.WriteString(headerStyle.Render("섹션 추가") + "\n\n")
		b.WriteString(m.renderFormStep("추가"))

	case sectionManageModeEdit:
		b.WriteString(headerStyle.Render("섹션 수정") + "\n\n")
		b.WriteString(m.renderFormStep("수정"))

	case sectionManageModeConfirmDelete:
		if len(m.sections) > 0 {
			s := m.sections[m.sectionList.Index()]
			b.WriteString(errorStyle.Render(fmt.Sprintf("'%s' 섹션을 삭제하시겠습니까?", s.Title)) + "\n\n")
			b.WriteString(helpStyle.Render("[y] 삭제  [n/Esc] 취소") + "\n")
		}

	case sectionManageModeError:
		b.WriteString(errorStyle.Render("[오류] "+m.errMsg) + "\n\n")
		b.WriteString(helpStyle.Render("[Enter] 재시도  [Esc] 뒤로") + "\n")
	}

	return b.String()
}

func (m *sectionManageSubModel) renderFormStep(action string) string {
	var b strings.Builder

	switch m.formStep {
	case sectionFormStepTitle:
		b.WriteString("제목: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case sectionFormStepCurrency:
		b.WriteString(fmt.Sprintf("제목: %s\n", m.formTitle))
		b.WriteString("통화: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case sectionFormStepMemo:
		b.WriteString(fmt.Sprintf("제목: %s  통화: %s\n", m.formTitle, m.formCurrency))
		b.WriteString("설명: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case sectionFormStepConfirm:
		b.WriteString(fmt.Sprintf("제목: %s\n통화: %s\n설명: %s\n\n", m.formTitle, m.formCurrency, m.formMemo))
		b.WriteString(helpStyle.Render(fmt.Sprintf("[Enter] %s 실행  [Esc] 취소", action)) + "\n")
	}

	return b.String()
}
