// cmd/flow_sub.go
// 흐름 분석 - bubbletea 서브 모델 (flow/changes 결과 표시)

package cmd

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"whoo-cli/api"
	"whoo-cli/config"
)

type flowMode int

const (
	flowModeSelect flowMode = iota // 분석 유형 선택
	flowModeParams                 // 파라미터 입력
	flowModeResult                 // 결과 표시
	flowModeLoading
	flowModeError
)

type flowAnalysisType int

const (
	flowTypeFlowAccount   flowAnalysisType = iota // flow_of_account
	flowTypeFlowAccountID                         // flow_of_account_id
	flowTypeChangesAccID                          // changes_of_account_id
	flowTypeChangesClient                         // changes_of_client
	flowTypeChangesItem                           // changes_of_item
)

var flowAnalysisLabels = []struct {
	typ   flowAnalysisType
	label string
}{
	{flowTypeFlowAccount, "계정 흐름"},
	{flowTypeFlowAccountID, "항목 흐름"},
	{flowTypeChangesAccID, "항목 일일 변동"},
	{flowTypeChangesClient, "거래처 일일 변동"},
	{flowTypeChangesItem, "아이템 일일 변동"},
}

const flowResultViewport = 18

type flowSubModel struct {
	cfg    *config.Config
	client *api.WhooingClient
	mode   flowMode
	errMsg string

	// 선택
	typeCursor   int
	analysisType flowAnalysisType

	// 파라미터 입력
	paramStep  int
	paramFrom  string
	paramTo    string
	paramExtra string // account / account_id / client / item
	textInput  string

	// 결과 (렌더된 줄 + 스크롤)
	resultLines  []string
	resultOffset int
	resultTitle  string
}

const (
	flowParamStepFrom = iota
	flowParamStepTo
	flowParamStepExtra
	flowParamStepConfirm
)

func newFlowSubModel(cfg *config.Config, typeIndex int) *flowSubModel {
	return &flowSubModel{
		cfg:        cfg,
		client:     NewClient(cfg),
		mode:       flowModeSelect,
		typeCursor: clampIndex(typeIndex, len(flowAnalysisLabels)),
	}
}

func (m *flowSubModel) typeIndex() int {
	return m.typeCursor
}

func (m *flowSubModel) Init() tea.Cmd { return nil }

func (m *flowSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case flowResultMsg:
		m.resultTitle = msg.title
		m.resultLines = msg.lines
		m.resultOffset = 0
		m.mode = flowModeResult
	case flowErrMsg:
		m.errMsg = msg.err.Error()
		m.mode = flowModeError
	case tea.KeyMsg:
		if GlobalAction(msg) == ActionQuit {
			return m, tea.Quit
		}
		return m.handleKey(msg)
	}
	return m, nil
}

type flowResultMsg struct {
	title string
	lines []string
}
type flowErrMsg struct{ err error }

func (m *flowSubModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case flowModeSelect:
		return m.handleSelectKey(msg)
	case flowModeParams:
		return m.handleParamsKey(msg)
	case flowModeResult:
		return m.handleResultKey(msg)
	case flowModeError:
		switch ErrorAction(msg) {
		case ActionBack:
			return m, func() tea.Msg { return backToMenuMsg{} }
		}
		if msg.Type == tea.KeyEnter {
			m.mode = flowModeSelect
		}
	}
	return m, nil
}

func (m *flowSubModel) handleResultKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = flowModeSelect
		m.resultLines = nil
		m.resultOffset = 0
		return m, nil
	case "up", "k":
		if m.resultOffset > 0 {
			m.resultOffset--
		}
	case "down", "j":
		maxOff := len(m.resultLines) - flowResultViewport
		if maxOff < 0 {
			maxOff = 0
		}
		if m.resultOffset < maxOff {
			m.resultOffset++
		}
	case "pgup", "ctrl+u":
		m.resultOffset -= flowResultViewport
		if m.resultOffset < 0 {
			m.resultOffset = 0
		}
	case "pgdown", "ctrl+d", " ":
		maxOff := len(m.resultLines) - flowResultViewport
		if maxOff < 0 {
			maxOff = 0
		}
		m.resultOffset += flowResultViewport
		if m.resultOffset > maxOff {
			m.resultOffset = maxOff
		}
	case "g", "home":
		m.resultOffset = 0
	case "G", "end":
		maxOff := len(m.resultLines) - flowResultViewport
		if maxOff < 0 {
			maxOff = 0
		}
		m.resultOffset = maxOff
	}
	return m, nil
}

