// cmd/monthly_sub.go
// 월별입력 관리 - bubbletea 서브 모델

package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"whoo-cli/api"
	"whoo-cli/config"
)

type monthlyMode int

const (
	monthlyModeSlotSelect    monthlyMode = iota // 슬롯 선택
	monthlyModeList                             // 항목 목록
	monthlyModeAdd                              // 추가 폼 (list 'a')
	monthlyModeEdit                             // 수정 폼 (list 'e')
	monthlyModeConfirmDelete                    // 삭제 확인
	monthlyModeLoading
	monthlyModeError
)

const (
	monthlyFormStepItem = iota
	monthlyFormStepMoney
	monthlyFormStepLAccount
	monthlyFormStepLAccountID
	monthlyFormStepRAccount
	monthlyFormStepRAccountID
	monthlyFormStepPayDate
	monthlyFormStepSkipHoliday
	monthlyFormStepConfirm
)

// monthlyItem은 TUI에서 표시하는 월별입력 항목
type monthlyItem struct {
	ID          string
	Item        string
	Money       int64
	PayDate     int
	DueDate     string
	DDay        int
	PaidDate    string
	LAccount    string
	LAccountID  string
	RAccount    string
	RAccountID  string
	SkipHoliday string
}

// monthlyListItem은 bubbles/list용 항목 래퍼
type monthlyListItem struct {
	item monthlyItem
}

func (i monthlyListItem) Title() string {
	badge := formatDDayBadge(i.item.DDay, i.item.PaidDate)
	return fmt.Sprintf("%-20s  %2d일  %s원  %s",
		i.item.Item, i.item.PayDate, FormatMoney(float64(i.item.Money)), badge)
}
func (i monthlyListItem) Description() string { return "" }
func (i monthlyListItem) FilterValue() string { return i.item.Item }

type monthlySubModel struct {
	cfg    *config.Config
	client *api.WhooingClient
	mode   monthlyMode
	errMsg string

	// 슬롯 탭 (가로 탐색, 3개 고정 → 수동 커서)
	slotCursor int
	slots      []string
	activeSlot string

	// 항목 목록 (bubbles/list)
	items       []monthlyItem // 원본 데이터 (선택 시 참조)
	monthlyList list.Model

	// 피드백
	feedback string

	// 폼 입력 (add/edit)
	formStep        int
	editingID       string
	formItem        string
	formMoney       string
	formLAccount    string
	formLAccountID  string
	formRAccount    string
	formRAccountID  string
	formPayDate     string
	formSkipHoliday string
	textInput       string
}

func newMonthlySubModel(cfg *config.Config) *monthlySubModel {
	return &monthlySubModel{
		cfg:    cfg,
		client: NewClient(cfg),
		mode:   monthlyModeSlotSelect,
		slots:  []string{"slot1", "slot2", "slot3"},
	}
}

// ─── 비동기 커맨드 ──────────────────────────────────────────────

type monthlyItemsLoadedMsg struct {
	slot  string
	items []monthlyItem
	err   error
}

type monthlyActionDoneMsg struct {
	feedback string
	err      error
}

func (m *monthlySubModel) loadItems(slot string) tea.Cmd {
	return func() tea.Msg {
		raw, err := m.client.GetMonthlyItemsSlot(m.cfg.SectionID, slot)
		if err != nil {
			return monthlyItemsLoadedMsg{slot: slot, err: err}
		}
		items := parseMonthlyItemsFromRaw(raw)
		return monthlyItemsLoadedMsg{slot: slot, items: items}
	}
}

func (m *monthlySubModel) doDelete(slot, itemID string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.DeleteMonthlyItem(m.cfg.SectionID, slot, itemID)
		if err != nil {
			return monthlyActionDoneMsg{err: err}
		}
		return monthlyActionDoneMsg{feedback: "삭제되었습니다"}
	}
}

func (m *monthlySubModel) buildMonthlyList(items []monthlyItem) list.Model {
	listItems := make([]list.Item, len(items))
	for i, it := range items {
		listItems[i] = monthlyListItem{item: it}
	}
	return newPlainList(listItems, 70, len(listItems)+2)
}

// ─── BubbleTea ────────────────────────────────────────────────

func (m *monthlySubModel) Init() tea.Cmd {
	return nil
}

