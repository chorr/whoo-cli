// cmd/account_manage_sub.go
// 항목 관리 - bubbletea 서브 모델 (생성/수정/삭제/정렬)

package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"whoo-cli/api"
	"whoo-cli/config"
)

// accountManageMode는 항목 관리 서브 모드
type accountManageMode int

const (
	accountManageModeTypeSelect accountManageMode = iota
	accountManageModeList
	accountManageModeAdd
	accountManageModeEdit
	accountManageModeConfirmDelete
	accountManageModeLoading
	accountManageModeError
)

// ─── 목록 항목 타입 ────────────────────────────────────────────

// accountTypeListItem은 계정 타입 선택용 목록 항목
type accountTypeListItem struct {
	code string
	name string
}

func (i accountTypeListItem) Title() string       { return fmt.Sprintf("%s (%s)", i.name, i.code) }
func (i accountTypeListItem) Description() string { return "" }
func (i accountTypeListItem) FilterValue() string { return i.name }

// accountDetailListItem은 계정 항목 목록용 래퍼
type accountDetailListItem struct {
	id     string
	detail api.AccountDetail
}

func (i accountDetailListItem) Title() string {
	typeTag := ""
	if i.detail.Type == "group" {
		typeTag = " [그룹]"
	}
	return fmt.Sprintf("%s%s (%s)", i.detail.Title, typeTag, i.detail.Category)
}
func (i accountDetailListItem) Description() string { return "" }
func (i accountDetailListItem) FilterValue() string { return i.detail.Title }

// ─── 서브 모델 ────────────────────────────────────────────────

type accountManageSubModel struct {
	cfg         *config.Config
	client      *api.WhooingClient
	accountsMap *api.AccountsMap
	mode        accountManageMode
	errMsg      string

	// 계정 타입 선택 (bubbles/list)
	typeList    list.Model
	accountType string // 현재 선택된 타입 코드

	// 항목 목록 (bubbles/list)
	accountList list.Model
	accountIDs  []string // accountList에 대응하는 ID 슬라이스
	accountData []api.AccountDetail

	// 폼 입력 (add/edit)
	formStep     int
	formTitle    string
	formMemo     string
	formCategory string
	textInput    string
	editingID    string
	existsResult *api.AccountExistsResult
}

const (
	accountFormStepTitle = iota
	accountFormStepMemo
	accountFormStepCategory
	accountFormStepConfirm
)

func newAccountManageSubModel(cfg *config.Config) *accountManageSubModel {
	// 계정 타입 선택 목록 초기화
	items := make([]list.Item, len(AccountTypes))
	for i, at := range AccountTypes {
		items[i] = accountTypeListItem{code: at.Code, name: at.Name}
	}
	typeList := newCompactList(items, 40, len(items)+2)

	return &accountManageSubModel{
		cfg:      cfg,
		client:   NewClient(cfg),
		mode:     accountManageModeLoading,
		typeList: typeList,
	}
}

func (m *accountManageSubModel) Init() tea.Cmd {
	return m.loadAccounts()
}

func (m *accountManageSubModel) loadAccounts() tea.Cmd {
	return func() tea.Msg {
		am, err := m.client.GetAccountsMap(m.cfg.SectionID)
		if err != nil {
			return accountManageErrMsg{err: err}
		}
		return accountManageLoadedMsg{accountsMap: am}
	}
}

type accountManageLoadedMsg struct{ accountsMap *api.AccountsMap }
type accountManageErrMsg struct{ err error }

func (m *accountManageSubModel) buildAccountList() {
	if m.accountsMap == nil {
		return
	}
	accounts := m.accountsMap.GetAccountsByType(m.accountType)
	m.accountIDs = nil
	m.accountData = nil
	var items []list.Item
	for id, detail := range accounts {
		m.accountIDs = append(m.accountIDs, id)
		m.accountData = append(m.accountData, detail)
		items = append(items, accountDetailListItem{id: id, detail: detail})
	}
	m.accountList = newPlainList(items, 60, len(items)+2)
}

func (m *accountManageSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case accountManageLoadedMsg:
		m.accountsMap = msg.accountsMap
		m.mode = accountManageModeTypeSelect
		m.typeList.Select(0)

	case accountManageErrMsg:
		m.mode = accountManageModeError
		m.errMsg = msg.err.Error()

	case tea.KeyMsg:
		if GlobalAction(msg) == ActionQuit {
			return m, tea.Quit
		}
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *accountManageSubModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case accountManageModeTypeSelect:
		return m.handleTypeSelectKey(msg)
	case accountManageModeList:
		return m.handleListKey(msg)
	case accountManageModeAdd, accountManageModeEdit:
		return m.handleFormKey(msg)
	case accountManageModeConfirmDelete:
		return m.handleDeleteKey(msg)
	case accountManageModeError:
		// enter는 재시도, esc/q는 메뉴 복귀 (이 화면의 특수 에러 정책)
		switch msg.String() {
		case "esc", "q":
			return m, func() tea.Msg { return backToMenuMsg{} }
		case "enter":
			m.mode = accountManageModeLoading
			return m, m.loadAccounts()
		}
	}
	return m, nil
}

