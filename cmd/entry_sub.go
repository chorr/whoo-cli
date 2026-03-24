// cmd/entry_sub.go
// 거래 입력/수정 - bubbletea 서브 모델

package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"whooing-cli/api"
	"whooing-cli/config"
)

// entrySubModel은 거래 입력 서브 모델
type entrySubModel struct {
	cfg         *config.Config
	client      *api.WhooingClient
	accountsMap *api.AccountsMap
	step        int
	err         error

	// 입력값
	date         string
	lAccountType string
	lAccountID   string
	lAccountName string
	rAccountType string
	rAccountID   string
	rAccountName string
	money        string
	item         string
	memo         string

	// 수정 모드
	editMode  bool
	editEntry *api.Entry

	// UI 상태
	textInput string
	cursor    int
	choices   []accountChoice
	done      bool
	cancelled bool
}

// accountChoice는 계정 선택지
type accountChoice struct {
	id    string
	title string
}

const (
	entryStepDate = iota
	entryStepLAccountType
	entryStepLAccountID
	entryStepRAccountType
	entryStepRAccountID
	entryStepMoney
	entryStepItem
	entryStepMemo
	entryStepConfirm
	entryStepDone
)

// newEntrySubModel은 새로운 거래 입력 모델을 생성
func newEntrySubModel(cfg *config.Config) *entrySubModel {
	return &entrySubModel{
		cfg:    cfg,
		client: NewClient(cfg),
		step:   entryStepDate,
		date:   time.Now().Format("20060102"),
	}
}

// newEntrySubModelForEdit은 수정 모드로 거래 입력 모델을 생성
func newEntrySubModelForEdit(cfg *config.Config, entry api.Entry, accountsMap *api.AccountsMap) *entrySubModel {
	return &entrySubModel{
		cfg:          cfg,
		client:       api.NewWhooingClient(cfg),
		accountsMap:  accountsMap,
		step:         entryStepDate,
		editMode:     true,
		editEntry:    &entry,
		date:         entry.DateOnly(),
		lAccountType: entry.LAccount,
		lAccountID:   entry.LAccountID,
		lAccountName: accountsMap.GetTitle(entry.LAccount, entry.LAccountID),
		rAccountType: entry.RAccount,
		rAccountID:   entry.RAccountID,
		rAccountName: accountsMap.GetTitle(entry.RAccount, entry.RAccountID),
		money:        fmt.Sprintf("%.0f", entry.Money),
		item:         entry.Item,
		memo:         entry.Memo,
		textInput:    entry.DateOnly(),
	}
}

// editEntryMsg는 거래 수정 요청 메시지
type editEntryMsg struct {
	entry       api.Entry
	accountsMap *api.AccountsMap
}

// entryUpdatedMsg는 거래 수정 완료 메시지
type entryUpdatedMsg struct {
	entry *api.Entry
}

// accountsLoadedMsg는 계정 목록 로드 완료 메시지
type accountsLoadedMsg struct {
	accountsMap *api.AccountsMap
}

// entryCreatedMsg는 거래 생성 완료 메시지
type entryCreatedMsg struct {
	entry *api.Entry
}

// entryErrMsg는 거래 처리 중 오류 메시지
type entryErrMsg struct {
	err error
}

func (m *entrySubModel) Init() tea.Cmd {
	if m.accountsMap != nil {
		// 수정 모드: 이미 계정 맵 보유, 로딩 불필요
		m.textInput = m.date
		return nil
	}
	return m.loadAccounts()
}

func (m *entrySubModel) loadAccounts() tea.Cmd {
	return func() tea.Msg {
		accountsMap, err := m.client.GetAccountsMap(m.cfg.SectionID)
		if err != nil {
			return entryErrMsg{err: fmt.Errorf("계정 목록 조회 실패: %w", err)}
		}
		return accountsLoadedMsg{accountsMap: accountsMap}
	}
}

func (m *entrySubModel) createEntry() tea.Cmd {
	return func() tea.Msg {
		money, err := strconv.ParseFloat(m.money, 64)
		if err != nil {
			return entryErrMsg{err: fmt.Errorf("금액 파싱 오류: %w", err)}
		}
		entry, err := m.client.CreateEntry(
			m.cfg.SectionID,
			m.date,
			m.lAccountType, m.lAccountID,
			m.rAccountType, m.rAccountID,
			m.item, m.memo, money,
		)

		if err != nil {
			return entryErrMsg{err: fmt.Errorf("거래 등록 실패: %w", err)}
		}
		return entryCreatedMsg{entry: entry}
	}
}

