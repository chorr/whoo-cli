// cmd/frequent_sub.go
// 자주입력 관리 - bubbletea 서브 모델

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

type frequentMode int

const (
	frequentModeSlotSelect    frequentMode = iota // 슬롯 선택
	frequentModeList                              // 항목 목록
	frequentModeAdd                               // 추가 폼 (list 'a')
	frequentModeEdit                              // 수정 폼 (list 'e')
	frequentModeConfirmUse                        // 거래 생성 확인
	frequentModeConfirmDelete                     // 삭제 확인
	frequentModeLoading
	frequentModeError
)

const (
	frequentFormStepItem = iota
	frequentFormStepMoney
	frequentFormStepLAccount
	frequentFormStepLAccountID
	frequentFormStepRAccount
	frequentFormStepRAccountID
	frequentFormStepConfirm
)

// frequentItem은 TUI에서 표시하는 자주입력 항목
type frequentItem struct {
	ID         string
	Item       string
	Money      int64
	LAccount   string
	LAccountID string
	RAccount   string
	RAccountID string
}

// frequentListItem은 bubbles/list용 항목 래퍼
type frequentListItem struct {
	item frequentItem
}

func (i frequentListItem) Title() string {
	return fmt.Sprintf("%-20s  %s원", i.item.Item, FormatMoney(float64(i.item.Money)))
}
func (i frequentListItem) Description() string { return "" }
func (i frequentListItem) FilterValue() string { return i.item.Item }

type frequentSubModel struct {
	cfg    *config.Config
	client *api.WhooingClient
	mode   frequentMode
	errMsg string

	// 슬롯 탭 (가로 탐색, 3개 고정 → 수동 커서)
	slotCursor int
	slots      []string
	activeSlot string

	// 항목 목록 (bubbles/list)
	items    []frequentItem // 원본 데이터 (선택 시 참조)
	freqList list.Model

	// 결과 피드백
	feedback string

	// 폼 입력 (add/edit)
	formStep       int
	editingID      string
	formItem       string
	formMoney      string
	formLAccount   string
	formLAccountID string
	formRAccount   string
	formRAccountID string
	textInput      string
}

func newFrequentSubModel(cfg *config.Config) *frequentSubModel {
	return &frequentSubModel{
		cfg:    cfg,
		client: NewClient(cfg),
		mode:   frequentModeSlotSelect,
		slots:  []string{"slot1", "slot2", "slot3"},
	}
}

// ─── 비동기 커맨드 ──────────────────────────────────────────────

type frequentItemsLoadedMsg struct {
	slot  string
	items []frequentItem
	err   error
}

type frequentActionDoneMsg struct {
	feedback string
	err      error
}

func (m *frequentSubModel) loadItems(slot string) tea.Cmd {
	return func() tea.Msg {
		raw, err := m.client.GetFrequentItemsSlot(m.cfg.SectionID, slot)
		if err != nil {
			return frequentItemsLoadedMsg{slot: slot, err: err}
		}
		items := parseFrequentItemsFromRaw(raw)
		return frequentItemsLoadedMsg{slot: slot, items: items}
	}
}

func (m *frequentSubModel) doUse(item frequentItem) tea.Cmd {
	return func() tea.Msg {
		today := time.Now().Format("20060102")
		entry, err := m.client.CreateEntry(
			m.cfg.SectionID, today,
			item.LAccount, item.LAccountID,
			item.RAccount, item.RAccountID,
			item.Item, "", float64(item.Money),
		)
		if err != nil {
			return frequentActionDoneMsg{err: err}
		}
		return frequentActionDoneMsg{
			feedback: fmt.Sprintf("거래 생성 완료: %s %s원", entry.Item, FormatMoney(entry.Money)),
		}
	}
}

func (m *frequentSubModel) doDelete(slot, itemID string) tea.Cmd {
	return func() tea.Msg {
		_, err := m.client.DeleteFrequentItem(m.cfg.SectionID, slot, itemID)
		if err != nil {
			return frequentActionDoneMsg{err: err}
		}
		return frequentActionDoneMsg{feedback: "삭제되었습니다"}
	}
}

func (m *frequentSubModel) buildFreqList(items []frequentItem) list.Model {
	listItems := make([]list.Item, len(items))
	for i, it := range items {
		listItems[i] = frequentListItem{item: it}
	}
	return newPlainList(listItems, 60, len(listItems)+2)
}

// ─── BubbleTea ────────────────────────────────────────────────

func (m *frequentSubModel) Init() tea.Cmd {
	return nil
}

