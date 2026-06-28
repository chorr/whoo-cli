// cmd/budget_sub.go
// 예산·목표 - bubbletea 서브 모델 (3탭: 월별 예산 / 장기목표 / 자본 목표)

package cmd

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"whoo-cli/api"
	"whoo-cli/config"
)

// ─── 모드 및 탭 상수 ─────────────────────────────────────────

type budgetTab int

const (
	budgetTabBudget  budgetTab = iota // 월별 예산
	budgetTabGoal                     // 장기목표
	budgetTabCapital                  // 자본 목표
)

type budgetMode int

const (
	budgetModeLoading      budgetMode = iota
	budgetModeView                     // 현재 탭 데이터 표시
	budgetModeTypeSelect               // 탭1: expenses/income 선택 (가로)
	budgetModeEdit                     // 탭1: 예산 편집 중
	budgetModeEditConfirm              // 탭1: budgetLong 경고 모달
	budgetModeError
)

// ─── 서브 모델 ───────────────────────────────────────────────

type budgetSubModel struct {
	cfg    *config.Config
	client *api.WhooingClient

	mode   budgetMode
	tab    budgetTab
	errMsg string

	// 섹션 budgetLong 여부
	budgetLong bool

	// 탭1: 월별 예산
	typeCursor   int    // 0=expenses, 1=income
	accountType  string // "expenses" | "income"
	budgetResp   *api.BudgetResponse
	budgetRows   []api.BudgetLineWithID // aggregate.accounts 정렬된 슬라이스
	rowCursor    int
	editInput    string
	editTargetID string // 편집 중인 account_id
	editTargetYM int    // 편집 대상 월 YYYYMM
	accountsMap  *api.AccountsMap

	// 탭2: 장기목표
	goalResp *api.BudgetGoalResponse

	// 탭3: 자본 목표
	capitalGoal api.GoalMap
	capitalKeys []string // 정렬된 YYYYMM 키
}

func newBudgetSubModel(cfg *config.Config) *budgetSubModel {
	return &budgetSubModel{
		cfg:         cfg,
		client:      NewClient(cfg),
		mode:        budgetModeTypeSelect,
		tab:         budgetTabBudget,
		accountType: "expenses",
	}
}

// ─── 메시지 타입 ──────────────────────────────────────────────

type budgetLoadedMsg struct {
	resp        *api.BudgetResponse
	accountsMap *api.AccountsMap
	budgetLong  bool
}

type budgetGoalLoadedMsg struct{ resp *api.BudgetGoalResponse }
type budgetCapitalLoadedMsg struct{ goal api.GoalMap; keys []string }
type budgetErrMsg struct{ err error }
type budgetUpdateDoneMsg struct{ feedback string }

// ─── Init / 비동기 커맨드 ────────────────────────────────────

func (m *budgetSubModel) Init() tea.Cmd {
	return nil
}

func (m *budgetSubModel) loadBudget() tea.Cmd {
	accountType := m.accountType
	sectionID := m.cfg.SectionID
	return func() tea.Msg {
		now := time.Now()
		ym := int(now.Year())*100 + int(now.Month())

		resp, err := m.client.GetBudget(sectionID, accountType, ym, ym)
		if err != nil {
			return budgetErrMsg{err: err}
		}

		accountsMap, err := m.client.GetAccountsMap(sectionID)
		if err != nil {
			return budgetErrMsg{err: err}
		}

		// 섹션 ui.budgetLong 확인 — GetSections 캐시 활용
		sections, _ := m.client.GetSections()
		budgetLong := false
		for _, s := range sections {
			if s.SectionID == sectionID {
				budgetLong = s.UI.BudgetLong == "y"
				break
			}
		}

		return budgetLoadedMsg{resp: resp, accountsMap: accountsMap, budgetLong: budgetLong}
	}
}

func (m *budgetSubModel) loadBudgetGoal() tea.Cmd {
	sectionID := m.cfg.SectionID
	return func() tea.Msg {
		resp, err := m.client.GetBudgetGoal(sectionID)
		if err != nil {
			return budgetErrMsg{err: err}
		}
		return budgetGoalLoadedMsg{resp: resp}
	}
}

