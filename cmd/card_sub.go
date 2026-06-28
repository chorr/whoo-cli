// cmd/card_sub.go
// 카드 관리 - bubbletea 서브 모델 (신용카드/체크카드 청구내역)

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"whoo-cli/api"
	"whoo-cli/config"
)

type cardMode int

const (
	cardModeTabSelect  cardMode = iota // 탭 선택 (신용/체크)
	cardModeCardList                   // 카드 목록
	cardModeMonthTable                 // 월별 청구 테이블
	cardModeDrilldown                  // 드릴다운: 해당 월 거래 목록
	cardModeLoading
	cardModeError
)

// cardTab은 탭 종류
type cardTab int

const (
	cardTabBill      cardTab = iota // 신용카드
	cardTabCheckcard                // 체크카드
)

// cardEntry는 카드 목록용 항목 정보
type cardEntry struct {
	AccountID    string
	Title        string
	Category     string // creditcard|checkcard
	PayDate      int    // 결제일 (1~31), 신용카드 전용
	PayAccountID string
	UrgentPay    bool // 결제일 임박 (3일 이내)
}

// cardMonthRow는 월별 청구 행
type cardMonthRow struct {
	YM    string // "202601"
	Money int64
}

// cardDrillEntry는 드릴다운 거래 항목
type cardDrillEntry struct {
	Date       string
	Item       string
	Money      int64
	LAccountID string
	RAccountID string
	Memo       string
}

// ─── 카드 목록 항목 및 Delegate ───────────────────────────────

type cardListItem struct {
	entry cardEntry
	index int // 번호 표시용
}

func (i cardListItem) Title() string       { return i.entry.Title }
func (i cardListItem) Description() string { return "" }
func (i cardListItem) FilterValue() string { return i.entry.Title }

// cardListDelegate는 카드 목록 한 줄 렌더 delegate
// [결제임박] 마커를 별도 스타일로 렌더
type cardListDelegate struct{}