func (m *monthlySubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case monthlyItemsLoadedMsg:
		if msg.err != nil {
			m.mode = monthlyModeError
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.activeSlot = msg.slot
		m.items = msg.items
		m.monthlyList = m.buildMonthlyList(msg.items)
		m.mode = monthlyModeList
		return m, nil

	case monthlyActionDoneMsg:
		if msg.err != nil {
			m.mode = monthlyModeError
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.feedback = msg.feedback
		m.mode = monthlyModeLoading
		return m, m.loadItems(m.activeSlot)

	case tea.KeyMsg:
		if GlobalAction(msg) == ActionQuit {
			return m, tea.Quit
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *monthlySubModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case monthlyModeSlotSelect:
		return m.handleSlotKey(msg)
	case monthlyModeList:
		return m.handleListKey(msg)
	case monthlyModeConfirmDelete:
		return m.handleConfirmKey(msg)
	case monthlyModeAdd, monthlyModeEdit:
		return m.handleFormKey(msg)
	case monthlyModeError:
		switch ErrorAction(msg) {
		case ActionBack, ActionExit:
			return m, func() tea.Msg { return backToMenuMsg{} }
		}
	}
	return m, nil
}

func (m *monthlySubModel) handleSlotKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch HorizontalSelectAction(msg) {
	case ActionBack, ActionExit:
		return m, func() tea.Msg { return backToMenuMsg{} }
	case ActionConfirm:
		slot := m.slots[m.slotCursor]
		m.activeSlot = slot
		m.mode = monthlyModeLoading
		return m, m.loadItems(slot)
	case ActionMoveLeft:
		if m.slotCursor > 0 {
			m.slotCursor--
		}
	case ActionMoveRight:
		if m.slotCursor < len(m.slots)-1 {
			m.slotCursor++
		}
	}
	return m, nil
}

func (m *monthlySubModel) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch ListAction(msg) {
	case ActionBack:
		m.mode = monthlyModeSlotSelect
		m.feedback = ""
		return m, nil
	case ActionEdit:
		if len(m.items) > 0 {
			idx := m.monthlyList.Index()
			it := m.items[idx]
			m.feedback = ""
			m.mode = monthlyModeEdit
			m.editingID = it.ID
			m.formStep = monthlyFormStepItem
			m.formItem = it.Item
			m.formMoney = fmt.Sprintf("%d", it.Money)
			m.formLAccount = it.LAccount
			m.formLAccountID = it.LAccountID
			m.formRAccount = it.RAccount
			m.formRAccountID = it.RAccountID
			m.formPayDate = ""
			if it.PayDate > 0 {
				m.formPayDate = fmt.Sprintf("%d", it.PayDate)
			}
			m.formSkipHoliday = it.SkipHoliday
			m.textInput = m.formItem
		}
		return m, nil
	case ActionDelete:
		if len(m.items) > 0 {
			m.mode = monthlyModeConfirmDelete
		}
		return m, nil
	case ActionExit:
		return m, func() tea.Msg { return backToMenuMsg{} }
	}
	// 도메인 전용: a = 추가
	if msg.String() == "a" {
		m.feedback = ""
		m.editingID = ""
		m.formStep = monthlyFormStepItem
		m.formItem = ""
		m.formMoney = ""
		m.formLAccount = ""
		m.formLAccountID = ""
		m.formRAccount = ""
		m.formRAccountID = ""
		m.formPayDate = ""
		m.formSkipHoliday = ""
		m.textInput = ""
		m.mode = monthlyModeAdd
		return m, nil
	}
	var cmd tea.Cmd
	m.monthlyList, cmd = m.monthlyList.Update(msg)
	return m, cmd
}

func (m *monthlySubModel) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch ConfirmAction(msg) {
	case ActionConfirm:
		if len(m.items) > 0 {
			item := m.items[m.monthlyList.Index()]
			m.mode = monthlyModeLoading
			return m, m.doDelete(m.activeSlot, item.ID)
		}
	case ActionBack:
		m.mode = monthlyModeList
	}
	return m, nil
}

func (m *monthlySubModel) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// raw Key handling to match account_manage_sub.go exactly (FormAction would steal 'q' as Exit for text input)
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit // unreachable (GlobalAction catches ctrl+c in Update before handleKey); kept for exact structural match to account_manage_sub.go handleFormKey
	case tea.KeyEscape:
		// full reset (editingID + forms + feedback) on cancel; esc only (not q) so 'q' is typeable rune
		m.editingID = ""
		m.formStep = monthlyFormStepItem
		m.formItem = ""
		m.formMoney = ""
		m.formLAccount = ""
		m.formLAccountID = ""
		m.formRAccount = ""
		m.formRAccountID = ""
		m.formPayDate = ""
		m.formSkipHoliday = ""
		m.feedback = ""
		m.textInput = ""
		m.mode = monthlyModeList
		return m, nil
	case tea.KeyEnter:
		return m.advanceFormStep()
	case tea.KeyBackspace, tea.KeyDelete:
		if len(m.textInput) > 0 {
			runes := []rune(m.textInput)
			m.textInput = string(runes[:len(runes)-1])
		}
	case tea.KeyRunes:
		m.textInput += string(msg.Runes)
	}
	return m, nil
}