func (m *frequentSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case frequentItemsLoadedMsg:
		if msg.err != nil {
			m.mode = frequentModeError
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.activeSlot = msg.slot
		m.items = msg.items
		m.freqList = m.buildFreqList(msg.items)
		m.mode = frequentModeList
		return m, nil

	case frequentActionDoneMsg:
		if msg.err != nil {
			m.mode = frequentModeError
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.feedback = msg.feedback
		m.mode = frequentModeLoading
		return m, m.loadItems(m.activeSlot)

	case tea.KeyMsg:
		if GlobalAction(msg) == ActionQuit {
			return m, tea.Quit
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *frequentSubModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case frequentModeSlotSelect:
		return m.handleSlotKey(msg)
	case frequentModeList:
		return m.handleListKey(msg)
	case frequentModeConfirmUse, frequentModeConfirmDelete:
		return m.handleConfirmKey(msg)
	case frequentModeAdd, frequentModeEdit:
		return m.handleFormKey(msg)
	case frequentModeError:
		switch ErrorAction(msg) {
		case ActionBack, ActionExit:
			return m, func() tea.Msg { return backToMenuMsg{} }
		}
	}
	return m, nil
}

func (m *frequentSubModel) handleSlotKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch HorizontalSelectAction(msg) {
	case ActionBack, ActionExit:
		return m, func() tea.Msg { return backToMenuMsg{} }
	case ActionConfirm:
		slot := m.slots[m.slotCursor]
		m.activeSlot = slot
		m.mode = frequentModeLoading
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

func (m *frequentSubModel) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch ListAction(msg) {
	case ActionBack:
		m.mode = frequentModeSlotSelect
		m.feedback = ""
		return m, nil
	case ActionConfirm:
		if len(m.items) > 0 {
			m.mode = frequentModeConfirmUse
		}
		return m, nil
	case ActionEdit:
		if len(m.items) > 0 {
			idx := m.freqList.Index()
			it := m.items[idx]
			m.feedback = ""
			m.mode = frequentModeEdit
			m.editingID = it.ID
			m.formStep = frequentFormStepItem
			m.formItem = it.Item
			m.formMoney = fmt.Sprintf("%d", it.Money)
			m.formLAccount = it.LAccount
			m.formLAccountID = it.LAccountID
			m.formRAccount = it.RAccount
			m.formRAccountID = it.RAccountID
			m.textInput = m.formItem
		}
		return m, nil
	case ActionDelete:
		if len(m.items) > 0 {
			m.mode = frequentModeConfirmDelete
		}
		return m, nil
	case ActionExit:
		return m, func() tea.Msg { return backToMenuMsg{} }
	}
	// 도메인 전용: a = 추가
	if msg.String() == "a" {
		m.feedback = ""
		m.editingID = ""
		m.formStep = frequentFormStepItem
		m.formItem = ""
		m.formMoney = ""
		m.formLAccount = ""
		m.formLAccountID = ""
		m.formRAccount = ""
		m.formRAccountID = ""
		m.textInput = ""
		m.mode = frequentModeAdd
		return m, nil
	}
	var cmd tea.Cmd
	m.freqList, cmd = m.freqList.Update(msg)
	return m, cmd
}

func (m *frequentSubModel) handleConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch ConfirmAction(msg) {
	case ActionConfirm:
		if m.mode == frequentModeConfirmUse {
			item := m.items[m.freqList.Index()]
			m.mode = frequentModeLoading
			return m, m.doUse(item)
		}
		if m.mode == frequentModeConfirmDelete {
			item := m.items[m.freqList.Index()]
			m.mode = frequentModeLoading
			return m, m.doDelete(m.activeSlot, item.ID)
		}
	case ActionBack:
		m.mode = frequentModeList
	}
	return m, nil
}

func (m *frequentSubModel) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// raw Key handling to match account_manage_sub.go exactly (FormAction would steal 'q' as Exit for text input)
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit // unreachable (GlobalAction catches ctrl+c in Update before handleKey); kept for exact structural match to account_manage_sub.go handleFormKey
	case tea.KeyEscape:
		// full reset (editingID + forms + feedback) on cancel; esc only (not q) so 'q' is typeable rune
		m.editingID = ""
		m.formStep = frequentFormStepItem
		m.formItem = ""
		m.formMoney = ""
		m.formLAccount = ""
		m.formLAccountID = ""
		m.formRAccount = ""
		m.formRAccountID = ""
		m.feedback = ""
		m.textInput = ""
		m.mode = frequentModeList
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

func (m *frequentSubModel) advanceFormStep() (tea.Model, tea.Cmd) {
	trim := strings.TrimSpace(m.textInput)
	switch m.formStep {
	case frequentFormStepItem:
		if trim == "" {
			return m, nil
		}
		m.formItem = trim
		m.formStep = frequentFormStepMoney
		m.textInput = m.formMoney
	case frequentFormStepMoney:
		if trim != "" {
			if _, err := strconv.ParseInt(trim, 10, 64); err != nil {
				return m, nil // stay in step, allow correction (no silent 0)
			}
		}
		m.formMoney = trim
		m.formStep = frequentFormStepLAccount
		m.textInput = m.formLAccount
	case frequentFormStepLAccount:
		m.formLAccount = trim
		m.formStep = frequentFormStepLAccountID
		m.textInput = m.formLAccountID
	case frequentFormStepLAccountID:
		m.formLAccountID = trim
		m.formStep = frequentFormStepRAccount
		m.textInput = m.formRAccount
	case frequentFormStepRAccount:
		m.formRAccount = trim
		m.formStep = frequentFormStepRAccountID
		m.textInput = m.formRAccountID
	case frequentFormStepRAccountID:
		m.formRAccountID = trim
		m.formStep = frequentFormStepConfirm
		m.textInput = ""
	case frequentFormStepConfirm:
		return m, m.submitForm()
	}
	return m, nil
}

func (m *frequentSubModel) submitForm() tea.Cmd {
	sectionID := m.cfg.SectionID
	slot := m.activeSlot
	input := api.FrequentItemInput{
		Item:       strings.TrimSpace(m.formItem),
		LAccount:   strings.TrimSpace(m.formLAccount),
		LAccountID: strings.TrimSpace(m.formLAccountID),
		RAccount:   strings.TrimSpace(m.formRAccount),
		RAccountID: strings.TrimSpace(m.formRAccountID),
	}
	if ms := strings.TrimSpace(m.formMoney); ms != "" {
		mv, err := strconv.ParseInt(ms, 10, 64)
		if err != nil {
			return func() tea.Msg {
				return frequentActionDoneMsg{err: fmt.Errorf("금액: 잘못된 숫자 입력 (%s)", ms)}
			}
		}
		if mv < 0 {
			mv = 0
		}
		input.Money = mv
	}
	// editingID guard/clear hoisted before any error return (for edit)
	editID := ""
	if m.mode == frequentModeAdd {
		m.editingID = ""
	} else {
		if m.editingID == "" {
			return func() tea.Msg { return frequentActionDoneMsg{err: fmt.Errorf("수정 ID가 없습니다")} }
		}
		editID = m.editingID
		m.editingID = ""
	}
	if m.mode == frequentModeAdd {
		return func() tea.Msg {
			_, err := m.client.CreateFrequentItem(sectionID, slot, input)
			if err != nil {
				return frequentActionDoneMsg{err: err}
			}
			return frequentActionDoneMsg{feedback: "추가되었습니다"}
		}
	}
	// edit path
	return func() tea.Msg {
		_, err := m.client.UpdateFrequentItem(sectionID, slot, editID, input)
		if err != nil {
			return frequentActionDoneMsg{err: err}
		}
		return frequentActionDoneMsg{feedback: "수정되었습니다"}
	}
}

func (m *frequentSubModel) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("자주입력") + "\n\n")

	switch m.mode {
	case frequentModeSlotSelect:
		b.WriteString(headerStyle.Render("슬롯 선택") + "\n\n")
		for i, s := range m.slots {
			if i == m.slotCursor {
				b.WriteString(selectedStyle.Render("> "+s) + "\n")
			} else {
				b.WriteString("  " + s + "\n")
			}
		}
		b.WriteString("\n" + helpStyle.Render("[←/→/h/l] 이동  [Enter] 선택  [Esc/q] 뒤로"))

	case frequentModeLoading:
		b.WriteString("불러오는 중...")

	case frequentModeList:
		if m.feedback != "" {
			b.WriteString(successStyle.Render(m.feedback) + "\n\n")
		}
		b.WriteString(headerStyle.Render(fmt.Sprintf("자주입력 목록 [%s]", m.activeSlot)) + "\n\n")
		if len(m.items) == 0 {
			b.WriteString("  항목이 없습니다\n")
		} else {
			b.WriteString(m.freqList.View() + "\n")
		}
		b.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [Enter] 거래생성  [a] 추가  [e] 수정  [d] 삭제  [Esc] 슬롯선택  [q] 메뉴")) // note: Enter for use; a/e order matches monthly list help

	case frequentModeAdd:
		b.WriteString(headerStyle.Render(fmt.Sprintf("자주입력 추가 [%s]", m.activeSlot)) + "\n\n")
		b.WriteString(m.renderFormStep("추가"))

	case frequentModeEdit:
		b.WriteString(headerStyle.Render(fmt.Sprintf("자주입력 수정 [%s]", m.activeSlot)) + "\n\n")
		b.WriteString(m.renderFormStep("수정"))

	case frequentModeConfirmUse:
		if len(m.items) > 0 {
			item := m.items[m.freqList.Index()]
			b.WriteString(headerStyle.Render("거래 생성 확인") + "\n\n")
			b.WriteString(fmt.Sprintf("  아이템 : %s\n", item.Item))
			b.WriteString(fmt.Sprintf("  금액   : %s원\n", FormatMoney(float64(item.Money))))
			b.WriteString(fmt.Sprintf("  날짜   : %s (오늘)\n", time.Now().Format("2006-01-02")))
			b.WriteString("\n" + helpStyle.Render("[y] 확인  [n/Esc] 취소"))
		}

	case frequentModeConfirmDelete:
		if len(m.items) > 0 {
			item := m.items[m.freqList.Index()]
			b.WriteString(headerStyle.Render("삭제 확인") + "\n\n")
			b.WriteString(fmt.Sprintf("  '%s' 항목을 삭제합니까?\n", item.Item))
			b.WriteString("\n" + helpStyle.Render("[y] 삭제  [n/Esc] 취소"))
		}

	case frequentModeError:
		b.WriteString(errorStyle.Render("[오류] "+m.errMsg) + "\n\n")
		b.WriteString(helpStyle.Render("[아무 키] 메뉴로 돌아가기"))
	}

	return b.String()
}

