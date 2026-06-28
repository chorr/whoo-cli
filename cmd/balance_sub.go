// cmd/balance_sub.go
// 자산부채 현황 + 자금증감 - bubbletea 서브 모델

package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"whoo-cli/api"
	"whoo-cli/config"
)

// balanceTab은 현재 활성 탭
type balanceTab int

const (
	balanceTabBS    balanceTab = iota // 잔액 (Balance Sheet)
	balanceTabInOut                   // 증감 (In/Out)
)

// balanceInOutPeriod는 In/Out 기간 선택
type balanceInOutPeriod int

const (
	inOutPeriodThisMonth balanceInOutPeriod = iota
	inOutPeriodLastMonth
	inOutPeriodCustom
)

// balanceMode는 서브 모델 내 화면 상태
type balanceMode int

const (
	balanceModeLoading  balanceMode = iota
	balanceModeBS                   // BS 탭 결과
	balanceModeInOutSel             // In/Out 기간 선택
	balanceModeInOut                // In/Out 결과
	balanceModeError
)

// balanceSubModel은 자산부채/자금증감 조회 서브 모델
type balanceSubModel struct {
	cfg    *config.Config
	client *api.WhooingClient

	// 공통
	accountsMap *api.AccountsMap
	mode        balanceMode
	tab         balanceTab
	errMsg      string
	width       int
	height      int

	// BS 탭
	bsResp  *api.BSResponse
	endDate string

	// In/Out 탭
	periodCursor  int // 0=이번달 1=지난달 2=직접입력
	customFrom    string
	customTo      string
	customInput   string
	customStep    int // 0=from 1=to
	inoutResp     *api.InOutResponse
	inoutStartDate string
	inoutEndDate   string
	rowCursor     int // In/Out 테이블 커서 (assets.accounts 행)
	inoutSection  int // 0=assets 1=liabilities
}

// newBalanceSubModel은 새로운 자산부채 모델을 생성
func newBalanceSubModel(cfg *config.Config) *balanceSubModel {
	return &balanceSubModel{
		cfg:     cfg,
		client:  NewClient(cfg),
		endDate: time.Now().Format("20060102"),
		mode:    balanceModeLoading,
		tab:     balanceTabBS,
	}
}

// ─── 메시지 타입 ──────────────────────────────────────────────

type balanceLoadedMsg struct {
	bsResp      *api.BSResponse
	accountsMap *api.AccountsMap
}

type balanceErrMsg struct{ err error }

type inOutLoadedMsg struct {
	resp      *api.InOutResponse
	startDate string
	endDate   string
}

// ─── Init / 비동기 커맨드 ────────────────────────────────────

func (m *balanceSubModel) Init() tea.Cmd {
	return m.fetchBS()
}

func (m *balanceSubModel) fetchBS() tea.Cmd {
	return func() tea.Msg {
		bsResp, err := m.client.GetBS(m.cfg.SectionID, m.endDate)
		if err != nil {
			return balanceErrMsg{err: err}
		}
		accountsMap, err := m.client.GetAccountsMap(m.cfg.SectionID)
		if err != nil {
			return balanceErrMsg{err: err}
		}
		return balanceLoadedMsg{bsResp: bsResp, accountsMap: accountsMap}
	}
}

func (m *balanceSubModel) fetchInOut(startDate, endDate string) tea.Cmd {
	return func() tea.Msg {
		resp, err := m.client.GetInOut(api.InOutQuery{
			SectionID: m.cfg.SectionID,
			StartDate: startDate,
			EndDate:   endDate,
		})
		if err != nil {
			return balanceErrMsg{err: err}
		}
		return inOutLoadedMsg{resp: resp, startDate: startDate, endDate: endDate}
	}
}

// ─── Update ──────────────────────────────────────────────────

