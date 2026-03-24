// cmd/transactions_sub.go
// 거래내역 조회 - bubbletea 서브 모델

package cmd

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"whooing-cli/api"
	"whooing-cli/config"
)

const pageSize = 20

// 범위 선택 프리셋 목록
var rangePresets = []string{"이번 달", "지난 달", "최근 3개월", "직접 입력"}

// calcDateRange는 프리셋 인덱스에 따라 시작일/종료일 반환
func calcDateRange(preset int) (string, string) {
	now := time.Now()
	switch preset {
	case 0: // 이번 달
		start := fmt.Sprintf("%04d%02d01", now.Year(), now.Month())
		end := fmt.Sprintf("%04d%02d%02d", now.Year(), now.Month(), now.Day())
		return start, end
	case 1: // 지난 달
		lastMonth := now.AddDate(0, -1, 0)
		start := fmt.Sprintf("%04d%02d01", lastMonth.Year(), lastMonth.Month())
		lastDay := time.Date(lastMonth.Year(), lastMonth.Month()+1, 0, 0, 0, 0, 0, time.Local)
		end := fmt.Sprintf("%04d%02d%02d", lastDay.Year(), lastDay.Month(), lastDay.Day())
		return start, end
	case 2: // 최근 3개월
		threeMonthsAgo := now.AddDate(0, -3, 0)
		start := fmt.Sprintf("%04d%02d01", threeMonthsAgo.Year(), threeMonthsAgo.Month())
		end := fmt.Sprintf("%04d%02d%02d", now.Year(), now.Month(), now.Day())
		return start, end
	default:
		return "", ""
	}
}

// validateDateInput은 시작일/종료일 문자열의 유효성을 검사
func validateDateInput(startStr, endStr string) error {
	if len(startStr) != 8 || len(endStr) != 8 {
		return fmt.Errorf("8자리 날짜를 입력하세요 (YYYYMMDD)")
	}
	start, err := time.Parse("20060102", startStr)
	if err != nil {
		return fmt.Errorf("잘못된 시작일 형식입니다")
	}
	end, err := time.Parse("20060102", endStr)
	if err != nil {
		return fmt.Errorf("잘못된 종료일 형식입니다")
	}
	if end.Before(start) {
		return fmt.Errorf("종료일은 시작일 이후여야 합니다")
	}
	if end.Sub(start) > 365*24*time.Hour {
		return fmt.Errorf("조회 범위는 최대 1년입니다")
	}
	return nil
}

// transactionsSubModel은 거래내역 조회 서브 모델
type transactionsSubModel struct {
	cfg           *config.Config
	client        *api.WhooingClient
	entries       []api.Entry
	accountsMap   *api.AccountsMap
	startDate     string
	endDate       string
	loading       bool
	err           error
	width         int
	height        int
	table         table.Model
	tableReady    bool
	confirmDelete bool
	deleting      bool
	deleteErr     error
	hasMore       bool // 추가 데이터 존재 여부
	loadingMore   bool // 더보기 로딩 중
	rangeMode     bool // 범위 선택 모드
	rangeCursor   int  // 범위 선택 커서 (0~3)
	customInput   bool            // 직접 입력 모드
	inputField    int             // 0: 시작일, 1: 종료일
	startInput    textinput.Model // 시작일 입력
	endInput      textinput.Model // 종료일 입력
	inputErr      string          // 입력 유효성 에러
}

// newTransactionsSubModel은 새로운 거래내역 모델을 생성
func newTransactionsSubModel(cfg *config.Config) *transactionsSubModel {
	now := time.Now()
	startDate := fmt.Sprintf("%04d%02d01", now.Year(), now.Month())
	endDate := fmt.Sprintf("%04d%02d%02d", now.Year(), now.Month(), now.Day())

	// 테이블 컬럼 정의
	columns := []table.Column{
		{Title: "날짜", Width: 12},
		{Title: "왼쪽", Width: 14},
		{Title: "오른쪽", Width: 14},
		{Title: "금액", Width: 12},
		{Title: "아이템", Width: 16},
		{Title: "메모", Width: 14},
	}

	// 커스텀 KeyMap: 'd'를 삭제용으로 사용하기 위해 HalfPageDown에서 제거
	keys := table.KeyMap{
		LineUp:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "위로")),
		LineDown:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "아래로")),
		PageUp:       key.NewBinding(key.WithKeys("pgup", "b"), key.WithHelp("pgup/b", "페이지 업")),
		PageDown:     key.NewBinding(key.WithKeys("pgdown", " " /*space*/, "f"), key.WithHelp("pgdn/space/f", "페이지 다운")),
		HalfPageUp:   key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "½ 페이지 업")),
		HalfPageDown: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "½ 페이지 다운")),
		GotoTop:      key.NewBinding(key.WithKeys("home", "g"), key.WithHelp("home/g", "처음으로")),
		GotoBottom:   key.NewBinding(key.WithKeys("end", "G"), key.WithHelp("end/G", "끝으로")),
	}

	// 테이블 스타일 설정
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("6")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("2")).
		Bold(true)

	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
		table.WithKeyMap(keys),
	)

	tbl.SetStyles(s)

	si := textinput.New()
	si.Placeholder = "YYYYMMDD"
	si.CharLimit = 8

	ei := textinput.New()
	ei.Placeholder = "YYYYMMDD"
	ei.CharLimit = 8

	return &transactionsSubModel{
		cfg:        cfg,
		client:     NewClient(cfg),
		startDate:  startDate,
		endDate:    endDate,
		loading:    true,
		table:      tbl,
		tableReady: false,
		startInput: si,
		endInput:   ei,
	}
}