func (m *accountManageSubModel) handleTypeSelectKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch ListAction(msg) {
	case ActionBack, ActionExit:
		return m, func() tea.Msg { return backToMenuMsg{} }
	case ActionConfirm:
		m.accountType = AccountTypes[m.typeList.Index()].Code
		m.buildAccountList()
		m.mode = accountManageModeList
		return m, nil
	}
	var cmd tea.Cmd
	m.typeList, cmd = m.typeList.Update(msg)
	return m, cmd
}

func (m *accountManageSubModel) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch ListAction(msg) {
	case ActionBack:
		m.mode = accountManageModeTypeSelect
		return m, nil
	case ActionExit:
		return m, func() tea.Msg { return backToMenuMsg{} }
	case ActionEdit:
		if len(m.accountData) > 0 {
			idx := m.accountList.Index()
			item := m.accountData[idx]
			m.mode = accountManageModeEdit
			m.editingID = m.accountIDs[idx]
			m.formStep = accountFormStepTitle
			m.formTitle = item.Title
			m.formMemo = item.Memo
			m.formCategory = item.Category
			m.textInput = item.Title
		}
		return m, nil
	case ActionDelete:
		if len(m.accountData) > 0 {
			return m, m.checkExists()
		}
		return m, nil
	}
	// 도메인 전용: a = 추가
	if msg.String() == "a" {
		m.mode = accountManageModeAdd
		m.formStep = accountFormStepTitle
		m.formTitle = ""
		m.formMemo = ""
		m.formCategory = "normal"
		m.textInput = ""
		return m, nil
	}
	var cmd tea.Cmd
	m.accountList, cmd = m.accountList.Update(msg)
	return m, cmd
}

func (m *accountManageSubModel) checkExists() tea.Cmd {
	idx := m.accountList.Index()
	account := m.accountType
	accountID := m.accountIDs[idx]
	sectionID := m.cfg.SectionID
	return func() tea.Msg {
		result, err := m.client.AccountExists(account, accountID, sectionID)
		if err != nil {
			return accountManageErrMsg{err: err}
		}
		m.existsResult = result
		m.mode = accountManageModeConfirmDelete
		return nil
	}
}

func (m *accountManageSubModel) handleFormKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEscape:
		m.mode = accountManageModeList
		m.textInput = ""
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

func (m *accountManageSubModel) advanceFormStep() (tea.Model, tea.Cmd) {
	switch m.formStep {
	case accountFormStepTitle:
		if strings.TrimSpace(m.textInput) == "" {
			return m, nil
		}
		m.formTitle = strings.TrimSpace(m.textInput)
		m.formStep = accountFormStepMemo
		m.textInput = m.formMemo
	case accountFormStepMemo:
		m.formMemo = strings.TrimSpace(m.textInput)
		m.formStep = accountFormStepCategory
		m.textInput = m.formCategory
	case accountFormStepCategory:
		if strings.TrimSpace(m.textInput) == "" {
			m.textInput = "normal"
		}
		m.formCategory = strings.TrimSpace(m.textInput)
		m.formStep = accountFormStepConfirm
		m.textInput = ""
	case accountFormStepConfirm:
		return m, m.submitForm()
	}
	return m, nil
}

func (m *accountManageSubModel) submitForm() tea.Cmd {
	account := m.accountType
	sectionID := m.cfg.SectionID
	if m.mode == accountManageModeAdd {
		return func() tea.Msg {
			_, err := m.client.CreateAccount(account, api.AccountCreateParams{
				SectionID: sectionID,
				Title:     m.formTitle,
				Type:      "account",
				OpenDate:  20010101,
				CloseDate: 29991231,
				Memo:      m.formMemo,
				Category:  m.formCategory,
			})
			if err != nil {
				return accountManageErrMsg{err: err}
			}
			am, err := m.client.GetAccountsMap(sectionID)
			if err != nil {
				return accountManageErrMsg{err: err}
			}
			return accountManageLoadedMsg{accountsMap: am}
		}
	}
	// 수정 모드 — 기존 open_date/close_date 유지
	editID := m.editingID
	var existingItem api.AccountDetail
	for i, id := range m.accountIDs {
		if id == editID {
			existingItem = m.accountData[i]
			break
		}
	}
	return func() tea.Msg {
		openDate := 20010101
		closeDate := 29991231
		if v, ok := existingItem.OpenDate.(float64); ok && v > 0 {
			openDate = int(v)
		}
		if v, ok := existingItem.CloseDate.(float64); ok && v > 0 {
			closeDate = int(v)
		}
		_, err := m.client.UpdateAccount(account, editID, api.AccountUpdateParams{
			SectionID: sectionID,
			Title:     m.formTitle,
			OpenDate:  openDate,
			CloseDate: closeDate,
			Memo:      m.formMemo,
			Category:  m.formCategory,
		})
		if err != nil {
			return accountManageErrMsg{err: err}
		}
		am, err := m.client.GetAccountsMap(sectionID)
		if err != nil {
			return accountManageErrMsg{err: err}
		}
		return accountManageLoadedMsg{accountsMap: am}
	}
}

