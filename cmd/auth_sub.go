// cmd/auth_sub.go
// 인증 화면 - bubbletea 서브 모델

package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"whooing-cli/auth"
	"whooing-cli/config"
)

// authCompleteMsg는 인증 완료 메시지
type authCompleteMsg struct {
	cfg *config.Config
}

// authSubModel은 인증 서브 모델
type authSubModel struct {
	state      int
	pinInput   textinput.Model
	oauth      *auth.OAuth
	cfg        *config.Config
	tokenResp  *auth.RequestTokenResponse
	accessResp *auth.AccessTokenResponse
	errMsg     string
	authURL    string
}

const (
	authStateInit = iota
	authStateRequestingToken
	authStateWaitingForPIN
	authStateExchangingToken
	authStateSuccess
	authStateError
)

// newAuthSubModel은 새로운 인증 모델을 생성
func newAuthSubModel(cfg *config.Config) *authSubModel {
	pinInput := textinput.New()
	pinInput.Placeholder = "PIN 번호 입력"
	pinInput.CharLimit = 20

	return &authSubModel{
		state:    authStateInit,
		pinInput: pinInput,
		oauth:    auth.NewOAuth(cfg),
		cfg:      cfg,
	}
}

func (m *authSubModel) Init() tea.Cmd {
	return tea.Batch(
		m.pinInput.Focus(),
		m.requestToken(),
	)
}

func (m *authSubModel) requestToken() tea.Cmd {
	return func() tea.Msg {
		tokenResp, err := m.oauth.RequestToken()
		if err != nil {
			return authErrMsg{err: err}
		}
		return tokenResp
	}
}

func (m *authSubModel) exchangeToken(pin string) tea.Cmd {
	return func() tea.Msg {
		accessResp, err := m.oauth.ExchangeToken(
			m.tokenResp.Token,
			m.tokenResp.Signiture,
			pin,
		)
		if err != nil {
			return authErrMsg{err: err}
		}
		return accessResp
	}
}

type authErrMsg struct {
	err error
}

func (m *authSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case *auth.RequestTokenResponse:
		m.tokenResp = msg
		m.authURL = m.oauth.GetAuthorizationURL(msg.Token)
		m.state = authStateWaitingForPIN
		openBrowser(m.authURL)
		return m, nil

	case *auth.AccessTokenResponse:
		m.accessResp = msg
		if err := m.oauth.CompleteAuth(msg.Token, msg.TokenSecret); err != nil {
			m.state = authStateError
			m.errMsg = fmt.Sprintf("토큰 저장 실패: %v", err)
			return m, nil
		}
		m.state = authStateSuccess
		// 설정 다시 로드
		cfg, _ := config.Load()
		return m, func() tea.Msg { return authCompleteMsg{cfg: cfg} }

	case authErrMsg:
		m.state = authStateError
		m.errMsg = msg.err.Error()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.state == authStateWaitingForPIN {
				pin := strings.TrimSpace(m.pinInput.Value())
				if pin != "" {
					m.state = authStateExchangingToken
					return m, m.exchangeToken(pin)
				}
			}
		}
	}

	var cmd tea.Cmd
	m.pinInput, cmd = m.pinInput.Update(msg)
	return m, cmd
}

func (m *authSubModel) View() string {
	urlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Underline(true)

	var content strings.Builder
	content.WriteString(titleStyle.Render("후잉 계정 인증") + "\n\n")

	switch m.state {
	case authStateInit, authStateRequestingToken:
		content.WriteString("인증 토큰을 요청하는 중...\n")

	case authStateWaitingForPIN:
		content.WriteString(headerStyle.Render("1.") + " 아래 URL을 브라우저에서 열어주세요\n")
		content.WriteString("   " + urlStyle.Render(m.authURL) + "\n\n")
		content.WriteString(headerStyle.Render("2.") + " 로그인 후 표시된 PIN 번호를 입력하세요\n\n")
		content.WriteString("   PIN: " + m.pinInput.View() + "\n\n")
		content.WriteString(helpStyle.Render("[Enter] 인증  [Esc] 취소") + "\n")

	case authStateExchangingToken:
		content.WriteString("액세스 토큰을 교환하는 중...\n")

	case authStateSuccess:
		content.WriteString(successStyle.Render("[완료] 인증이 완료되었습니다!") + "\n")

	case authStateError:
		content.WriteString(errorStyle.Render("[오류] "+m.errMsg) + "\n\n")
		content.WriteString(helpStyle.Render("[Esc] 종료") + "\n")
	}

	return content.String()
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
