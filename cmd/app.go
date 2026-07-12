// cmd/app.go
// 통합 애플리케이션 - views 패턴으로 구현

package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"whoo-cli/config"
)

// appState는 앱의 현재 화면 상태
type appState int

const (
	stateAuth appState = iota
	stateSection
	stateSectionHub // 메인 메뉴 "섹션" → 변경/관리 선택
	stateMenu
	stateTransactions
	stateEntry
	stateBalance
	stateUserInfo
	stateSectionManage
	stateAccountManage
	stateFlow
	stateFrequent
	stateMonthly
	stateCard
	stateBudget
	stateExit
)

// stateTransitionMsg는 화면 전환 메시지
type stateTransitionMsg struct {
	newState appState
}

// backToMenuMsg는 메뉴로 돌아가기 메시지
type backToMenuMsg struct{}

// backToSectionHubMsg는 섹션 허브(변경/관리 선택)로 돌아가기
type backToSectionHubMsg struct{}

// backToTransactionsMsg는 거래내역으로 돌아가기 메시지
type backToTransactionsMsg struct{}

// appModel은 통합 앱의 메인 모델
type appModel struct {
	state  appState
	cfg    *config.Config
	width  int
	height int
	nav    navMemory // 메인/1뎁스 커서 복원

	// 서브 모델들
	authModel            *authSubModel
	sectionModel         *sectionSubModel
	sectionHubModel      *sectionHubSubModel
	menuModel            *menuSubModel
	transactionsModel    *transactionsSubModel
	balanceModel         *balanceSubModel
	entryModel           *entrySubModel
	userInfoModel        *userInfoSubModel
	sectionManageModel   *sectionManageSubModel
	accountManageModel   *accountManageSubModel
	flowModel            *flowSubModel
	frequentModel        *frequentSubModel
	monthlyModel         *monthlySubModel
	cardModel            *cardSubModel
	budgetModel          *budgetSubModel
}

// newAppModel은 새로운 앱 모델을 생성
func newAppModel(cfg *config.Config) *appModel {
	return &appModel{
		state: stateAuth,
		cfg:   cfg,
	}
}

// determineInitialState는 인증/섹션 상태에 따라 초기 상태 결정
func (m *appModel) determineInitialState() {
	if !m.cfg.IsAuthenticated() {
		m.state = stateAuth
		m.authModel = newAuthSubModel(m.cfg)
	} else if m.cfg.SectionID == "" {
		m.state = stateSection
		m.sectionModel = newSectionSubModel(m.cfg)
	} else {
		m.state = stateMenu
		m.menuModel = newMenuSubModel(m.cfg, m.nav.menuIndex)
	}
}

// Init은 초기 커맨드를 반환
func (m *appModel) Init() tea.Cmd {
	m.determineInitialState()

	switch m.state {
	case stateAuth:
		if m.authModel != nil {
			return m.authModel.Init()
		}
	case stateSection:
		if m.sectionModel != nil {
			return m.sectionModel.Init()
		}
	case stateMenu:
		if m.menuModel != nil {
			return m.menuModel.Init()
		}
	}
	return nil
}