func (m *monthlySubModel) advanceFormStep() (tea.Model, tea.Cmd) {
	trim := strings.TrimSpace(m.textInput)
	switch m.formStep {
	case monthlyFormStepItem:
		if trim == "" {
			return m, nil
		}
		m.formItem = trim
		m.formStep = monthlyFormStepMoney
		m.textInput = m.formMoney
	case monthlyFormStepMoney:
		if trim != "" {
			if _, err := strconv.ParseInt(trim, 10, 64); err != nil {
				return m, nil // stay in step, allow correction (no silent 0)
			}
		}
		m.formMoney = trim
		m.formStep = monthlyFormStepLAccount
		m.textInput = m.formLAccount
	case monthlyFormStepLAccount:
		m.formLAccount = trim
		m.formStep = monthlyFormStepLAccountID
		m.textInput = m.formLAccountID
	case monthlyFormStepLAccountID:
		m.formLAccountID = trim
		m.formStep = monthlyFormStepRAccount
		m.textInput = m.formRAccount
	case monthlyFormStepRAccount:
		m.formRAccount = trim
		m.formStep = monthlyFormStepRAccountID
		m.textInput = m.formRAccountID
	case monthlyFormStepRAccountID:
		m.formRAccountID = trim
		m.formStep = monthlyFormStepPayDate
		m.textInput = m.formPayDate
	case monthlyFormStepPayDate:
		if trim != "" {
			if pv, err := strconv.Atoi(trim); err != nil || pv < 1 || pv > 31 {
				return m, nil // stay in step (reject bad; submit will also validate)
			}
		}
		m.formPayDate = trim
		m.formStep = monthlyFormStepSkipHoliday
		m.textInput = m.formSkipHoliday
	case monthlyFormStepSkipHoliday:
		if trim == "" {
			trim = "none"
		}
		m.formSkipHoliday = trim
		m.formStep = monthlyFormStepConfirm
		m.textInput = ""
	case monthlyFormStepConfirm:
		return m, m.submitForm()
	}
	return m, nil
}

func (m *monthlySubModel) submitForm() tea.Cmd {
	sectionID := m.cfg.SectionID
	slot := m.activeSlot
	input := api.MonthlyItemInput{
		FrequentItemInput: api.FrequentItemInput{
			Item:       strings.TrimSpace(m.formItem),
			LAccount:   strings.TrimSpace(m.formLAccount),
			LAccountID: strings.TrimSpace(m.formLAccountID),
			RAccount:   strings.TrimSpace(m.formRAccount),
			RAccountID: strings.TrimSpace(m.formRAccountID),
		},
	}
	// editingID guard/clear hoisted before parses (addresses lingering ID on early error returns for edit)
	editID := ""
	if m.mode == monthlyModeAdd {
		m.editingID = ""
	} else {
		if m.editingID == "" {
			return func() tea.Msg { return monthlyActionDoneMsg{err: fmt.Errorf("수정 ID가 없습니다")} }
		}
		editID = m.editingID
		m.editingID = ""
	}
	if ms := strings.TrimSpace(m.formMoney); ms != "" {
		mv, err := strconv.ParseInt(ms, 10, 64)
		if err != nil {
			return func() tea.Msg {
				return monthlyActionDoneMsg{err: fmt.Errorf("금액: 잘못된 숫자 입력 (%s)", ms)}
			}
		}
		if mv < 0 {
			mv = 0
		}
		input.Money = mv
	}
	if ps := strings.TrimSpace(m.formPayDate); ps != "" {
		pv, err := strconv.Atoi(ps)
		if err != nil {
			return func() tea.Msg {
				return monthlyActionDoneMsg{err: fmt.Errorf("결제일: 잘못된 숫자 입력 (%s)", ps)}
			}
		}
		if pv < 1 {
			pv = 1
		}
		if pv > 31 {
			pv = 31
		}
		input.PayDate = pv
	}
	input.SkipHoliday = strings.TrimSpace(m.formSkipHoliday)
	if input.SkipHoliday == "" {
		input.SkipHoliday = "none"
	}
	if m.mode == monthlyModeAdd {
		return func() tea.Msg {
			_, err := m.client.CreateMonthlyItem(sectionID, slot, input)
			if err != nil {
				return monthlyActionDoneMsg{err: err}
			}
			return monthlyActionDoneMsg{feedback: "추가되었습니다"}
		}
	}
	// edit path
	return func() tea.Msg {
		_, err := m.client.UpdateMonthlyItem(sectionID, slot, editID, input)
		if err != nil {
			return monthlyActionDoneMsg{err: err}
		}
		return monthlyActionDoneMsg{feedback: "수정되었습니다"}
	}
}