func (m *budgetSubModel) loadCapitalGoal() tea.Cmd {
	sectionID := m.cfg.SectionID
	return func() tea.Msg {
		now := time.Now()
		from := int(now.Year())*100 + int(now.Month())
		to := from + 11
		goal, err := m.client.GetGoal(sectionID, from, to)
		if err != nil {
			return budgetErrMsg{err: err}
		}
		keys := make([]string, 0, len(goal))
		for k := range goal {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return budgetCapitalLoadedMsg{goal: goal, keys: keys}
	}
}

func (m *budgetSubModel) doUpdateBudget(accountID string, amount int64) tea.Cmd {
	sectionID := m.cfg.SectionID
	accountType := m.accountType
	targetYM := m.editTargetYM
	return func() tea.Msg {
		_, err := m.client.UpdateBudget(sectionID, accountType, targetYM, map[string]int64{accountID: amount})
		if err != nil {
			return budgetErrMsg{err: err}
		}
		return budgetUpdateDoneMsg{feedback: "예산이 수정되었습니다"}
	}
}

// ─── Update ──────────────────────────────────────────────────

func (m *budgetSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case budgetLoadedMsg:
		m.budgetResp = msg.resp
		m.accountsMap = msg.accountsMap
		m.budgetLong = msg.budgetLong
		m.budgetRows = sortedBudgetAccounts(msg.resp.Aggregate.Accounts)
		m.rowCursor = 0
		m.mode = budgetModeView

	case budgetGoalLoadedMsg:
		m.goalResp = msg.resp
		m.mode = budgetModeView

	case budgetCapitalLoadedMsg:
		m.capitalGoal = msg.goal
		m.capitalKeys = msg.keys
		m.mode = budgetModeView

	case budgetErrMsg:
		m.errMsg = msg.err.Error()
		m.mode = budgetModeError

	case budgetUpdateDoneMsg:
		m.mode = budgetModeLoading
		return m, m.loadBudget()

	case tea.KeyMsg:
		if GlobalAction(msg) == ActionQuit {
			return m, tea.Quit
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *budgetSubModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case budgetModeTypeSelect:
		return m.handleTypeSelectKey(msg)
	case budgetModeView:
		return m.handleViewKey(msg)
	case budgetModeEdit:
		return m.handleEditKey(msg)
	case budgetModeEditConfirm:
		return m.handleEditConfirmKey(msg)
	case budgetModeError:
		switch ErrorAction(msg) {
		case ActionBack, ActionExit:
			return m, func() tea.Msg { return backToMenuMsg{} }
		}
	}
	return m, nil
}

func (m *budgetSubModel) handleTypeSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch HorizontalSelectAction(msg) {
	case ActionBack, ActionExit:
		return m, func() tea.Msg { return backToMenuMsg{} }
	case ActionMoveLeft:
		if m.typeCursor > 0 {
			m.typeCursor--
		}
	case ActionMoveRight:
		if m.typeCursor < 1 {
			m.typeCursor++
		}
	case ActionConfirm:
		types := []string{"expenses", "income"}
		m.accountType = types[m.typeCursor]
		m.mode = budgetModeLoading
		return m, m.loadBudget()
	}
	// 탭 전환 도메인 키
	switch msg.String() {
	case "2":
		m.tab = budgetTabGoal
		m.mode = budgetModeLoading
		return m, m.loadBudgetGoal()
	case "3":
		m.tab = budgetTabCapital
		m.mode = budgetModeLoading
		return m, m.loadCapitalGoal()
	}
	return m, nil
}

func (m *budgetSubModel) handleViewKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, func() tea.Msg { return backToMenuMsg{} }
	case "esc":
		if m.tab == budgetTabBudget {
			m.mode = budgetModeTypeSelect
		} else {
			m.tab = budgetTabBudget
			m.mode = budgetModeTypeSelect
		}
		return m, nil
	case "1":
		m.tab = budgetTabBudget
		if m.budgetResp != nil {
			m.mode = budgetModeView
		} else {
			m.mode = budgetModeTypeSelect
		}
		return m, nil
	case "2":
		m.tab = budgetTabGoal
		m.mode = budgetModeLoading
		return m, m.loadBudgetGoal()
	case "3":
		m.tab = budgetTabCapital
		m.mode = budgetModeLoading
		return m, m.loadCapitalGoal()
	case "r":
		m.mode = budgetModeLoading
		return m, m.currentLoadCmd()
	}

	// 탭별 키 처리
	if m.tab == budgetTabBudget {
		return m.handleBudgetRowKey(msg)
	}
	return m, nil
}