func (m *balanceSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case balanceLoadedMsg:
		m.bsResp = msg.bsResp
		m.accountsMap = msg.accountsMap
		m.mode = balanceModeBS

	case inOutLoadedMsg:
		m.inoutResp = msg.resp
		m.inoutStartDate = msg.startDate
		m.inoutEndDate = msg.endDate
		m.rowCursor = 0
		m.inoutSection = 0
		m.mode = balanceModeInOut

	case balanceErrMsg:
		m.errMsg = msg.err.Error()
		m.mode = balanceModeError

	case tea.KeyMsg:
		if GlobalAction(msg) == ActionQuit {
			return m, tea.Quit
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *balanceSubModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case balanceModeBS:
		return m.handleBSKey(msg)
	case balanceModeInOutSel:
		return m.handleInOutSelKey(msg)
	case balanceModeInOut:
		return m.handleInOutKey(msg)
	case balanceModeError:
		switch ErrorAction(msg) {
		case ActionBack, ActionExit:
			return m, func() tea.Msg { return backToMenuMsg{} }
		}
	}
	return m, nil
}

func (m *balanceSubModel) handleBSKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// BS 탭은 최상위: esc/q/enter 모두 메뉴 복귀
	switch msg.String() {
	case "q", "esc", "enter":
		return m, func() tea.Msg { return backToMenuMsg{} }
	case "2", "right", "l":
		m.tab = balanceTabInOut
		m.mode = balanceModeInOutSel
		m.periodCursor = 0
		m.customFrom = ""
		m.customTo = ""
		m.customInput = ""
		m.customStep = 0
	}
	return m, nil
}

func (m *balanceSubModel) handleInOutSelKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch HorizontalSelectAction(msg) {
	case ActionBack:
		if m.customStep == 0 {
			m.tab = balanceTabBS
			m.mode = balanceModeBS
		}
		return m, nil
	case ActionExit:
		return m, func() tea.Msg { return backToMenuMsg{} }
	case ActionMoveLeft:
		if m.customStep == 0 {
			m.tab = balanceTabBS
			m.mode = balanceModeBS
		}
		return m, nil
	case ActionMoveUp:
		if m.periodCursor > 0 {
			m.periodCursor--
		}
	case ActionMoveDown:
		if m.periodCursor < 2 {
			m.periodCursor++
		}
	case ActionConfirm:
		return m.confirmInOutPeriod()
	}
	// 직접입력 모드 전용: backspace, 숫자 타이핑
	switch msg.String() {
	case "backspace":
		if m.periodCursor == 2 && len(m.customInput) > 0 {
			m.customInput = m.customInput[:len(m.customInput)-1]
		}
	default:
		if m.periodCursor == 2 && len(msg.String()) == 1 {
			r := rune(msg.String()[0])
			if r >= '0' && r <= '9' && len(m.customInput) < 8 {
				m.customInput += msg.String()
			}
		}
	}
	return m, nil
}

func (m *balanceSubModel) confirmInOutPeriod() (tea.Model, tea.Cmd) {
	now := time.Now()
	var start, end string

	switch m.periodCursor {
	case 0: // 이번달
		start = fmt.Sprintf("%d%02d01", now.Year(), now.Month())
		end = now.Format("20060102")
	case 1: // 지난달
		last := now.AddDate(0, -1, 0)
		start = fmt.Sprintf("%d%02d01", last.Year(), last.Month())
		// 지난달 마지막 날
		firstOfThis := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		end = firstOfThis.AddDate(0, 0, -1).Format("20060102")
	case 2: // 직접입력
		if m.customStep == 0 {
			// from 입력 완료 → to 입력으로
			if len(m.customInput) == 8 {
				m.customFrom = m.customInput
				m.customInput = ""
				m.customStep = 1
			}
			return m, nil
		}
		// to 입력 완료
		if len(m.customInput) == 8 {
			m.customTo = m.customInput
			m.customInput = ""
			m.customStep = 0
		} else {
			return m, nil
		}
		start = m.customFrom
		end = m.customTo
	}

	m.mode = balanceModeLoading
	return m, m.fetchInOut(start, end)
}