func (m *monthlySubModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("월별입력") + "\n\n")

	switch m.mode {
	case monthlyModeSlotSelect:
		b.WriteString(headerStyle.Render("슬롯 선택") + "\n\n")
		for i, s := range m.slots {
			if i == m.slotCursor {
				b.WriteString(selectedStyle.Render("> "+s) + "\n")
			} else {
				b.WriteString("  " + s + "\n")
			}
		}
		b.WriteString("\n" + helpStyle.Render("[←/→/h/l] 이동  [Enter] 선택  [Esc/q] 뒤로"))

	case monthlyModeLoading:
		b.WriteString("불러오는 중...")

	case monthlyModeList:
		if m.feedback != "" {
			b.WriteString(successStyle.Render(m.feedback) + "\n\n")
		}
		b.WriteString(headerStyle.Render(fmt.Sprintf("월별입력 목록 [%s]", m.activeSlot)) + "\n\n")
		if len(m.items) == 0 {
			b.WriteString("  항목이 없습니다\n")
		} else {
			b.WriteString(m.monthlyList.View() + "\n")
		}
		b.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [a] 추가  [e] 수정  [d] 삭제  [Esc] 슬롯선택  [q] 메뉴"))

	case monthlyModeAdd:
		b.WriteString(headerStyle.Render(fmt.Sprintf("월별입력 추가 [%s]", m.activeSlot)) + "\n\n")
		b.WriteString(m.renderFormStep("추가"))

	case monthlyModeEdit:
		b.WriteString(headerStyle.Render(fmt.Sprintf("월별입력 수정 [%s]", m.activeSlot)) + "\n\n")
		b.WriteString(m.renderFormStep("수정"))

	case monthlyModeConfirmDelete:
		if len(m.items) > 0 {
			item := m.items[m.monthlyList.Index()]
			b.WriteString(headerStyle.Render("삭제 확인") + "\n\n")
			b.WriteString(fmt.Sprintf("  '%s' 항목을 삭제합니까?\n", item.Item))
			b.WriteString("\n" + helpStyle.Render("[y] 삭제  [n/Esc] 취소"))
		}

	case monthlyModeError:
		b.WriteString(errorStyle.Render("[오류] "+m.errMsg) + "\n\n")
		b.WriteString(helpStyle.Render("[아무 키] 메뉴로 돌아가기"))
	}

	return b.String()
}

// formatDDayBadge는 D-Day 뱃지 문자열 반환
func formatDDayBadge(dday int, paidDate string) string {
	if paidDate != "" {
		return "[납부완료]"
	}
	if dday == 0 {
		return "[D-Day]"
	}
	if dday < 0 {
		return fmt.Sprintf("[D+%d]", -dday)
	}
	return fmt.Sprintf("[D-%d]", dday)
}

// calcDDay는 결제일로부터 오늘까지 D-Day 계산
func calcDDay(dueDateStr string) int {
	if dueDateStr == "" {
		return 0
	}
	due, err := time.Parse("20060102", dueDateStr)
	if err != nil {
		return 0
	}
	today := time.Now().Truncate(24 * time.Hour)
	diff := int(due.Sub(today).Hours() / 24)
	return diff
}