// Update는 메시지를 처리하고 상태를 업데이트
func (m *appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.state = stateExit
			return m, tea.Quit
		}

	case stateTransitionMsg:
		m.transitionTo(msg.newState)
		return m, m.initCurrentState()

	case backToMenuMsg:
		m.transitionTo(stateMenu)
		return m, m.initCurrentState()

	case backToSectionHubMsg:
		// 섹션 변경/관리에서 Esc → 허브로 복귀
		m.transitionTo(stateSectionHub)
		return m, m.initCurrentState()

	case authCompleteMsg:
		// 인증 완료 후 섹션 선택으로
		m.cfg = msg.cfg
		m.transitionTo(stateSection)
		return m, m.initCurrentState()

	case sectionSelectedMsg:
		// 섹션 선택 후 메뉴로
		m.cfg = msg.cfg
		m.transitionTo(stateMenu)
		return m, m.initCurrentState()

	case menuSelectionMsg:
		// 번호 바로선택 포함 — 떠날 커서 명시 저장
		m.nav.menuIndex = msg.selection
		// 메뉴 선택에 따라 해당 화면으로 (menu_sub.go 항목 순서와 일치)
		switch msg.selection {
		case 0: // 거래내역
			m.transitionTo(stateTransactions)
		case 1: // 거래입력
			m.transitionTo(stateEntry)
		case 2: // 자산부채
			m.transitionTo(stateBalance)
		case 3: // 자주입력
			m.transitionTo(stateFrequent)
		case 4: // 월별입력
			m.transitionTo(stateMonthly)
		case 5: // 항목 관리
			m.transitionTo(stateAccountManage)
		case 6: // 흐름 분석
			m.transitionTo(stateFlow)
		case 7: // 카드 관리
			m.transitionTo(stateCard)
		case 8: // 예산/목표
			m.transitionTo(stateBudget)
		case 9: // 섹션 (변경/관리 허브)
			m.transitionTo(stateSectionHub)
		case 10: // 사용자 정보
			m.transitionTo(stateUserInfo)
		}
		return m, m.initCurrentState()

	case editEntryMsg:
		// 거래 수정: transitionTo는 신규 entry 모델을 만들므로
		// 이전 상태만 정리한 뒤 수정용 모델을 직접 설정한다.
		m.clearSubModel(m.state)
		m.state = stateEntry
		m.entryModel = newEntrySubModelForEdit(m.cfg, msg.entry, msg.accountsMap)
		return m, m.entryModel.Init()

	case backToTransactionsMsg:
		// 거래내역으로 복귀
		m.transitionTo(stateTransactions)
		return m, m.initCurrentState()
	}

	// 현재 상태에 따른 서브 업데이트
	return m.updateSubModel(msg)
}

// clearSubModel은 지정 상태의 서브모델을 정리한다
func (m *appModel) clearSubModel(state appState) {
	switch state {
	case stateAuth:
		m.authModel = nil
	case stateSection:
		m.sectionModel = nil
	case stateSectionHub:
		m.sectionHubModel = nil
	case stateMenu:
		m.menuModel = nil
	case stateTransactions:
		m.transactionsModel = nil
	case stateBalance:
		m.balanceModel = nil
	case stateEntry:
		m.entryModel = nil
	case stateUserInfo:
		m.userInfoModel = nil
	case stateSectionManage:
		m.sectionManageModel = nil
	case stateAccountManage:
		m.accountManageModel = nil
	case stateFlow:
		m.flowModel = nil
	case stateFrequent:
		m.frequentModel = nil
	case stateMonthly:
		m.monthlyModel = nil
	case stateCard:
		m.cardModel = nil
	case stateBudget:
		m.budgetModel = nil
	}
}

// captureNav는 현재 화면의 1뎁스 커서를 navMemory에 저장한다
func (m *appModel) captureNav() {
	switch m.state {
	case stateMenu:
		if m.menuModel != nil {
			m.nav.menuIndex = m.menuModel.cursorIndex()
		}
	case stateSectionHub:
		if m.sectionHubModel != nil {
			m.nav.sectionHubIndex = m.sectionHubModel.cursorIndex()
		}
	case stateFrequent:
		if m.frequentModel != nil {
			m.nav.frequentSlot = m.frequentModel.slotIndex()
		}
	case stateMonthly:
		if m.monthlyModel != nil {
			m.nav.monthlySlot = m.monthlyModel.slotIndex()
		}
	case stateCard:
		if m.cardModel != nil {
			m.nav.cardTab = m.cardModel.tabIndex()
		}
	case stateBudget:
		if m.budgetModel != nil {
			m.nav.budgetType = m.budgetModel.typeIndex()
		}
	case stateAccountManage:
		if m.accountManageModel != nil {
			m.nav.accountType = m.accountManageModel.typeIndex()
		}
	case stateFlow:
		if m.flowModel != nil {
			m.nav.flowType = m.flowModel.typeIndex()
		}
	}
}

