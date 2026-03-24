// cmd/app.go
// 통합 애플리케이션 - views 패턴으로 구현

package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"whooing-cli/config"
)

// appState는 앱의 현재 화면 상태
type appState int

const (
	stateAuth appState = iota
	stateSection
	stateMenu
	stateTransactions
	stateEntry
	stateBalance
	stateUserInfo
	stateExit
)

// stateTransitionMsg는 화면 전환 메시지
type stateTransitionMsg struct {
	newState appState
}

// backToMenuMsg는 메뉴로 돌아가기 메시지
type backToMenuMsg struct{}

// backToTransactionsMsg는 거래내역으로 돌아가기 메시지
type backToTransactionsMsg struct{}

// appModel은 통합 앱의 메인 모델
type appModel struct {
	state  appState
	cfg    *config.Config
	width  int
	height int

	// 서브 모델들
	authModel         *authSubModel
	sectionModel      *sectionSubModel
	menuModel         *menuSubModel
	transactionsModel *transactionsSubModel
	balanceModel      *balanceSubModel
	entryModel        *entrySubModel
	userInfoModel     *userInfoSubModel
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
		m.menuModel = newMenuSubModel(m.cfg)
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
		// 메뉴 선택에 따라 해당 화면으로
		switch msg.selection {
		case 0: // 거래내역
			m.transitionTo(stateTransactions)
		case 1: // 거래입력
			m.transitionTo(stateEntry)
		case 2: // 자산부채
			m.transitionTo(stateBalance)
		case 3: // 섹션변경
			m.transitionTo(stateSection)
		case 4: // 사용자 정보
			m.transitionTo(stateUserInfo)
		}
		return m, m.initCurrentState()

	case editEntryMsg:
		// 거래 수정
		m.entryModel = newEntrySubModelForEdit(m.cfg, msg.entry, msg.accountsMap)
		m.transitionTo(stateEntry)
		return m, m.entryModel.Init()

	case backToTransactionsMsg:
		// 거래내역으로 복귀
		m.transitionTo(stateTransactions)
		return m, m.initCurrentState()
	}

	// 현재 상태에 따른 서브 업데이트
	return m.updateSubModel(msg)
}

// transitionTo는 새로운 상태로 전환
func (m *appModel) transitionTo(newState appState) {
	// 이전 서브모델 정리
	switch m.state {
	case stateAuth:
		m.authModel = nil
	case stateSection:
		m.sectionModel = nil
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
	}

	m.state = newState

	// 새 서브모델 생성
	switch newState {
	case stateAuth:
		m.authModel = newAuthSubModel(m.cfg)
	case stateSection:
		m.sectionModel = newSectionSubModel(m.cfg)
	case stateMenu:
		m.menuModel = newMenuSubModel(m.cfg)
	case stateTransactions:
		m.transactionsModel = newTransactionsSubModel(m.cfg)
	case stateBalance:
		m.balanceModel = newBalanceSubModel(m.cfg)
	case stateEntry:
		m.entryModel = newEntrySubModel(m.cfg)
	case stateUserInfo:
		m.userInfoModel = newUserInfoSubModel(m.cfg)
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