func (m *transactionsSubModel) Init() tea.Cmd {
	return m.fetchEntries()
}

func (m *transactionsSubModel) fetchEntries() tea.Cmd {
	return func() tea.Msg {
		// 거래 내역 조회
		entries, err := m.client.GetEntries(m.cfg.SectionID, m.startDate, m.endDate, pageSize, "")
		if err != nil {
			return transactionsErrMsg{err: err}
		}
		// 계정 정보 조회
		accountsMap, err := m.client.GetAccountsMap(m.cfg.SectionID)
		if err != nil {
			return transactionsErrMsg{err: err}
		}
		return transactionsLoadedMsg{
			entries:     entries,
			accountsMap: accountsMap,
			hasMore:     len(entries) >= pageSize,
		}
	}
}

type transactionsLoadedMsg struct {
	entries     []api.Entry
	accountsMap *api.AccountsMap
	hasMore     bool
}

type transactionsMoreLoadedMsg struct {
	entries []api.Entry
}

type transactionsErrMsg struct {
	err error
}

type transactionDeletedMsg struct{}

type transactionDeleteErrMsg struct {
	err error
}

func (m *transactionsSubModel) fetchMoreEntries() tea.Cmd {
	if len(m.entries) == 0 {
		return nil
	}
	cursor := m.entries[len(m.entries)-1].EntryDate
	return func() tea.Msg {
		entries, err := m.client.GetEntries(m.cfg.SectionID, m.startDate, m.endDate, pageSize, cursor)
		if err != nil {
			return transactionsErrMsg{err: err}
		}
		return transactionsMoreLoadedMsg{entries: entries}
	}
}

func (m *transactionsSubModel) buildTableRows() []table.Row {
	rows := make([]table.Row, len(m.entries))
	for i, e := range m.entries {
		lTitle := ""
		rTitle := ""
		if m.accountsMap != nil {
			lTitle = m.accountsMap.GetTitle(e.LAccount, e.LAccountID)
			rTitle = m.accountsMap.GetTitle(e.RAccount, e.RAccountID)
		}
		rows[i] = table.Row{
			FormatDate(e.DateOnly()),
			lTitle,
			rTitle,
			FormatMoney(e.Money),
			e.Item,
			e.Memo,
		}
	}
	return rows
}