func (m *budgetSubModel) handleBudgetRowKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	rows := m.budgetRows
	switch ListAction(msg) {
	case ActionMoveUp:
		if m.rowCursor > 0 {
			m.rowCursor--
		}
	case ActionMoveDown:
		if m.rowCursor < len(rows)-1 {
			m.rowCursor++
		}
	case ActionEdit:
		if len(rows) == 0 {
			return m, nil
		}
		m.editTargetID = rows[m.rowCursor].AccountID
		now := time.Now()
		m.editTargetYM = int(now.Year())*100 + int(now.Month())
		m.editInput = fmt.Sprintf("%d", rows[m.rowCursor].Budget)
		if m.budgetLong {
			m.mode = budgetModeEditConfirm
		} else {
			m.mode = budgetModeEdit
		}
	}
	return m, nil
}

func (m *budgetSubModel) handleEditConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch ConfirmAction(msg) {
	case ActionConfirm: // y/Y
		m.mode = budgetModeEdit
	case ActionBack: // n/N/esc
		m.mode = budgetModeView
	default:
		// enter도 진행 (비파괴적 작업이므로 허용)
		if msg.String() == "enter" {
			m.mode = budgetModeEdit
		}
	}
	return m, nil
}

func (m *budgetSubModel) handleEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.mode = budgetModeView
		m.editInput = ""
		return m, nil
	case tea.KeyEnter:
		if m.editInput == "" {
			return m, nil
		}
		var amount int64
		_, err := fmt.Sscanf(m.editInput, "%d", &amount)
		if err != nil || amount < 0 {
			m.editInput = ""
			return m, nil
		}
		m.mode = budgetModeLoading
		return m, m.doUpdateBudget(m.editTargetID, amount)
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.editInput) > 0 {
			m.editInput = m.editInput[:len(m.editInput)-1]
		}
	case tea.KeyRunes:
		for _, r := range msg.Runes {
			if r >= '0' && r <= '9' {
				m.editInput += string(r)
			}
		}
	}
	return m, nil
}

func (m *budgetSubModel) currentLoadCmd() tea.Cmd {
	switch m.tab {
	case budgetTabBudget:
		return m.loadBudget()
	case budgetTabGoal:
		return m.loadBudgetGoal()
	case budgetTabCapital:
		return m.loadCapitalGoal()
	}
	return nil
}

// ─── View ────────────────────────────────────────────────────

func (m *budgetSubModel) View() string {
	switch m.mode {
	case budgetModeLoading:
		return titleStyle.Render("예산·목표") + "\n\n불러오는 중...\n"
	case budgetModeError:
		return titleStyle.Render("예산·목표") + "\n\n" +
			errorStyle.Render("[오류] "+m.errMsg) + "\n\n" +
			helpStyle.Render("[Enter/q] 메뉴로 돌아가기")
	case budgetModeTypeSelect:
		return m.renderTypeSelect()
	case budgetModeView:
		return m.renderView()
	case budgetModeEdit:
		return m.renderEdit()
	case budgetModeEditConfirm:
		return m.renderEditConfirm()
	}
	return ""
}

func (m *budgetSubModel) renderTabBar() string {
	tabs := []string{"[1] 월별 예산", "[2] 장기목표", "[3] 자본 목표"}
	parts := make([]string, len(tabs))
	for i, t := range tabs {
		if budgetTab(i) == m.tab {
			parts[i] = selectedStyle.Render(t)
		} else {
			parts[i] = t
		}
	}
	return strings.Join(parts, "  ")
}

func (m *budgetSubModel) renderTypeSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("예산·목표") + "\n")
	b.WriteString(m.renderTabBar() + "\n\n")
	b.WriteString(headerStyle.Render("계정 유형 선택") + "\n\n")

	types := []string{"지출(expenses)", "수입(income)"}
	for i, t := range types {
		if i == m.typeCursor {
			b.WriteString(selectedStyle.Render("> "+t) + "\n")
		} else {
			b.WriteString("  " + t + "\n")
		}
	}
	b.WriteString("\n" + helpStyle.Render("[←/→/h/l] 이동  [Enter] 선택  [2] 장기목표  [3] 자본목표  [q] 메뉴"))
	return b.String()
}