func (m *balanceSubModel) handleInOutKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	accounts := m.currentInOutAccounts()
	switch HorizontalSelectAction(msg) {
	case ActionBack:
		m.mode = balanceModeInOutSel
		m.periodCursor = 0
		return m, nil
	case ActionExit:
		return m, func() tea.Msg { return backToMenuMsg{} }
	case ActionMoveLeft:
		m.tab = balanceTabBS
		m.mode = balanceModeBS
		return m, nil
	case ActionMoveUp:
		if m.rowCursor > 0 {
			m.rowCursor--
		}
	case ActionMoveDown:
		if m.rowCursor < len(accounts)-1 {
			m.rowCursor++
		}
	}
	switch msg.String() {
	case "1":
		m.tab = balanceTabBS
		m.mode = balanceModeBS
		return m, nil
	case "tab":
		// 자산/부채 섹션 전환
		m.inoutSection = 1 - m.inoutSection
		m.rowCursor = 0
	case "r":
		m.mode = balanceModeLoading
		return m, m.fetchInOut(m.inoutStartDate, m.inoutEndDate)
	}
	return m, nil
}

// currentInOutAccounts는 현재 섹션의 계정 목록 반환
func (m *balanceSubModel) currentInOutAccounts() []api.InOutAccount {
	if m.inoutResp == nil {
		return nil
	}
	if m.inoutSection == 0 {
		return m.inoutResp.Assets.Accounts
	}
	return m.inoutResp.Liabilities.Accounts
}

// ─── View ────────────────────────────────────────────────────

func (m *balanceSubModel) View() string {
	switch m.mode {
	case balanceModeLoading:
		return titleStyle.Render("자산부채 현황") + "\n\n데이터를 불러오는 중...\n"
	case balanceModeError:
		return titleStyle.Render("자산부채 현황") + "\n\n" +
			errorStyle.Render("[오류] "+m.errMsg) + "\n\n" +
			helpStyle.Render("[Enter/q] 메뉴로 돌아가기")
	case balanceModeBS:
		return m.renderBS()
	case balanceModeInOutSel:
		return m.renderInOutSel()
	case balanceModeInOut:
		return m.renderInOut()
	}
	return ""
}

func (m *balanceSubModel) renderTabBar() string {
	bs := "[1] 잔액(BS)"
	io := "[2] 증감(In/Out)"
	if m.tab == balanceTabBS {
		return selectedStyle.Render(bs) + "  " + io
	}
	return bs + "  " + selectedStyle.Render(io)
}

func (m *balanceSubModel) renderBS() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("자산부채 현황") + "\n")
	b.WriteString(m.renderTabBar() + "\n\n")
	b.WriteString(fmt.Sprintf("(%s 기준)\n\n", FormatDate(m.endDate)))

	if m.bsResp == nil || (len(m.bsResp.Assets.Accounts) == 0 && len(m.bsResp.Liabilities.Accounts) == 0) {
		b.WriteString("자산부채 데이터가 없습니다\n\n")
		b.WriteString(helpStyle.Render("[2/→/l] In/Out  [Enter/q] 메뉴"))
		return b.String()
	}

	var tableBuf strings.Builder
	w := tabwriter.NewWriter(&tableBuf, 0, 0, 2, ' ', 0)

	if len(m.bsResp.Assets.Accounts) > 0 {
		fmt.Fprintln(w, "[자산]")
		m.printBSGroup(w, m.bsResp.Assets, "assets")
		fmt.Fprintf(w, "  소계\t%s\n", FormatMoney(m.bsResp.Assets.Total))
		fmt.Fprintln(w)
	}
	if len(m.bsResp.Liabilities.Accounts) > 0 {
		fmt.Fprintln(w, "[부채]")
		m.printBSGroup(w, m.bsResp.Liabilities, "liabilities")
		fmt.Fprintf(w, "  소계\t%s\n", FormatMoney(m.bsResp.Liabilities.Total))
		fmt.Fprintln(w)
	}
	w.Flush()
	b.WriteString(tableBuf.String())

	netWorth := m.bsResp.Assets.Total - m.bsResp.Liabilities.Total
	b.WriteString(fmt.Sprintf("순자산: %s\n", FormatMoney(netWorth)))
	b.WriteString("\n" + helpStyle.Render("[2/→/l] In/Out  [Enter/q] 메뉴"))
	return b.String()
}