func (m *transactionsSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// 삭제 확인 상태일 때
	if m.confirmDelete {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "y", "Y", "enter":
				// 삭제 실행 (Y가 기본값)
				m.confirmDelete = false
				m.deleting = true
				return m, m.deleteEntry()
			case "n", "N", "esc":
				// 취소
				m.confirmDelete = false
				return m, nil
			}
		}
		return m, nil
	}

	// 삭제 중 상태
	if m.deleting {
		switch msg.(type) {
		case transactionDeletedMsg:
			m.deleting = false
			m.deleteErr = nil
			// 데이터 재조회
			return m, m.fetchEntries()
		case transactionDeleteErrMsg:
			m.deleting = false
			m.deleteErr = msg.(transactionDeleteErrMsg).err
			return m, nil
		}
	}

	// 삭제 에러 표시 중
	if m.deleteErr != nil {
		switch msg.(type) {
		case tea.KeyMsg:
			// 아무 키나 누르면 에러 클리어
			m.deleteErr = nil
			return m, nil
		}
	}

	// 직접 입력 모드
	if m.rangeMode && m.customInput {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.customInput = false
				m.inputErr = ""
				return m, nil
			case "enter":
				if m.inputField == 0 {
					// 시작일 유효성 검사
					startStr := m.startInput.Value()
					if len(startStr) != 8 {
						m.inputErr = "8자리 날짜를 입력하세요 (YYYYMMDD)"
						return m, nil
					}
					if _, err := time.Parse("20060102", startStr); err != nil {
						m.inputErr = "잘못된 시작일 형식입니다"
						return m, nil
					}
					m.inputField = 1
					m.startInput.Blur()
					m.endInput.Focus()
					m.inputErr = ""
					return m, textinput.Blink
				}
				// 종료일 입력 완료 → 유효성 검사
				startStr := m.startInput.Value()
				endStr := m.endInput.Value()
				if err := validateDateInput(startStr, endStr); err != nil {
					m.inputErr = err.Error()
					return m, nil
				}
				m.startDate = startStr
				m.endDate = endStr
				m.rangeMode = false
				m.customInput = false
				m.inputField = 0
				m.inputErr = ""
				m.loading = true
				m.entries = nil
				m.tableReady = false
				m.hasMore = false
				return m, m.fetchEntries()
			}
		}
		// textinput에 메시지 전달
		var cmd tea.Cmd
		if m.inputField == 0 {
			m.startInput, cmd = m.startInput.Update(msg)
		} else {
			m.endInput, cmd = m.endInput.Update(msg)
		}
		return m, cmd
	}

	// 범위 선택 모드
	if m.rangeMode && !m.customInput {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.rangeCursor > 0 {
					m.rangeCursor--
				}
				return m, nil
			case "down", "j":
				if m.rangeCursor < len(rangePresets)-1 {
					m.rangeCursor++
				}
				return m, nil
			case "enter":
				if m.rangeCursor == 3 {
					// 직접 입력 모드로 전환
					m.customInput = true
					m.inputField = 0
					m.startInput.Reset()
					m.endInput.Reset()
					m.startInput.Focus()
					return m, textinput.Blink
				}
				// 프리셋 선택
				start, end := calcDateRange(m.rangeCursor)
				m.startDate = start
				m.endDate = end
				m.rangeMode = false
				m.loading = true
				m.entries = nil
				m.tableReady = false
				m.hasMore = false
				return m, m.fetchEntries()
			case "esc", "q":
				m.rangeMode = false
				return m, nil
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// 테이블 높이 조정 (헤더 3 + 테이블 + 합계 1 + 도움말 2 + 여백 4 = 10)
		tableHeight := msg.Height - 10
		if tableHeight < 3 {
			tableHeight = 3
		}
		m.table.SetHeight(tableHeight)
		m.table.SetWidth(msg.Width - 4)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "Q", "esc":
			return m, func() tea.Msg { return backToMenuMsg{} }
		case "e":
			// 수정
			if m.tableReady && len(m.entries) > 0 {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.entries) {
					entry := m.entries[idx]
					return m, func() tea.Msg {
						return editEntryMsg{
							entry:       entry,
							accountsMap: m.accountsMap,
						}
					}
				}
			}
			return m, nil
		case "d":
			// 삭제 확인
			if m.tableReady && len(m.entries) > 0 {
				m.confirmDelete = true
				return m, nil
			}
			return m, nil
		case "r":
			if !m.loading && !m.loadingMore {
				m.rangeMode = true
				m.rangeCursor = 0
				return m, nil
			}
			return m, nil
		case "tab":
			if m.hasMore && !m.loadingMore && m.tableReady {
				m.loadingMore = true
				return m, m.fetchMoreEntries()
			}
			return m, nil
		}


	case transactionsLoadedMsg:
		m.entries = msg.entries
		m.accountsMap = msg.accountsMap
		m.loading = false
		m.loadingMore = false
		m.hasMore = msg.hasMore
		m.table.SetRows(m.buildTableRows())
		m.tableReady = true
		return m, nil

	case transactionsMoreLoadedMsg:
		m.loadingMore = false
		if len(msg.entries) == 0 {
			m.hasMore = false
			return m, nil
		}
		m.hasMore = len(msg.entries) >= pageSize
		m.entries = append(m.entries, msg.entries...)
		m.table.SetRows(m.buildTableRows())
		return m, nil

	case transactionsErrMsg:
		m.err = msg.err
		m.loading = false
		m.loadingMore = false
		return m, nil
	}

	// 테이블에 나머지 메시지 전달 (내비게이션)
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *transactionsSubModel) deleteEntry() tea.Cmd {
	idx := m.table.Cursor()
	if idx < 0 || idx >= len(m.entries) {
		return func() tea.Msg {
			return transactionDeleteErrMsg{err: fmt.Errorf("잘못된 선택")}
		}
	}
	entry := m.entries[idx]
	return func() tea.Msg {
		err := m.client.DeleteEntry(m.cfg.SectionID, entry.EntryID)
		if err != nil {
			return transactionDeleteErrMsg{err: err}
		}
		return transactionDeletedMsg{}
	}
}