func (m *entrySubModel) updateEntry() tea.Cmd {
	return func() tea.Msg {
		fields := m.buildChangedFields()
		if len(fields) == 0 {
			return entryUpdatedMsg{entry: m.editEntry}
		}
		entry, err := m.client.UpdateEntry(m.cfg.SectionID, m.editEntry.EntryID, fields)
		if err != nil {
			return entryErrMsg{err: fmt.Errorf("거래 수정 실패: %w", err)}
		}
		return entryUpdatedMsg{entry: entry}
	}
}

func (m *entrySubModel) buildChangedFields() map[string]string {
	fields := map[string]string{}
	orig := m.editEntry
	if m.date != orig.DateOnly() {
		fields["entry_date"] = m.date
	}
	if m.lAccountType != orig.LAccount {
		fields["l_account"] = m.lAccountType
	}
	if m.lAccountID != orig.LAccountID {
		fields["l_account_id"] = m.lAccountID
	}
	if m.rAccountType != orig.RAccount {
		fields["r_account"] = m.rAccountType
	}
	if m.rAccountID != orig.RAccountID {
		fields["r_account_id"] = m.rAccountID
	}
	money := fmt.Sprintf("%.0f", orig.Money)
	if m.money != money {
		fields["money"] = m.money
	}
	if m.item != orig.Item {
		fields["item"] = m.item
	}
	if m.memo != orig.Memo {
		fields["memo"] = m.memo
	}
	return fields
}

func (m *entrySubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case accountsLoadedMsg:
		m.accountsMap = msg.accountsMap
		m.textInput = m.date
		return m, nil

	case entryCreatedMsg:
		m.step = entryStepDone
		m.done = true
		return m, nil

	case entryUpdatedMsg:
		m.step = entryStepDone
		m.done = true
		return m, nil

	case entryErrMsg:
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.cancelled = true
			return m, func() tea.Msg { return backToMenuMsg{} }
		}
		// 에러 상태에서는 Enter/ESC로 메뉴로 돌아감
		if m.err != nil {
			if msg.String() == "enter" || msg.String() == "esc" {
				m.err = nil
				m.cancelled = true
				return m, func() tea.Msg { return backToMenuMsg{} }
			}
			return m, nil
		}
		if msg.String() == "esc" {
			return m, m.goBack()
		}

		switch m.step {
		case entryStepDate:
			return m.updateTextInput(msg, m.submitDate)
		case entryStepLAccountType:
			return m.updateTypeSelect(msg, func() {
				m.lAccountType = AccountTypes[m.cursor].Code
				m.step = entryStepLAccountID
				m.buildAccountChoicesForDate(m.lAccountType)
				m.cursor = 0
			})
		case entryStepLAccountID:
			return m.updateAccountSelect(msg, func() {
				m.lAccountID = m.choices[m.cursor].id
				m.lAccountName = m.choices[m.cursor].title
				m.step = entryStepRAccountType
				m.cursor = m.cursorForAccountType(m.rAccountType)
			})
		case entryStepRAccountType:
			return m.updateTypeSelect(msg, func() {
				m.rAccountType = AccountTypes[m.cursor].Code
				m.step = entryStepRAccountID
				m.buildAccountChoicesForDate(m.rAccountType)
				m.cursor = 0
			})
		case entryStepRAccountID:
			return m.updateAccountSelect(msg, func() {
				m.rAccountID = m.choices[m.cursor].id
				m.rAccountName = m.choices[m.cursor].title
				m.step = entryStepMoney
				m.textInput = m.money
				m.cursor = 0
			})
		case entryStepMoney:
			return m.updateTextInput(msg, m.submitMoney)
		case entryStepItem:
			return m.updateTextInput(msg, m.submitItem)
		case entryStepMemo:
			return m.updateTextInput(msg, m.submitMemo)
		case entryStepConfirm:
			return m.updateConfirm(msg)
		case entryStepDone:
			if msg.String() == "enter" {
				if m.editMode {
					return m, func() tea.Msg { return backToTransactionsMsg{} }
				}
				return m, func() tea.Msg { return backToMenuMsg{} }
			}
		}
	}
	return m, nil
}