func (m *balanceSubModel) printBSGroup(w *tabwriter.Writer, group api.BSGroup, accountType string) {
	for _, item := range group.Accounts {
		if item.Money == 0 {
			continue
		}
		title := m.accountsMap.GetTitle(accountType, item.AccountID)
		fmt.Fprintf(w, "  %s\t%s\n", title, FormatMoney(item.Money))
	}
}

func (m *balanceSubModel) renderInOutSel() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("자산부채 현황") + "\n")
	b.WriteString(m.renderTabBar() + "\n\n")
	b.WriteString(headerStyle.Render("조회 기간 선택") + "\n\n")

	periods := []string{"이번달", "지난달", "직접입력"}
	for i, p := range periods {
		if i == m.periodCursor {
			b.WriteString(selectedStyle.Render("> "+p) + "\n")
		} else {
			b.WriteString("  " + p + "\n")
		}
	}

	if m.periodCursor == 2 {
		b.WriteString("\n")
		if m.customStep == 0 {
			b.WriteString(fmt.Sprintf("시작일 (YYYYMMDD): %s_\n", m.customInput))
		} else {
			b.WriteString(fmt.Sprintf("시작일: %s\n", m.customFrom))
			b.WriteString(fmt.Sprintf("종료일 (YYYYMMDD): %s_\n", m.customInput))
		}
	}

	b.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [Enter] 선택  [1/←] BS  [Esc] BS로"))
	return b.String()
}

func (m *balanceSubModel) renderInOut() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("자산부채 현황") + "\n")
	b.WriteString(m.renderTabBar() + "\n\n")

	if m.inoutResp == nil {
		b.WriteString("데이터 없음\n")
		return b.String()
	}

	period := fmt.Sprintf("%s ~ %s", FormatDate(m.inoutStartDate), FormatDate(m.inoutEndDate))
	b.WriteString(fmt.Sprintf("기간: %s\n\n", period))

	// 섹션 탭 (자산/부채)
	assetTab := "[자산]"
	liabTab := "[부채]"
	if m.inoutSection == 0 {
		assetTab = selectedStyle.Render(assetTab)
	} else {
		liabTab = selectedStyle.Render(liabTab)
	}
	b.WriteString(assetTab + "  " + liabTab + "\n\n")

	m.writeInOutTable(&b)

	b.WriteString("\n" + helpStyle.Render("[Tab] 자산/부채  [↑/↓/j/k] 이동  [r] 새로고침  [1/←] BS  [Esc] 기간선택  [q] 메뉴"))
	return b.String()
}

func (m *balanceSubModel) writeInOutTable(b *strings.Builder) {
	var group api.InOutGroup
	var accountType string
	if m.inoutSection == 0 {
		group = m.inoutResp.Assets
		accountType = "assets"
	} else {
		group = m.inoutResp.Liabilities
		accountType = "liabilities"
	}

	var tableBuf strings.Builder
	w := tabwriter.NewWriter(&tableBuf, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "  항목\t유입(in)\t유출(out)\t순증감(margin)")
	fmt.Fprintln(w, "  ----\t--------\t---------\t--------------")

	for i, acc := range group.Accounts {
		title := ""
		if m.accountsMap != nil {
			title = m.accountsMap.GetTitle(accountType, acc.AccountID)
		}
		if title == "" {
			title = acc.AccountID
		}
		row := fmt.Sprintf("  %s\t%s\t%s\t%s",
			title,
			FormatMoney(float64(acc.In)),
			FormatMoney(float64(acc.Out)),
			FormatMoney(float64(acc.Margin)),
		)
		if i == m.rowCursor {
			fmt.Fprintln(w, selectedStyle.Render(">"+row[1:]))
		} else {
			fmt.Fprintln(w, row)
		}
	}

	// 합계
	t := group.Total
	fmt.Fprintf(w, "  합계\t%s\t%s\t%s\n",
		FormatMoney(float64(t.In)),
		FormatMoney(float64(t.Out)),
		FormatMoney(float64(t.Margin)),
	)
	w.Flush()
	b.WriteString(tableBuf.String())
}