func (m *accountManageSubModel) handleDeleteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch ConfirmAction(msg) {
	case ActionConfirm:
		return m, m.deleteAccount()
	case ActionBack:
		m.mode = accountManageModeList
		m.existsResult = nil
	}
	return m, nil
}

func (m *accountManageSubModel) deleteAccount() tea.Cmd {
	if len(m.accountData) == 0 {
		return nil
	}
	idx := m.accountList.Index()
	account := m.accountType
	accountID := m.accountIDs[idx]
	sectionID := m.cfg.SectionID
	return func() tea.Msg {
		_, err := m.client.DeleteAccount(account, accountID, sectionID)
		if err != nil {
			return accountManageErrMsg{err: err}
		}
		am, err := m.client.GetAccountsMap(sectionID)
		if err != nil {
			return accountManageErrMsg{err: err}
		}
		return accountManageLoadedMsg{accountsMap: am}
	}
}

func (m *accountManageSubModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("항목 관리") + "\n\n")

	switch m.mode {
	case accountManageModeLoading:
		b.WriteString("항목 목록을 불러오는 중...\n")

	case accountManageModeTypeSelect:
		b.WriteString(headerStyle.Render("계정 선택") + "\n\n")
		b.WriteString(m.typeList.View() + "\n")
		b.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [Enter] 선택  [Esc/q] 뒤로") + "\n")

	case accountManageModeList:
		b.WriteString(headerStyle.Render(fmt.Sprintf("%s 항목", FormatAccount(m.accountType))) + "\n\n")
		if len(m.accountData) == 0 {
			b.WriteString("항목이 없습니다.\n")
		} else {
			b.WriteString(m.accountList.View() + "\n")
		}
		b.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [a] 추가  [e] 수정  [d] 삭제  [Esc/q] 뒤로") + "\n")

	case accountManageModeAdd:
		b.WriteString(headerStyle.Render(fmt.Sprintf("%s 항목 추가", FormatAccount(m.accountType))) + "\n\n")
		b.WriteString(m.renderFormStep("추가"))

	case accountManageModeEdit:
		b.WriteString(headerStyle.Render("항목 수정") + "\n\n")
		b.WriteString(m.renderFormStep("수정"))

	case accountManageModeConfirmDelete:
		if len(m.accountData) > 0 {
			idx := m.accountList.Index()
			item := m.accountData[idx]
			b.WriteString(errorStyle.Render(fmt.Sprintf("'%s' 항목을 삭제하시겠습니까?", item.Title)) + "\n\n")
			if m.existsResult != nil {
				if m.existsResult.Count > 0 {
					b.WriteString(errorStyle.Render(fmt.Sprintf("  경고: 거래 %d건이 있습니다. 삭제 시 해당 거래의 항목이 x0으로 변환됩니다.", m.existsResult.Count)) + "\n")
				}
				if m.existsResult.LastOne == "y" {
					b.WriteString(errorStyle.Render("  마지막 항목이라 삭제할 수 없습니다.") + "\n\n")
					b.WriteString(helpStyle.Render("[Esc] 취소") + "\n")
					return b.String()
				}
			}
			b.WriteString("\n" + helpStyle.Render("[y] 삭제  [n/Esc] 취소") + "\n")
		}

	case accountManageModeError:
		b.WriteString(errorStyle.Render("[오류] "+m.errMsg) + "\n\n")
		b.WriteString(helpStyle.Render("[Enter] 재시도  [Esc] 뒤로") + "\n")
	}

	return b.String()
}

func (m *accountManageSubModel) renderFormStep(action string) string {
	var b strings.Builder
	switch m.formStep {
	case accountFormStepTitle:
		b.WriteString("이름: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case accountFormStepMemo:
		b.WriteString(fmt.Sprintf("이름: %s\n", m.formTitle))
		b.WriteString("설명: " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case accountFormStepCategory:
		b.WriteString(fmt.Sprintf("이름: %s  설명: %s\n", m.formTitle, m.formMemo))
		b.WriteString("종류 (normal/client/creditcard/checkcard/steady/floating): " + m.textInput + "_\n\n")
		b.WriteString(helpStyle.Render("[Enter] 다음  [Esc] 취소") + "\n")
	case accountFormStepConfirm:
		b.WriteString(fmt.Sprintf("이름: %s\n설명: %s\n종류: %s\n\n", m.formTitle, m.formMemo, m.formCategory))
		b.WriteString(helpStyle.Render(fmt.Sprintf("[Enter] %s 실행  [Esc] 취소", action)) + "\n")
	}
	return b.String()
}
