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
	{flowTypeFlowAccount, "계정 흐름 (flow_of_account)"},
	{flowTypeFlowAccountID, "항목 흐름 (flow_of_account_id)"},
	{flowTypeChangesAccID, "항목 일일 변동 (changes_of_account_id)"},
	{flowTypeChangesClient, "거래처 일일 변동 (changes_of_client)"},
	{flowTypeChangesItem, "아이템 일일 변동 (changes_of_item)"},
}

type flowSubModel struct {
	cfg    *config.Config
	client *api.WhooingClient
	mode   flowMode
	errMsg string

	// 선택
	typeCursor int
	analysisType flowAnalysisType

	// 파라미터 입력
	paramStep  int
	paramFrom  string
	paramTo    string
	paramExtra string // account / account_id / client / item
	textInput  string

	// 결과
	resultJSON []byte
}

const (
	flowParamStepFrom = iota
	flowParamStepTo
	flowParamStepExtra
	flowParamStepConfirm
)

func newFlowSubModel(cfg *config.Config) *flowSubModel {
	return &flowSubModel{
		cfg:    cfg,
		client: NewClient(cfg),
		mode:   flowModeSelect,
	}
}

func (m *flowSubModel) Init() tea.Cmd { return nil }

func (m *flowSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case flowResultMsg:
		m.resultJSON = msg.data
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

type flowResultMsg struct{ data []byte }
type flowErrMsg struct{ err error }

func (m *flowSubModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case flowModeSelect:
		return m.handleSelectKey(msg)
	case flowModeParams:
		return m.handleParamsKey(msg)
	case flowModeResult:
		// esc 또는 q → 선택 화면으로 복귀
		switch msg.Type {
		case tea.KeyEscape:
			m.mode = flowModeSelect
		case tea.KeyRunes:
			if string(msg.Runes) == "q" {
				m.mode = flowModeSelect
			}
		}
	case flowModeError:
		// esc/q → 메뉴 복귀, enter → 선택 화면으로 복귀(재시도)
		switch ErrorAction(msg) {
		case ActionBack, ActionExit:
			return m, func() tea.Msg { return backToMenuMsg{} }
		}
		if msg.Type == tea.KeyEnter {
			m.mode = flowModeSelect
		}
	}
	return m, nil
}

func (m *flowSubModel) handleSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		return m, func() tea.Msg { return backToMenuMsg{} }
	case tea.KeyUp:
		if m.typeCursor > 0 {
			m.typeCursor--
		}
	case tea.KeyDown:
		if m.typeCursor < len(flowAnalysisLabels)-1 {
			m.typeCursor++
		}
	case tea.KeyEnter:
		m.analysisType = flowAnalysisLabels[m.typeCursor].typ
		m.mode = flowModeParams
		m.paramStep = flowParamStepFrom
		now := time.Now()
		m.paramFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("20060102")
		m.paramTo = now.Format("20060102")
		m.paramExtra = ""
		m.textInput = m.paramFrom
	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'k':
				if m.typeCursor > 0 {
					m.typeCursor--
				}
			case 'j':
				if m.typeCursor < len(flowAnalysisLabels)-1 {
					m.typeCursor++
				}
			case 'q':
				return m, func() tea.Msg { return backToMenuMsg{} }
			}
		}
	}
	return m, nil
}

func (m *flowSubModel) handleParamsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.mode = flowModeSelect
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.textInput) > 0 {
			m.textInput = m.textInput[:len(m.textInput)-1]
		}
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
		// 추가 파라미터가 필요한 타입인지 확인
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
	client := m.client
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
		return flowResultMsg{data: data}
	}
}

func (m *flowSubModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("흐름 분석") + "\n\n")

	switch m.mode {
	case flowModeSelect:
		b.WriteString(headerStyle.Render("분석 유형 선택") + "\n\n")
		for i, item := range flowAnalysisLabels {
			if i == m.typeCursor {
				b.WriteString(selectedStyle.Render("> "+item.label) + "\n")
			} else {
				b.WriteString("  " + item.label + "\n")
			}
		}
		b.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [Enter] 선택  [Esc/q] 뒤로") + "\n")

	case flowModeParams:
		b.WriteString(headerStyle.Render(flowAnalysisLabels[m.analysisType].label) + "\n\n")
		switch m.paramStep {
		case flowParamStepFrom:
			b.WriteString("시작 날짜 (YYYYMMDD): " + m.textInput + "_\n")
			b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
		case flowParamStepTo:
			b.WriteString(fmt.Sprintf("시작: %s\n", m.paramFrom))
			b.WriteString("종료 날짜 (YYYYMMDD): " + m.textInput + "_\n")
			b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
		case flowParamStepExtra:
			b.WriteString(fmt.Sprintf("기간: %s ~ %s\n", m.paramFrom, m.paramTo))
			label := m.extraParamLabel()
			b.WriteString(label + ": " + m.textInput + "_\n")
			b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
		case flowParamStepConfirm:
			b.WriteString(fmt.Sprintf("기간: %s ~ %s\n", m.paramFrom, m.paramTo))
			if m.paramExtra != "" {
				b.WriteString(fmt.Sprintf("%s: %s\n", m.extraParamLabel(), m.paramExtra))
			}
			b.WriteString("\n" + helpStyle.Render("[Enter] 조회  [Esc] 취소") + "\n")
		}

	case flowModeLoading:
		b.WriteString("데이터를 불러오는 중...\n")

	case flowModeResult:
		b.WriteString(headerStyle.Render("결과") + "\n\n")
		// raw JSON을 그대로 표시 (pretty-print)
		var buf strings.Builder
		prettyPrintJSON(m.resultJSON, &buf)
		b.WriteString(buf.String() + "\n")
		b.WriteString("\n" + helpStyle.Render("[Esc/q] 뒤로") + "\n")

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

// prettyPrintJSON은 raw JSON을 들여쓰기하여 strings.Builder에 출력
func prettyPrintJSON(data []byte, b *strings.Builder) {
	indent := 0
	inString := false
	for i := 0; i < len(data); i++ {
		c := data[i]
		switch {
		case c == '"' && (i == 0 || data[i-1] != '\\'):
			inString = !inString
			b.WriteByte(c)
		case inString:
			b.WriteByte(c)
		case c == '{' || c == '[':
			b.WriteByte(c)
			b.WriteByte('\n')
			indent++
			b.WriteString(strings.Repeat("  ", indent))
		case c == '}' || c == ']':
			b.WriteByte('\n')
			indent--
			b.WriteString(strings.Repeat("  ", indent))
			b.WriteByte(c)
		case c == ',':
			b.WriteByte(c)
			b.WriteByte('\n')
			b.WriteString(strings.Repeat("  ", indent))
		case c == ':':
			b.WriteByte(c)
			b.WriteByte(' ')
		case c == ' ' || c == '\n' || c == '\r' || c == '\t':
			// 공백 스킵 (이미 들여쓰기 처리)
		default:
			b.WriteByte(c)
		}
	}
}