func (m *budgetSubModel) renderView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("예산·목표") + "\n")
	b.WriteString(m.renderTabBar() + "\n\n")

	switch m.tab {
	case budgetTabBudget:
		m.writeBudgetView(&b)
	case budgetTabGoal:
		m.writeBudgetGoalView(&b)
	case budgetTabCapital:
		m.writeCapitalGoalView(&b)
	}
	return b.String()
}

func (m *budgetSubModel) writeBudgetView(b *strings.Builder) {
	if m.budgetResp == nil {
		b.WriteString("데이터 없음\n")
		return
	}

	now := time.Now()
	ym := fmt.Sprintf("%d년 %02d월", now.Year(), now.Month())
	typeLabel := "지출"
	if m.accountType == "income" {
		typeLabel = "수입"
	}
	b.WriteString(fmt.Sprintf("%s %s 예산\n\n", ym, typeLabel))

	agg := m.budgetResp.Aggregate

	// 가능성 바
	poss := agg.Misc.Possibility
	b.WriteString(renderPossibilityBar(poss))
	b.WriteString(fmt.Sprintf("예산 %s원  실적 %s원  잔여 %s원\n\n",
		FormatMoney(float64(agg.Total.Budget)),
		FormatMoney(float64(agg.Total.Money)),
		FormatMoney(float64(agg.Total.Remains)),
	))

	// 계정별 테이블
	if len(m.budgetRows) == 0 {
		b.WriteString("계정 데이터 없음\n")
	} else {
		var tableBuf strings.Builder
		w := tabwriter.NewWriter(&tableBuf, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "  항목\t예산\t실적\t잔여")
		fmt.Fprintln(w, "  ----\t----\t----\t----")
		for i, row := range m.budgetRows {
			title := row.AccountID
			if m.accountsMap != nil {
				t := m.accountsMap.GetTitle(m.accountType, row.AccountID)
				if t != "" {
					title = t
				}
			}
			line := fmt.Sprintf("  %s\t%s\t%s\t%s",
				title,
				FormatMoney(float64(row.Budget)),
				FormatMoney(float64(row.Money)),
				FormatMoney(float64(row.Remains)),
			)
			if i == m.rowCursor {
				fmt.Fprintln(w, selectedStyle.Render("> "+line[2:]))
			} else {
				fmt.Fprintln(w, line)
			}
		}
		w.Flush()
		b.WriteString(tableBuf.String())
	}

	b.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [e] 예산 편집  [r] 새로고침  [2] 장기목표  [3] 자본목표  [Esc] 유형선택  [q] 메뉴"))
}