func (m *transactionsSubModel) View() string {
	if m.loading {
		return m.renderLoading()
	}

	if m.err != nil {
		return m.renderError()
	}

	return m.renderTransactions()
}

func (m *transactionsSubModel) renderLoading() string {
	return titleStyle.Render("거래 내역 조회") + "\n\n거래 내역을 불러오는 중...\n"
}

func (m *transactionsSubModel) renderError() string {
	return titleStyle.Render("거래 내역 조회") + "\n\n" +
		errorStyle.Render("[오류] "+m.err.Error()) + "\n\n" +
		helpStyle.Render("[Enter] 메뉴로 돌아가기")
}

func (m *transactionsSubModel) renderTransactions() string {
	var content string

	// 헤더
	content += titleStyle.Render("거래 내역") + "\n"
	content += fmt.Sprintf("(%s ~ %s)\n\n", FormatDate(m.startDate), FormatDate(m.endDate))

	// 범위 선택 모드
	if m.rangeMode {
		return content + m.renderRangeMode()
	}

	// 거래 내역이 없는 경우
	if len(m.entries) == 0 {
		return content + "해당 기간에 거래 내역이 없습니다\n\n" +
			helpStyle.Render("[r] 범위  [q] 메뉴로 돌아가기")
	}

	// 테이블
	content += m.table.View() + "\n"

	// 합계
	totalAmount := 0.0
	for _, e := range m.entries {
		totalAmount += e.Money
	}
	countStr := fmt.Sprintf("%d", len(m.entries))
	if m.hasMore {
		countStr += "+"
	}
	content += fmt.Sprintf("총 %s건 | 합계 %s\n", countStr, FormatMoney(totalAmount))

	// 삭제 확인 프롬프트
	if m.confirmDelete {
		idx := m.table.Cursor()
		if idx >= 0 && idx < len(m.entries) {
			entry := m.entries[idx]
			content += "\n" + confirmStyle.Render(fmt.Sprintf("[%s] %s 삭제하시겠습니까? (Y/n)", entry.DateOnly(), entry.Item)) + "\n"
		}
		// 테이블은 유지하고 프롬프트만 추가 (화면 높이 일관성 유지)
	}

	// 더보기 로딩 중
	if m.loadingMore {
		content += loadingStyle.Render("추가 데이터를 불러오는 중...") + "\n"
	}

	// 삭제 중
	if m.deleting {
		content += "\n" + loadingStyle.Render("삭제 중...") + "\n"
	}

	// 삭제 에러
	if m.deleteErr != nil {
		content += "\n" + errorStyle.Render("[오류] "+m.deleteErr.Error()) + "\n"
		content += helpStyle.Render("[아무 키] 계속") + "\n"
	}

	// 도움말
	help := "[↑/↓/j/k] 이동  [e] 수정  [d] 삭제  [r] 범위  [q] 메뉴"
	if m.hasMore {
		help = "[Tab] 더보기  " + help
	}
	content += "\n" + helpStyle.Render(help)

	return content
}

// renderRangeMode는 범위 선택 UI를 렌더링
func (m *transactionsSubModel) renderRangeMode() string {
	if m.customInput {
		return m.renderCustomInput()
	}
	var content string
	for i, preset := range rangePresets {
		if i == m.rangeCursor {
			content += selectedStyle.Render("  > "+preset) + "\n"
		} else {
			content += "    " + preset + "\n"
		}
	}
	content += "\n" + helpStyle.Render("[↑/↓/j/k] 이동  [Enter] 선택  [esc] 취소")
	return content
}

// renderCustomInput은 직접 날짜 입력 UI를 렌더링
func (m *transactionsSubModel) renderCustomInput() string {
	var content string
	if m.inputField == 0 {
		content += "시작일 (YYYYMMDD): " + m.startInput.View() + "\n"
	} else {
		content += fmt.Sprintf("시작일: %s\n", FormatDate(m.startInput.Value()))
		content += "종료일 (YYYYMMDD): " + m.endInput.View() + "\n"
	}
	if m.inputErr != "" {
		content += "\n" + errorStyle.Render(m.inputErr) + "\n"
	}
	content += "\n" + helpStyle.Render("[Enter] 확인  [esc] 취소")
	return content
}