// parseMonthlyItemsFromRaw는 raw API 응답에서 월별입력 항목 목록 파싱
func parseMonthlyItemsFromRaw(raw []byte) []monthlyItem {
	var wrapper map[string]interface{}
	if err := parseJSONResponse(raw, &wrapper); err != nil {
		return nil
	}

	var root map[string]interface{}
	if r, ok := wrapper["results"]; ok {
		if rm, ok := r.(map[string]interface{}); ok {
			root = rm
		}
	}
	if root == nil {
		root = wrapper
	}

	var items []monthlyItem
	for _, v := range root {
		arr, ok := v.([]interface{})
		if !ok {
			continue
		}
		for _, elem := range arr {
			m, ok := elem.(map[string]interface{})
			if !ok {
				continue
			}
			mi := monthlyItem{}
			mi.ID, _ = m["monthly_item_id"].(string)
			mi.Item, _ = m["item"].(string)
			mi.LAccount, _ = m["l_account"].(string)
			mi.LAccountID, _ = m["l_account_id"].(string)
			mi.RAccount, _ = m["r_account"].(string)
			mi.RAccountID, _ = m["r_account_id"].(string)
			mi.SkipHoliday, _ = m["skip_holiday"].(string)
			mi.DueDate, _ = m["due_date"].(string)
			mi.PaidDate, _ = m["paid_date"].(string)
			if mv, ok := m["money"].(float64); ok {
				mi.Money = int64(mv)
			}
			if pd, ok := m["pay_date"].(float64); ok {
				mi.PayDate = int(pd)
			}
			if dd, ok := m["d_day"].(float64); ok {
				mi.DDay = int(dd)
			} else {
				mi.DDay = calcDDay(mi.DueDate)
			}
			items = append(items, mi)
		}
	}
	return items
}

// renderFormStep은 add/edit 폼의 현재 단계 렌더링 (account_manage_sub 패턴)
func (m *monthlySubModel) renderFormStep(action string) string {
	var b strings.Builder
	switch m.formStep {
	case monthlyFormStepItem:
		b.WriteString("아이템: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case monthlyFormStepMoney:
		b.WriteString(fmt.Sprintf("아이템: %s\n", m.formItem))
		b.WriteString("금액: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case monthlyFormStepLAccount:
		b.WriteString(fmt.Sprintf("아이템: %s  금액: %s\n", m.formItem, m.formMoney))
		b.WriteString("왼쪽: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case monthlyFormStepLAccountID:
		b.WriteString(fmt.Sprintf("아이템: %s  금액: %s\n왼쪽: %s\n", m.formItem, m.formMoney, m.formLAccount))
		b.WriteString("왼쪽 ID: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case monthlyFormStepRAccount:
		b.WriteString(fmt.Sprintf("아이템: %s  금액: %s\n왼쪽: %s (%s)\n", m.formItem, m.formMoney, m.formLAccount, m.formLAccountID))
		b.WriteString("오른쪽: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case monthlyFormStepRAccountID:
		b.WriteString(fmt.Sprintf("아이템: %s  금액: %s\n왼쪽: %s (%s)\n오른쪽: %s\n", m.formItem, m.formMoney, m.formLAccount, m.formLAccountID, m.formRAccount))
		b.WriteString("오른쪽 ID: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case monthlyFormStepPayDate:
		b.WriteString(fmt.Sprintf("아이템: %s  금액: %s\n왼쪽: %s (%s)\n오른쪽: %s (%s)\n", m.formItem, m.formMoney, m.formLAccount, m.formLAccountID, m.formRAccount, m.formRAccountID))
		b.WriteString("결제일 (1-31): " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case monthlyFormStepSkipHoliday:
		b.WriteString(fmt.Sprintf("아이템: %s  금액: %s\n왼쪽: %s (%s)\n오른쪽: %s (%s)\n결제일: %s\n", m.formItem, m.formMoney, m.formLAccount, m.formLAccountID, m.formRAccount, m.formRAccountID, m.formPayDate))
		b.WriteString("휴일처리 (before/after/none): " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case monthlyFormStepConfirm:
		mny := int64(0)
		if ms := strings.TrimSpace(m.formMoney); ms != "" {
			if v, err := strconv.ParseInt(ms, 10, 64); err == nil {
				mny = v
			}
		}
		pd := 0
		if ps := strings.TrimSpace(m.formPayDate); ps != "" {
			if v, err := strconv.Atoi(ps); err == nil {
				pd = v
			}
		}
		b.WriteString(fmt.Sprintf("아이템: %s\n금액: %s원\n왼쪽: %s (%s)\n오른쪽: %s (%s)\n결제일: %d\n휴일처리: %s\n\n",
			m.formItem, FormatMoney(float64(mny)), m.formLAccount, m.formLAccountID, m.formRAccount, m.formRAccountID, pd, m.formSkipHoliday))
		b.WriteString(helpStyle.Render(fmt.Sprintf("[Enter] %s 실행  [Esc] 취소", action)) + "\n")
	}
	return b.String()
}