func (m *budgetSubModel) writeBudgetGoalView(b *strings.Builder) {
	if m.goalResp == nil {
		b.WriteString("장기목표가 설정되어 있지 않습니다\n")
		b.WriteString("\n" + helpStyle.Render("[1] 월별 예산  [3] 자본 목표  [r] 새로고침  [q] 메뉴"))
		return
	}

	g := m.goalResp
	b.WriteString(headerStyle.Render("장기 예산목표") + "\n\n")

	var tableBuf strings.Builder
	w := tabwriter.NewWriter(&tableBuf, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  시작월\t%d\n", g.BaseYM)
	fmt.Fprintf(w, "  목표월\t%d\n", g.GoalYM)
	fmt.Fprintf(w, "  현재 자본\t%s원\n", FormatMoney(float64(g.BaseMoney)))
	fmt.Fprintf(w, "  목표 자본\t%s원\n", FormatMoney(float64(g.GoalMoney)))
	fmt.Fprintf(w, "  월 수입\t%s원\n", FormatMoney(float64(g.BaseIncome)))
	fmt.Fprintf(w, "  월 지출\t%s원\n", FormatMoney(float64(g.BaseExpenses)))
	fmt.Fprintf(w, "  배분 방식\t%s\n", g.SplitType)
	w.Flush()
	b.WriteString(tableBuf.String())

	progress := float64(0)
	if g.GoalMoney > g.BaseMoney {
		needed := g.GoalMoney - g.BaseMoney
		monthly := g.BaseIncome - g.BaseExpenses
		if monthly > 0 {
			months := float64(needed) / float64(monthly)
			_ = months
		}
		progress = float64(g.BaseMoney) / float64(g.GoalMoney) * 100
	}
	b.WriteString(fmt.Sprintf("\n목표 달성률: %s\n", renderPossibilityBar(progress)))
	b.WriteString("\n" + helpStyle.Render("[1] 월별 예산  [3] 자본 목표  [r] 새로고침  [q] 메뉴"))
}

func (m *budgetSubModel) writeCapitalGoalView(b *strings.Builder) {
	if len(m.capitalGoal) == 0 {
		b.WriteString("자본 목표가 설정되어 있지 않습니다\n")
		b.WriteString("(장기목표 미사용 섹션이거나 목표가 없습니다)\n")
		b.WriteString("\n" + helpStyle.Render("[1] 월별 예산  [2] 장기목표  [r] 새로고침  [q] 메뉴"))
		return
	}

	b.WriteString(headerStyle.Render("월별 자본 목표") + "\n\n")

	maxMoney := int64(0)
	for _, v := range m.capitalGoal {
		if v > maxMoney {
			maxMoney = v
		}
	}

	for _, k := range m.capitalKeys {
		v := m.capitalGoal[k]
		bar := renderMoneyBar(v, maxMoney, 20)
		ym := k
		if len(ym) == 6 {
			ym = ym[:4] + "-" + ym[4:]
		}
		b.WriteString(fmt.Sprintf("  %s  %s  %s원\n", ym, bar, FormatMoney(float64(v))))
	}

	b.WriteString("\n" + helpStyle.Render("[1] 월별 예산  [2] 장기목표  [r] 새로고침  [q] 메뉴"))
}

func (m *budgetSubModel) renderEdit() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("예산·목표") + "\n")
	b.WriteString(m.renderTabBar() + "\n\n")
	b.WriteString(headerStyle.Render("예산 편집") + "\n\n")

	title := m.editTargetID
	if m.accountsMap != nil {
		t := m.accountsMap.GetTitle(m.accountType, m.editTargetID)
		if t != "" {
			title = t
		}
	}
	b.WriteString(fmt.Sprintf("항목: %s\n", title))
	b.WriteString(fmt.Sprintf("대상 월: %d\n\n", m.editTargetYM))
	b.WriteString("새 예산액: " + m.editInput + "_\n\n")
	b.WriteString(helpStyle.Render("[Enter] 저장  [Esc] 취소"))
	return b.String()
}

func (m *budgetSubModel) renderEditConfirm() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("예산·목표") + "\n\n")
	b.WriteString(errorStyle.Render("[주의] 이 섹션은 장기목표 연동 모드입니다") + "\n\n")
	b.WriteString("예산을 수정하면 이 월 이후의 모든 자본 목표(goal)가\n자동으로 재계산됩니다.\n\n")
	b.WriteString("계속 진행하시겠습니까?\n\n")
	b.WriteString(helpStyle.Render("[y/Enter] 계속  [n/Esc] 취소"))
	return b.String()
}

// ─── 헬퍼 함수 ──────────────────────────────────────────────

// sortedBudgetAccounts는 accounts map을 예산 내림차순 정렬 슬라이스로 변환
func sortedBudgetAccounts(accounts map[string]api.BudgetLineWithID) []api.BudgetLineWithID {
	rows := make([]api.BudgetLineWithID, 0, len(accounts))
	for _, v := range accounts {
		rows = append(rows, v)
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Budget > rows[j].Budget
	})
	return rows
}

// renderPossibilityBar는 0~100 범위의 가능성 지수를 ASCII 바로 표시
func renderPossibilityBar(poss float64) string {
	const width = 20
	if poss > 100 {
		poss = 100
	}
	filled := int(poss / 100 * width)
	bar := strings.Repeat("#", filled) + strings.Repeat(".", width-filled)
	return fmt.Sprintf("가능성 [%s] %.0f%%\n", bar, poss)
}

// renderMoneyBar는 금액을 상대적 막대로 표시
func renderMoneyBar(amount, maxAmount int64, width int) string {
	if maxAmount == 0 {
		return strings.Repeat(".", width)
	}
	filled := int(float64(amount) / float64(maxAmount) * float64(width))
	if filled > width {
		filled = width
	}
	return strings.Repeat("#", filled) + strings.Repeat(".", width-filled)
}

// FormatYM은 YYYYMM 정수를 YYYY-MM 문자열로 변환
func FormatYM(ym int) string {
	return fmt.Sprintf("%d-%02d", ym/100, ym%100)
}