// transitionTo는 새로운 상태로 전환
func (m *appModel) transitionTo(newState appState) {
	// 허브에서 변경/관리로 들어가면 Esc 시 허브로 복귀
	fromSectionHub := m.state == stateSectionHub

	// 소멸 전에 커서 기억
	m.captureNav()

	m.clearSubModel(m.state)
	m.state = newState

	// 새 서브모델 생성 (nav 인덱스로 커서 복원)
	switch newState {
	case stateAuth:
		m.authModel = newAuthSubModel(m.cfg)
	case stateSection:
		m.sectionModel = newSectionSubModel(m.cfg)
		m.sectionModel.fromHub = fromSectionHub
	case stateSectionHub:
		m.sectionHubModel = newSectionHubSubModel(m.cfg, m.nav.sectionHubIndex)
	case stateMenu:
		m.menuModel = newMenuSubModel(m.cfg, m.nav.menuIndex)
	case stateTransactions:
		m.transactionsModel = newTransactionsSubModel(m.cfg)
	case stateBalance:
		m.balanceModel = newBalanceSubModel(m.cfg)
	case stateEntry:
		m.entryModel = newEntrySubModel(m.cfg)
	case stateUserInfo:
		m.userInfoModel = newUserInfoSubModel(m.cfg)
	case stateSectionManage:
		m.sectionManageModel = newSectionManageSubModel(m.cfg)
		m.sectionManageModel.fromHub = fromSectionHub
	case stateAccountManage:
		m.accountManageModel = newAccountManageSubModel(m.cfg, m.nav.accountType)
		// 이미 수신된 터미널 크기 반영 (모델 생성 전 WindowSize 유실 방지)
		if m.width > 0 {
			m.accountManageModel.width = m.width
			m.accountManageModel.height = m.height
			m.accountManageModel.resizeLists()
		}
	case stateFlow:
		m.flowModel = newFlowSubModel(m.cfg, m.nav.flowType)
	case stateFrequent:
		m.frequentModel = newFrequentSubModel(m.cfg, m.nav.frequentSlot)
	case stateMonthly:
		m.monthlyModel = newMonthlySubModel(m.cfg, m.nav.monthlySlot)
	case stateCard:
		m.cardModel = newCardSubModel(m.cfg, m.nav.cardTab)
	case stateBudget:
		m.budgetModel = newBudgetSubModel(m.cfg, m.nav.budgetType)
	}
}

// initCurrentState는 현재 상태의 서브모델 Init을 반환
func (m *appModel) initCurrentState() tea.Cmd {
	switch m.state {
	case stateAuth:
		if m.authModel != nil {
			return m.authModel.Init()
		}
	case stateSection:
		if m.sectionModel != nil {
			return m.sectionModel.Init()
		}
	case stateSectionHub:
		if m.sectionHubModel != nil {
			return m.sectionHubModel.Init()
		}
	case stateMenu:
		if m.menuModel != nil {
			return m.menuModel.Init()
		}
	case stateTransactions:
		if m.transactionsModel != nil {
			return m.transactionsModel.Init()
		}
	case stateBalance:
		if m.balanceModel != nil {
			return m.balanceModel.Init()
		}
	case stateEntry:
		if m.entryModel != nil {
			return m.entryModel.Init()
		}
	case stateUserInfo:
		if m.userInfoModel != nil {
			return m.userInfoModel.Init()
		}
	case stateSectionManage:
		if m.sectionManageModel != nil {
			return m.sectionManageModel.Init()
		}
	case stateAccountManage:
		if m.accountManageModel != nil {
			return m.accountManageModel.Init()
		}
	case stateFlow:
		if m.flowModel != nil {
			return m.flowModel.Init()
		}
	case stateFrequent:
		if m.frequentModel != nil {
			return m.frequentModel.Init()
		}
	case stateMonthly:
		if m.monthlyModel != nil {
			return m.monthlyModel.Init()
		}
	case stateCard:
		if m.cardModel != nil {
			return m.cardModel.Init()
		}
	case stateBudget:
		if m.budgetModel != nil {
			return m.budgetModel.Init()
		}
	}
	return nil
}