func (d cardListDelegate) Height() int                               { return 1 }
func (d cardListDelegate) Spacing() int                              { return 0 }
func (d cardListDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd  { return nil }
func (d cardListDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ci := item.(cardListItem)
	urgentMark := ""
	if ci.entry.UrgentPay {
		urgentMark = confirmStyle.Render(" [결제임박]")
	}
	line := fmt.Sprintf(" %d. %-22s", ci.index+1, ci.entry.Title)
	if ci.entry.Category == "creditcard" && ci.entry.PayDate > 0 {
		line += fmt.Sprintf("  결제일 %2d일", ci.entry.PayDate)
	}
	if index == m.Index() {
		fmt.Fprint(w, selectedStyle.Render(">"+line)+urgentMark)
	} else {
		fmt.Fprint(w, " "+line+urgentMark)
	}
}

// ─── 서브 모델 ────────────────────────────────────────────────

type cardSubModel struct {
	cfg    *config.Config
	client *api.WhooingClient
	mode   cardMode
	errMsg string

	// 탭 선택 (번호 목록)
	activeTab cardTab
	tabList   list.Model

	// 카드 목록 (bubbles/list)
	cards    []cardEntry // 원본 데이터 (loadTable에서 참조)
	cardList list.Model

	// 월별 테이블 (수동 커서 — 테이블 뷰)
	rows      []cardMonthRow
	rowCursor int

	// 드릴다운 거래 목록 (수동 커서 — 테이블 뷰)
	drillEntries []cardDrillEntry
	drillCursor  int
	drillYM      string // 드릴다운 중인 연월
}

func newCardSubModel(cfg *config.Config) *cardSubModel {
	tabItems := []list.Item{
		simpleItem{"신용카드"},
		simpleItem{"체크카드"},
	}
	return &cardSubModel{
		cfg:     cfg,
		client:  NewClient(cfg),
		mode:    cardModeTabSelect,
		tabList: newCompactList(tabItems, 30, len(tabItems)+2),
	}
}

func (m *cardSubModel) buildCardList(cards []cardEntry) list.Model {
	items := make([]list.Item, len(cards))
	for i, c := range cards {
		items[i] = cardListItem{entry: c, index: i}
	}
	return newCompactListWith(items, cardListDelegate{}, 60, len(items)+2)
}

// ─── 비동기 메시지 ────────────────────────────────────────────

type cardCardsLoadedMsg struct {
	tab   cardTab
	cards []cardEntry
	err   error
}

type cardTableLoadedMsg struct {
	rows []cardMonthRow
	err  error
}

type cardDrillLoadedMsg struct {
	ym      string
	entries []cardDrillEntry
	err     error
}

// ─── 비동기 커맨드 ────────────────────────────────────────────

func (m *cardSubModel) loadCards(tab cardTab) tea.Cmd {
	return func() tea.Msg {
		raw, err := m.client.GetAccountsMap(m.cfg.SectionID)
		if err != nil {
			return cardCardsLoadedMsg{tab: tab, err: err}
		}
		cards := filterCardAccounts(raw, tab)
		return cardCardsLoadedMsg{tab: tab, cards: cards}
	}
}

func (m *cardSubModel) loadTable(card cardEntry) tea.Cmd {
	return func() tea.Msg {
		q := api.CardQuery{
			SectionID: m.cfg.SectionID,
			AccountID: card.AccountID,
		}
		var raw []byte
		var err error
		if card.Category == "creditcard" {
			raw, err = m.client.GetBill(q)
		} else {
			raw, err = m.client.GetCheckcard(q)
		}
		if err != nil {
			return cardTableLoadedMsg{err: err}
		}
		rows := parseCardTableRows(raw)
		return cardTableLoadedMsg{rows: rows}
	}
}

func (m *cardSubModel) loadDrilldown(card cardEntry, ym string) tea.Cmd {
	return func() tea.Msg {
		ymInt := 0
		fmt.Sscanf(ym, "%d", &ymInt)
		startDate := ymInt * 100  // YYYYMM01
		endDate := ymInt*100 + 31 // YYYYMM31 (서버가 실제 말일로 처리)

		q := api.EntrySearch{
			SectionID: m.cfg.SectionID,
			StartDate: startDate,
			EndDate:   endDate,
			Limit:     100,
		}
		if card.Category == "creditcard" {
			q.LAccountID = card.AccountID
			q.LAccount = "liabilities"
		} else {
			q.LAccountID = card.AccountID
			q.LAccount = "assets"
		}
		raw, err := m.client.SearchEntries(q)
		if err != nil {
			return cardDrillLoadedMsg{ym: ym, err: err}
		}
		entries := parseCardDrillEntries(raw)
		return cardDrillLoadedMsg{ym: ym, entries: entries}
	}
}

// ─── BubbleTea ────────────────────────────────────────────────

func (m *cardSubModel) Init() tea.Cmd {
	return nil
}

func (m *cardSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case cardCardsLoadedMsg:
		if msg.err != nil {
			m.mode = cardModeError
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.activeTab = msg.tab
		m.cards = msg.cards
		m.cardList = m.buildCardList(msg.cards)
		m.mode = cardModeCardList
		return m, nil

	case cardTableLoadedMsg:
		if msg.err != nil {
			m.mode = cardModeError
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.rows = msg.rows
		m.rowCursor = 0
		m.mode = cardModeMonthTable
		return m, nil

	case cardDrillLoadedMsg:
		if msg.err != nil {
			m.mode = cardModeError
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.drillYM = msg.ym
		m.drillEntries = msg.entries
		m.drillCursor = 0
		m.mode = cardModeDrilldown
		return m, nil

	case tea.KeyMsg:
		if GlobalAction(msg) == ActionQuit {
			return m, tea.Quit
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *cardSubModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case cardModeTabSelect:
		return m.handleTabKey(msg)
	case cardModeCardList:
		return m.handleCardListKey(msg)
	case cardModeMonthTable:
		return m.handleTableKey(msg)
	case cardModeDrilldown:
		return m.handleDrillKey(msg)
	case cardModeError:
		switch ErrorAction(msg) {
		case ActionBack, ActionExit:
			return m, func() tea.Msg { return backToMenuMsg{} }
		}
	}
	return m, nil
}

func (m *cardSubModel) handleTabKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch ListAction(msg) {
	case ActionBack, ActionExit:
		return m, func() tea.Msg { return backToMenuMsg{} }
	case ActionConfirm:
		tab := cardTab(m.tabList.Index())
		m.activeTab = tab
		m.mode = cardModeLoading
		return m, m.loadCards(tab)
	}
	if idx, ok := NumberAction(msg); ok && idx < 2 {
		tab := cardTab(idx)
		m.tabList.Select(idx)
		m.activeTab = tab
		m.mode = cardModeLoading
		return m, m.loadCards(tab)
	}
	var cmd tea.Cmd
	m.tabList, cmd = m.tabList.Update(msg)
	return m, cmd
}

func (m *cardSubModel) handleCardListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch ListAction(msg) {
	case ActionBack:
		m.mode = cardModeTabSelect
		return m, nil
	case ActionConfirm:
		if len(m.cards) > 0 {
			card := m.cards[m.cardList.Index()]
			m.mode = cardModeLoading
			return m, m.loadTable(card)
		}
	case ActionExit:
		return m, func() tea.Msg { return backToMenuMsg{} }
	}
	if idx, ok := NumberAction(msg); ok && idx < len(m.cards) {
		m.cardList.Select(idx)
		m.mode = cardModeLoading
		return m, m.loadTable(m.cards[idx])
	}
	var cmd tea.Cmd
	m.cardList, cmd = m.cardList.Update(msg)
	return m, cmd
}

func (m *cardSubModel) handleTableKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch ListAction(msg) {
	case ActionBack:
		m.mode = cardModeCardList
		return m, nil
	case ActionConfirm:
		if len(m.rows) > 0 && len(m.cards) > 0 {
			row := m.rows[m.rowCursor]
			card := m.cards[m.cardList.Index()]
			m.mode = cardModeLoading
			return m, m.loadDrilldown(card, row.YM)
		}
	case ActionMoveUp:
		if m.rowCursor > 0 {
			m.rowCursor--
		}
	case ActionMoveDown:
		if m.rowCursor < len(m.rows)-1 {
			m.rowCursor++
		}
	case ActionExit:
		return m, func() tea.Msg { return backToMenuMsg{} }
	}
	return m, nil
}

func (m *cardSubModel) handleDrillKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch ListAction(msg) {
	case ActionBack:
		m.mode = cardModeMonthTable
		return m, nil
	case ActionMoveUp:
		if m.drillCursor > 0 {
			m.drillCursor--
		}
	case ActionMoveDown:
		if m.drillCursor < len(m.drillEntries)-1 {
			m.drillCursor++
		}
	case ActionExit:
		return m, func() tea.Msg { return backToMenuMsg{} }
	}
	return m, nil
}