// parseFrequentItemsFromRaw는 raw API 응답에서 자주입력 항목 목록 파싱
func parseFrequentItemsFromRaw(raw []byte) []frequentItem {
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

	var items []frequentItem
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
			fi := frequentItem{}
			fi.ID, _ = m["frequent_item_id"].(string)
			fi.Item, _ = m["item"].(string)
			fi.LAccount, _ = m["l_account"].(string)
			fi.LAccountID, _ = m["l_account_id"].(string)
			fi.RAccount, _ = m["r_account"].(string)
			fi.RAccountID, _ = m["r_account_id"].(string)
			if mv, ok := m["money"].(float64); ok {
				fi.Money = int64(mv)
			}
			items = append(items, fi)
		}
	}
	return items
}

// renderFormStep은 add/edit 폼의 현재 단계 렌더링 (account_manage_sub 패턴)
func (m *frequentSubModel) renderFormStep(action string) string {
	var b strings.Builder
	switch m.formStep {
	case frequentFormStepItem:
		b.WriteString("아이템: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case frequentFormStepMoney:
		b.WriteString(fmt.Sprintf("아이템: %s\n", m.formItem))
		b.WriteString("금액: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case frequentFormStepLAccount:
		b.WriteString(fmt.Sprintf("아이템: %s  금액: %s\n", m.formItem, m.formMoney))
		b.WriteString("왼쪽: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case frequentFormStepLAccountID:
		b.WriteString(fmt.Sprintf("아이템: %s  금액: %s\n왼쪽: %s\n", m.formItem, m.formMoney, m.formLAccount))
		b.WriteString("왼쪽 ID: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case frequentFormStepRAccount:
		b.WriteString(fmt.Sprintf("아이템: %s  금액: %s\n왼쪽: %s (%s)\n", m.formItem, m.formMoney, m.formLAccount, m.formLAccountID))
		b.WriteString("오른쪽: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case frequentFormStepRAccountID:
		b.WriteString(fmt.Sprintf("아이템: %s  금액: %s\n왼쪽: %s (%s)\n오른쪽: %s\n", m.formItem, m.formMoney, m.formLAccount, m.formLAccountID, m.formRAccount))
		b.WriteString("오른쪽 ID: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case frequentFormStepConfirm:
		mny := int64(0)
		if ms := strings.TrimSpace(m.formMoney); ms != "" {
			if v, err := strconv.ParseInt(ms, 10, 64); err == nil {
				mny = v
			}
		}
		b.WriteString(fmt.Sprintf("아이템: %s\n금액: %s원\n왼쪽: %s (%s)\n오른쪽: %s (%s)\n\n",
			m.formItem, FormatMoney(float64(mny)), m.formLAccount, m.formLAccountID, m.formRAccount, m.formRAccountID))
		b.WriteString(helpStyle.Render(fmt.Sprintf("[Enter] %s 실행  [Esc] 취소", action)) + "\n")
	}
	return b.String()
}