func (m *flowSubModel) handleSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	r := handleVerticalMenuKey(msg, m.typeCursor, len(flowAnalysisLabels))
	m.typeCursor = r.Cursor
	if r.Back {
		return m, func() tea.Msg { return backToMenuMsg{} }
	}
	if r.Confirm {
		return m, m.enterAnalysisType(m.typeCursor)
	}
	return m, nil
}

func (m *flowSubModel) enterAnalysisType(idx int) tea.Cmd {
	if idx < 0 || idx >= len(flowAnalysisLabels) {
		return nil
	}
	m.typeCursor = idx
	m.analysisType = flowAnalysisLabels[idx].typ
	m.mode = flowModeParams
	m.paramStep = flowParamStepFrom
	now := time.Now()
	m.paramFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("20060102")
	m.paramTo = now.Format("20060102")
	m.paramExtra = ""
	m.textInput = m.paramFrom
	return nil
}

func (m *flowSubModel) handleParamsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.mode = flowModeSelect
	case tea.KeyBackspace, tea.KeyDelete:
		m.textInput = backspaceRunes(m.textInput)
	case tea.KeyEnter:
		return m.advanceParamStep()
	case tea.KeyRunes:
		m.textInput += string(msg.Runes)
	}
	return m, nil
}

func (m *flowSubModel) advanceParamStep() (tea.Model, tea.Cmd) {
	switch m.paramStep {
	case flowParamStepFrom:
		if strings.TrimSpace(m.textInput) != "" {
			m.paramFrom = strings.TrimSpace(m.textInput)
		}
		m.paramStep = flowParamStepTo
		m.textInput = m.paramTo
	case flowParamStepTo:
		if strings.TrimSpace(m.textInput) != "" {
			m.paramTo = strings.TrimSpace(m.textInput)
		}
		if m.needsExtraParam() {
			m.paramStep = flowParamStepExtra
			m.textInput = ""
		} else {
			m.paramStep = flowParamStepConfirm
			m.textInput = ""
		}
	case flowParamStepExtra:
		m.paramExtra = strings.TrimSpace(m.textInput)
		m.paramStep = flowParamStepConfirm
		m.textInput = ""
	case flowParamStepConfirm:
		return m, m.fetchResult()
	}
	return m, nil
}

func (m *flowSubModel) needsExtraParam() bool {
	switch m.analysisType {
	case flowTypeFlowAccount,
		flowTypeFlowAccountID, flowTypeChangesAccID,
		flowTypeChangesClient, flowTypeChangesItem:
		return true
	}
	return false
}

func (m *flowSubModel) fetchResult() tea.Cmd {
	fromInt := 0
	toInt := 0
	fmt.Sscanf(m.paramFrom, "%d", &fromInt)
	fmt.Sscanf(m.paramTo, "%d", &toInt)

	q := api.FlowQuery{
		SectionID: m.cfg.SectionID,
		StartDate: fromInt,
		EndDate:   toInt,
	}
	switch m.analysisType {
	case flowTypeFlowAccount:
		q.Account = m.paramExtra
	case flowTypeFlowAccountID, flowTypeChangesAccID:
		q.AccountID = m.paramExtra
	case flowTypeChangesClient:
		q.Item = m.paramExtra
	case flowTypeChangesItem:
		q.Item = m.paramExtra
	}

	analysisType := m.analysisType
	label := flowAnalysisLabels[m.typeCursor].label
	client := m.client
	sectionID := m.cfg.SectionID
	paramFrom, paramTo, paramExtra := m.paramFrom, m.paramTo, m.paramExtra
	m.mode = flowModeLoading

	return func() tea.Msg {
		var data []byte
		var err error
		switch analysisType {
		case flowTypeFlowAccount:
			data, err = client.FlowOfAccount(q)
		case flowTypeFlowAccountID:
			data, err = client.FlowOfAccountID(q)
		case flowTypeChangesAccID:
			data, err = client.ChangesOfAccountID(q)
		case flowTypeChangesClient:
			data, err = client.ChangesOfClient(q)
		case flowTypeChangesItem:
			data, err = client.ChangesOfItem(q)
		}
		if err != nil {
			return flowErrMsg{err: err}
		}

		// 계정 이름은 실패해도 결과는 표시
		var am *api.AccountsMap
		if a, aerr := client.GetAccountsMap(sectionID); aerr == nil {
			am = a
		}

		title := fmt.Sprintf("%s  %s ~ %s", label, FormatDate(paramFrom), FormatDate(paramTo))
		if paramExtra != "" {
			title += "  " + paramExtra
		}

		var lines []string
		switch analysisType {
		case flowTypeFlowAccount, flowTypeFlowAccountID:
			groups, perr := parseFlowGroups(data)
			if perr != nil {
				return flowErrMsg{err: perr}
			}
			lines = renderFlowGroupsLines(groups, am)
		default:
			ch, perr := parseChangesView(data)
			if perr != nil {
				return flowErrMsg{err: perr}
			}
			lines = renderChangesLines(ch)
		}
		return flowResultMsg{title: title, lines: lines}
	}
}