// ─── View ─────────────────────────────────────────────────────

func (m *cardSubModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("카드 관리") + "\n\n")

	switch m.mode {
	case cardModeTabSelect:
		b.WriteString(headerStyle.Render("카드 유형 선택") + "\n\n")
		b.WriteString(m.tabList.View() + "\n")
		b.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [1-2] 번호 선택  [Enter] 확인  [Esc/q] 메뉴"))

	case cardModeLoading:
		b.WriteString(loadingStyle.Render("불러오는 중..."))

	case cardModeCardList:
		tabLabel := "신용카드"
		if m.activeTab == cardTabCheckcard {
			tabLabel = "체크카드"
		}
		b.WriteString(headerStyle.Render(tabLabel) + "\n\n")
		if len(m.cards) == 0 {
			b.WriteString("  카드 항목이 없습니다\n")
		} else {
			b.WriteString(m.cardList.View() + "\n")
		}
		b.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [1-9] 번호 선택  [Enter] 확인  [Esc] 탭선택  [q] 메뉴"))

	case cardModeMonthTable:
		card := cardEntry{}
		if len(m.cards) > m.cardList.Index() {
			card = m.cards[m.cardList.Index()]
		}
		b.WriteString(headerStyle.Render(fmt.Sprintf("월별 청구 [%s]", card.Title)) + "\n\n")
		if len(m.rows) == 0 {
			b.WriteString("  청구 내역이 없습니다\n")
		} else {
			b.WriteString(fmt.Sprintf("  %-8s  %12s\n", "연월", "금액"))
			b.WriteString("  " + strings.Repeat("-", 22) + "\n")
			for i, row := range m.rows {
				line := fmt.Sprintf("  %-8s  %12s", formatYM(row.YM), FormatMoney(float64(row.Money)))
				if i == m.rowCursor {
					b.WriteString(selectedStyle.Render(">"+line[1:]) + "\n")
				} else {
					b.WriteString(line + "\n")
				}
			}
		}
		b.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [Enter] 거래 목록  [Esc] 카드목록  [q] 메뉴"))

	case cardModeDrilldown:
		card := cardEntry{}
		if len(m.cards) > m.cardList.Index() {
			card = m.cards[m.cardList.Index()]
		}
		b.WriteString(headerStyle.Render(fmt.Sprintf("거래 목록 [%s %s]", card.Title, formatYM(m.drillYM))) + "\n\n")
		if len(m.drillEntries) == 0 {
			b.WriteString("  거래 내역이 없습니다\n")
		} else {
			b.WriteString(fmt.Sprintf("  %-10s  %-20s  %10s\n", "날짜", "아이템", "금액"))
			b.WriteString("  " + strings.Repeat("-", 44) + "\n")
			for i, e := range m.drillEntries {
				line := fmt.Sprintf("  %-10s  %-20s  %10s",
					formatEntryDate(e.Date), truncate(e.Item, 20), FormatMoney(float64(e.Money)))
				if i == m.drillCursor {
					b.WriteString(selectedStyle.Render(">"+line[1:]) + "\n")
				} else {
					b.WriteString(line + "\n")
				}
			}
		}
		b.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [Esc] 월별테이블  [q] 메뉴"))

	case cardModeError:
		b.WriteString(errorStyle.Render("[오류] "+m.errMsg) + "\n\n")
		b.WriteString(helpStyle.Render("[아무 키] 메뉴로 돌아가기"))
	}

	return b.String()
}