// updateSubModel은 현재 상태의 서브모델 업데이트
func (m *appModel) updateSubModel(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateAuth:
		if m.authModel != nil {
			model, cmd := m.authModel.Update(msg)
			m.authModel = model.(*authSubModel)
			return m, cmd
		}
	case stateSection:
		if m.sectionModel != nil {
			model, cmd := m.sectionModel.Update(msg)
			m.sectionModel = model.(*sectionSubModel)
			return m, cmd
		}
	case stateSectionHub:
		if m.sectionHubModel != nil {
			model, cmd := m.sectionHubModel.Update(msg)
			m.sectionHubModel = model.(*sectionHubSubModel)
			return m, cmd
		}
	case stateMenu:
		if m.menuModel != nil {
			model, cmd := m.menuModel.Update(msg)
			m.menuModel = model.(*menuSubModel)
			return m, cmd
		}
	case stateTransactions:
		if m.transactionsModel != nil {
			model, cmd := m.transactionsModel.Update(msg)
			m.transactionsModel = model.(*transactionsSubModel)
			return m, cmd
		}
	case stateBalance:
		if m.balanceModel != nil {
			model, cmd := m.balanceModel.Update(msg)
			m.balanceModel = model.(*balanceSubModel)
			return m, cmd
		}
	case stateEntry:
		if m.entryModel != nil {
			model, cmd := m.entryModel.Update(msg)
			m.entryModel = model.(*entrySubModel)
			return m, cmd
		}
	case stateUserInfo:
		if m.userInfoModel != nil {
			model, cmd := m.userInfoModel.Update(msg)
			m.userInfoModel = model.(*userInfoSubModel)
			return m, cmd
		}
	case stateSectionManage:
		if m.sectionManageModel != nil {
			model, cmd := m.sectionManageModel.Update(msg)
			m.sectionManageModel = model.(*sectionManageSubModel)
			return m, cmd
		}
	case stateAccountManage:
		if m.accountManageModel != nil {
			model, cmd := m.accountManageModel.Update(msg)
			m.accountManageModel = model.(*accountManageSubModel)
			return m, cmd
		}
	case stateFlow:
		if m.flowModel != nil {
			model, cmd := m.flowModel.Update(msg)
			m.flowModel = model.(*flowSubModel)
			return m, cmd
		}
	case stateFrequent:
		if m.frequentModel != nil {
			model, cmd := m.frequentModel.Update(msg)
			m.frequentModel = model.(*frequentSubModel)
			return m, cmd
		}
	case stateMonthly:
		if m.monthlyModel != nil {
			model, cmd := m.monthlyModel.Update(msg)
			m.monthlyModel = model.(*monthlySubModel)
			return m, cmd
		}
	case stateCard:
		if m.cardModel != nil {
			model, cmd := m.cardModel.Update(msg)
			m.cardModel = model.(*cardSubModel)
			return m, cmd
		}
	case stateBudget:
		if m.budgetModel != nil {
			model, cmd := m.budgetModel.Update(msg)
			m.budgetModel = model.(*budgetSubModel)
			return m, cmd
		}
	}
	return m, nil
}

// View는 현재 상태에 따른 뷰를 렌더링
func (m *appModel) View() string {
	if m.state == stateExit {
		return ""
	}

	switch m.state {
	case stateAuth:
		if m.authModel != nil {
			return m.authModel.View()
		}
	case stateSection:
		if m.sectionModel != nil {
			return m.sectionModel.View()
		}
	case stateSectionHub:
		if m.sectionHubModel != nil {
			return m.sectionHubModel.View()
		}
	case stateMenu:
		if m.menuModel != nil {
			return m.menuModel.View()
		}
	case stateTransactions:
		if m.transactionsModel != nil {
			return m.transactionsModel.View()
		}
	case stateBalance:
		if m.balanceModel != nil {
			return m.balanceModel.View()
		}
	case stateEntry:
		if m.entryModel != nil {
			return m.entryModel.View()
		}
	case stateUserInfo:
		if m.userInfoModel != nil {
			return m.userInfoModel.View()
		}
	case stateSectionManage:
		if m.sectionManageModel != nil {
			return m.sectionManageModel.View()
		}
	case stateAccountManage:
		if m.accountManageModel != nil {
			return m.accountManageModel.View()
		}
	case stateFlow:
		if m.flowModel != nil {
			return m.flowModel.View()
		}
	case stateFrequent:
		if m.frequentModel != nil {
			return m.frequentModel.View()
		}
	case stateMonthly:
		if m.monthlyModel != nil {
			return m.monthlyModel.View()
		}
	case stateCard:
		if m.cardModel != nil {
			return m.cardModel.View()
		}
	case stateBudget:
		if m.budgetModel != nil {
			return m.budgetModel.View()
		}
	}

	return ""
}

// RunApp은 통합 앱을 실행
func RunApp(cfg *config.Config) {
	m := newAppModel(cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("[오류] 앱 실행 실패: %v\n", err)
	}
}