func (m *entrySubModel) goBack() tea.Cmd {
	switch m.step {
	case entryStepDate:
		m.cancelled = true
		return func() tea.Msg { return backToMenuMsg{} }
	case entryStepLAccountType:
		m.step = entryStepDate
		m.textInput = m.date
		m.cursor = 0
	case entryStepLAccountID:
		m.step = entryStepLAccountType
		m.cursor = m.cursorForAccountType(m.lAccountType)
	case entryStepRAccountType:
		m.step = entryStepLAccountID
		m.buildAccountChoicesForDate(m.lAccountType)
		m.cursor = 0
	case entryStepRAccountID:
		m.step = entryStepRAccountType
		m.cursor = m.cursorForAccountType(m.rAccountType)
	case entryStepMoney:
		m.step = entryStepRAccountID
		m.buildAccountChoicesForDate(m.rAccountType)
		m.cursor = 0
	case entryStepItem:
		m.step = entryStepMoney
		m.textInput = m.money
	case entryStepMemo:
		m.step = entryStepItem
		m.textInput = m.item
	case entryStepConfirm:
		m.step = entryStepMemo
		m.textInput = m.memo
	}
	return nil
}

func (m *entrySubModel) updateTextInput(msg tea.KeyMsg, onSubmit func()) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		onSubmit()
	case "backspace":
		if len(m.textInput) > 0 {
			m.textInput = m.textInput[:len(m.textInput)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.textInput += msg.String()
		}
	}
	return m, nil
}

func (m *entrySubModel) updateTypeSelect(msg tea.KeyMsg, onSelect func()) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(AccountTypes)-1 {
			m.cursor++
		}
	case "enter":
		onSelect()
	case "1", "2", "3", "4", "5":
		idx, _ := strconv.Atoi(msg.String())
		idx--
		if idx >= 0 && idx < len(AccountTypes) {
			m.cursor = idx
			onSelect()
		}
	}
	return m, nil
}

func (m *entrySubModel) updateAccountSelect(msg tea.KeyMsg, onSelect func()) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.choices)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.choices) > 0 {
			onSelect()
		}
	default:
		if idx, err := strconv.Atoi(msg.String()); err == nil {
			idx--
			if idx >= 0 && idx < len(m.choices) {
				m.cursor = idx
				onSelect()
			}
		}
	}
	return m, nil
}

func (m *entrySubModel) updateConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y", "enter":
		// 확인 (Y가 기본값)
		if m.editMode {
			return m, m.updateEntry()
		}
		return m, m.createEntry()
	case "n", "N", "esc":
		// 취소
		m.cancelled = true
		return m, func() tea.Msg { return backToMenuMsg{} }
	}
	return m, nil
}

func (m *entrySubModel) submitDate() {
	input := strings.TrimSpace(m.textInput)
	if input == "" {
		input = time.Now().Format("20060102")
	}
	if len(input) != 8 {
		return
	}
	if _, err := time.Parse("20060102", input); err != nil {
		return
	}
	m.date = input
	m.step = entryStepLAccountType
	m.cursor = m.cursorForAccountType(m.lAccountType)
}

func (m *entrySubModel) submitMoney() {
	input := strings.TrimSpace(m.textInput)
	if input == "" {
		return
	}
	cleaned := strings.ReplaceAll(input, ",", "")
	if _, err := strconv.ParseFloat(cleaned, 64); err != nil {
		return
	}
	m.money = cleaned
	m.step = entryStepItem
	m.textInput = m.item
}

func (m *entrySubModel) submitItem() {
	input := strings.TrimSpace(m.textInput)
	if input == "" {
		return
	}
	m.item = input
	m.step = entryStepMemo
	m.textInput = m.memo
}

func (m *entrySubModel) submitMemo() {
	input := strings.TrimSpace(m.textInput)
	m.memo = input
	m.step = entryStepConfirm
}

func (m *entrySubModel) cursorForAccountType(code string) int {
	for i, at := range AccountTypes {
		if at.Code == code {
			return i
		}
	}
	return 0
}