// ─── 파싱 헬퍼 ────────────────────────────────────────────────

// filterCardAccounts는 AccountsMap에서 카드 카테고리 항목만 필터링
func filterCardAccounts(accountsMap *api.AccountsMap, tab cardTab) []cardEntry {
	today := time.Now()
	var result []cardEntry

	checkCategory := func(category string) bool {
		if tab == cardTabBill {
			return category == "creditcard"
		}
		return category == "checkcard"
	}

	allTypes := []string{"assets", "liabilities", "capital", "expenses", "income"}
	for _, accType := range allTypes {
		accounts := accountsMap.GetAccountsByType(accType)
		for id, detail := range accounts {
			if !checkCategory(detail.Category) {
				continue
			}
			result = append(result, cardEntry{
				AccountID: id,
				Title:     detail.Title,
				Category:  detail.Category,
			})
		}
	}

	_ = today

	sort.Slice(result, func(i, j int) bool {
		return result[i].Title < result[j].Title
	})
	return result
}

// parseCardTableRows는 bill/checkcard API 응답에서 월별 행 파싱
func parseCardTableRows(raw []byte) []cardMonthRow {
	var outer struct {
		Results json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(raw, &outer); err != nil {
		return nil
	}

	var results struct {
		Rows map[string]struct {
			Money int64 `json:"money"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(outer.Results, &results); err != nil {
		return nil
	}

	var rows []cardMonthRow
	for ym, row := range results.Rows {
		rows = append(rows, cardMonthRow{YM: ym, Money: row.Money})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].YM > rows[j].YM // 최신순
	})
	return rows
}

// parseCardDrillEntries는 SearchEntries 응답에서 드릴다운 거래 목록 파싱
func parseCardDrillEntries(raw []byte) []cardDrillEntry {
	var outer struct {
		Results json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(raw, &outer); err != nil {
		return nil
	}

	var entries []api.Entry
	if err := json.Unmarshal(outer.Results, &entries); err != nil {
		var wrapper struct {
			Entries []api.Entry `json:"entries"`
		}
		if err2 := json.Unmarshal(outer.Results, &wrapper); err2 != nil {
			return nil
		}
		entries = wrapper.Entries
	}

	var result []cardDrillEntry
	for _, e := range entries {
		result = append(result, cardDrillEntry{
			Date:       e.DateOnly(),
			Item:       e.Item,
			Money:      int64(e.Money),
			LAccountID: e.LAccountID,
			RAccountID: e.RAccountID,
			Memo:       e.Memo,
		})
	}
	return result
}

// ─── 뷰 헬퍼 ──────────────────────────────────────────────────

func formatYM(ym string) string {
	if len(ym) == 6 {
		return ym[:4] + "-" + ym[4:]
	}
	return ym
}

func formatEntryDate(d string) string {
	if len(d) == 8 {
		return d[:4] + "-" + d[4:6] + "-" + d[6:]
	}
	return d
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n-1]) + "…"
	}
	return s
}