func (m *flowSubModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("흐름 분석") + "\n\n")

	switch m.mode {
	case flowModeSelect:
		b.WriteString(headerStyle.Render("분석 유형 선택") + "\n\n")
		labels := make([]string, len(flowAnalysisLabels))
		for i, item := range flowAnalysisLabels {
			labels[i] = item.label
		}
		b.WriteString(renderNumberedMenuLines(labels, m.typeCursor))
		b.WriteString("\n" + helpStyle.Render(numberedMenuHelp(len(flowAnalysisLabels))) + "\n")

	case flowModeParams:
		b.WriteString(headerStyle.Render(flowAnalysisLabels[m.typeCursor].label) + "\n\n")
		switch m.paramStep {
		case flowParamStepFrom:
			b.WriteString("시작 날짜 (YYYYMMDD): " + m.textInput + "_\n")
			b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
		case flowParamStepTo:
			b.WriteString(fmt.Sprintf("시작: %s\n", FormatDate(m.paramFrom)))
			b.WriteString("종료 날짜 (YYYYMMDD): " + m.textInput + "_\n")
			b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
		case flowParamStepExtra:
			b.WriteString(fmt.Sprintf("기간: %s ~ %s\n", FormatDate(m.paramFrom), FormatDate(m.paramTo)))
			b.WriteString(m.extraParamLabel() + ": " + m.textInput + "_\n")
			b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
		case flowParamStepConfirm:
			b.WriteString(fmt.Sprintf("기간: %s ~ %s\n", FormatDate(m.paramFrom), FormatDate(m.paramTo)))
			if m.paramExtra != "" {
				b.WriteString(fmt.Sprintf("%s: %s\n", m.extraParamLabel(), m.paramExtra))
			}
			b.WriteString("\n" + helpStyle.Render("[Enter] 조회  [Esc] 취소") + "\n")
		}

	case flowModeLoading:
		b.WriteString(loadingStyle.Render("데이터를 불러오는 중...") + "\n")

	case flowModeResult:
		b.WriteString(headerStyle.Render(m.resultTitle) + "\n\n")
		view, off := sliceViewport(m.resultLines, m.resultOffset, flowResultViewport)
		for _, line := range view {
			b.WriteString(line + "\n")
		}
		// 스크롤 위치 표시
		if len(m.resultLines) > flowResultViewport {
			b.WriteString("\n")
			b.WriteString(helpStyle.Render(fmt.Sprintf("  (%d–%d / %d줄)",
				off+1, off+len(view), len(m.resultLines))))
			b.WriteString("\n")
		}
		b.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 스크롤  [PgUp/PgDn] 페이지  [Esc] 뒤로") + "\n")

	case flowModeError:
		b.WriteString(errorStyle.Render("[오류] "+m.errMsg) + "\n\n")
		b.WriteString(helpStyle.Render("[Enter] 처음으로  [Esc] 메뉴로") + "\n")
	}

	return b.String()
}

func (m *flowSubModel) extraParamLabel() string {
	switch m.analysisType {
	case flowTypeFlowAccount:
		return "계정 (assets|liabilities|capital|expenses|income)"
	case flowTypeFlowAccountID, flowTypeChangesAccID:
		return "항목 ID"
	case flowTypeChangesClient:
		return "거래처명"
	case flowTypeChangesItem:
		return "아이템명"
	}
	return "값"
}