// buildAccountChoicesForDate는 날짜 기준으로 유효한 계정만 필터링
func (m *entrySubModel) buildAccountChoicesForDate(accountType string) {
	m.choices = nil
	m.cursor = 0

	if m.accountsMap == nil {
		return
	}

	accounts := m.accountsMap.GetAccountsByType(accountType)
	if accounts == nil {
		return
	}

	ids := make([]string, 0, len(accounts))
	for id, detail := range accounts {
		if m.isAccountActiveOnDate(detail, m.date) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)

	for _, id := range ids {
		detail := accounts[id]
		m.choices = append(m.choices, accountChoice{
			id:    id,
			title: detail.Title,
		})
	}
}

// isAccountActiveOnDate는 계정이 특정 날짜에 활성 상태인지 확인
func (m *entrySubModel) isAccountActiveOnDate(detail api.AccountDetail, dateStr string) bool {
	openDate := normalizeDate(detail.OpenDate)
	closeDate := normalizeDate(detail.CloseDate)

	if openDate != "" && dateStr < openDate {
		return false
	}
	if closeDate != "" && dateStr > closeDate {
		return false
	}
	return true
}

// normalizeDate는 interface{} 타입의 날짜를 YYYYMMDD 문자열로 변환
func normalizeDate(v interface{}) string {
	if v == nil {
		return ""
	}
	switch d := v.(type) {
	case string:
		if d == "" || d == "0" {
			return ""
		}
		return d
	case float64:
		if d == 0 {
			return ""
		}
		return fmt.Sprintf("%.0f", d)
	}
	return ""
}

func (m *entrySubModel) View() string {
	if m.err != nil {
		return m.renderError()
	}

	if m.accountsMap == nil {
		label := "거래 입력"
		if m.editMode {
			label = "거래 수정"
		}
		return titleStyle.Render(label) + "\n\n계정 목록 로딩 중...\n"
	}

	if m.step == entryStepDone {
		return m.renderDone()
	}

	return m.renderForm()
}

func (m *entrySubModel) renderError() string {
	return titleStyle.Render("거래 입력") + "\n\n" +
		errorStyle.Render("[오류] "+m.err.Error()) + "\n\n" +
		helpStyle.Render("[Enter] 메뉴로 돌아가기")
}

func (m *entrySubModel) renderDone() string {
	money, _ := strconv.ParseFloat(m.money, 64)

	label := "거래가 등록되었습니다"
	if m.editMode {
		label = "거래가 수정되었습니다"
	}

	return titleStyle.Render("거래 입력") + "\n\n" +
		successStyle.Render("[완료] "+label) + "\n\n" +
		fmt.Sprintf("  %s | %s:%s <- %s:%s | %s | %s\n",
			FormatDate(m.date),
			FormatAccount(m.lAccountType), m.lAccountName,
			FormatAccount(m.rAccountType), m.rAccountName,
			FormatMoney(money),
			m.item,
		) + "\n" +
		helpStyle.Render("[Enter] 메뉴로 돌아가기")
}

func (m *entrySubModel) renderForm() string {
	formTitleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	formSelectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	var b strings.Builder

	if m.editMode {
		b.WriteString(formTitleStyle.Render("[거래 수정]"))
	} else {
		b.WriteString(formTitleStyle.Render("[거래 입력]"))
	}
	b.WriteString("\n\n")

	// 이미 입력된 값 표시
	if m.step > entryStepDate {
		b.WriteString(fmt.Sprintf("  날짜: %s\n", FormatDate(m.date)))
	}
	if m.step > entryStepLAccountID {
		b.WriteString(fmt.Sprintf("  왼쪽: %s > %s\n", FormatAccount(m.lAccountType), m.lAccountName))
	}
	if m.step > entryStepRAccountID {
		b.WriteString(fmt.Sprintf("  오른쪽: %s > %s\n", FormatAccount(m.rAccountType), m.rAccountName))
	}
	if m.step > entryStepMoney {
		money, _ := strconv.ParseFloat(m.money, 64)
		b.WriteString(fmt.Sprintf("  금액: %s\n", FormatMoney(money)))
	}
	if m.step > entryStepItem {
		b.WriteString(fmt.Sprintf("  아이템: %s\n", m.item))
	}
	if m.step > entryStepMemo {
		b.WriteString(fmt.Sprintf("  메모: %s\n", m.memo))
	}

	if m.step > entryStepDate {
		b.WriteString("\n")
	}

	// 현재 단계 표시
	switch m.step {
	case entryStepDate:
		b.WriteString(headerStyle.Render("날짜 (YYYYMMDD):"))
		b.WriteString(" ")
		b.WriteString(inputStyle.Render(m.textInput))
		b.WriteString(inputStyle.Render("_"))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("  Enter=확인 (빈 값=오늘)"))

	case entryStepLAccountType:
		b.WriteString(headerStyle.Render("왼쪽 계정 타입:"))
		b.WriteString("\n")
		m.renderTypeList(&b, formSelectedStyle, dimStyle)

	case entryStepLAccountID:
		b.WriteString(headerStyle.Render(fmt.Sprintf("왼쪽 %s 계정:", FormatAccount(m.lAccountType))))
		b.WriteString("\n")
		m.renderAccountList(&b, formSelectedStyle, dimStyle)

	case entryStepRAccountType:
		b.WriteString(headerStyle.Render("오른쪽 계정 타입:"))
		b.WriteString("\n")
		m.renderTypeList(&b, formSelectedStyle, dimStyle)

	case entryStepRAccountID:
		b.WriteString(headerStyle.Render(fmt.Sprintf("오른쪽 %s 계정:", FormatAccount(m.rAccountType))))
		b.WriteString("\n")
		m.renderAccountList(&b, formSelectedStyle, dimStyle)

	case entryStepMoney:
		b.WriteString(headerStyle.Render("금액:"))
		b.WriteString(" ")
		b.WriteString(inputStyle.Render(m.textInput))
		b.WriteString(inputStyle.Render("_"))

	case entryStepItem:
		b.WriteString(headerStyle.Render("아이템:"))
		b.WriteString(" ")
		b.WriteString(inputStyle.Render(m.textInput))
		b.WriteString(inputStyle.Render("_"))

	case entryStepMemo:
		b.WriteString(headerStyle.Render("메모 (선택):"))
		b.WriteString(" ")
		b.WriteString(inputStyle.Render(m.textInput))
		b.WriteString(inputStyle.Render("_"))

	case entryStepConfirm:
		money, _ := strconv.ParseFloat(m.money, 64)
		label := "[확인]"
		if m.editMode {
			label = "[수정 확인]"
		}
		b.WriteString(headerStyle.Render(label))
		b.WriteString(fmt.Sprintf(" %s | %s:%s <- %s:%s | %s | %s\n",
			FormatDate(m.date),
			FormatAccount(m.lAccountType), m.lAccountName,
			FormatAccount(m.rAccountType), m.rAccountName,
			FormatMoney(money),
			m.item,
		))
		if m.memo != "" {
			b.WriteString(fmt.Sprintf("  메모: %s\n", m.memo))
		}
		b.WriteString("\n")
		if m.editMode {
			b.WriteString("수정하시겠습니까? (Y/n): ")
		} else {
			b.WriteString("등록하시겠습니까? (Y/n): ")
		}
	}

	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("[esc] 이전  [ctrl+c] 취소"))
	b.WriteString("\n")

	return b.String()
}

func (m *entrySubModel) renderTypeList(b *strings.Builder, selectedStyle, dimStyle lipgloss.Style) {
	for i, at := range AccountTypes {
		line := fmt.Sprintf("%d. %s (%s)", i+1, at.Name, at.Code)
		if i == m.cursor {
			b.WriteString("  " + selectedStyle.Render("> "+line) + "\n")
		} else {
			b.WriteString("    " + line + "\n")
		}
	}
	b.WriteString(dimStyle.Render("  [↑/↓/j/k] 이동  [enter/숫자] 선택"))
}

func (m *entrySubModel) renderAccountList(b *strings.Builder, selectedStyle, dimStyle lipgloss.Style) {
	if len(m.choices) == 0 {
		b.WriteString("  (등록된 계정이 없습니다)\n")
		return
	}
	for i, c := range m.choices {
		line := fmt.Sprintf("%d. %s", i+1, c.title)
		if i == m.cursor {
			b.WriteString("  " + selectedStyle.Render("> "+line) + "\n")
		} else {
			b.WriteString("    " + line + "\n")
		}
	}
	b.WriteString(dimStyle.Render("  [↑/↓/j/k] 이동  [enter/숫자] 선택"))
}
