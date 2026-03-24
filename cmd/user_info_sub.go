// cmd/user_info_sub.go
// 사용자 정보 - bubbletea 서브 모델
// GetUser() raw bytes를 재활용하여 TUI 표시

package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"whooing-cli/api"
	"whooing-cli/config"
)

// userInfoSubModel은 사용자 정보 서브 모델
type userInfoSubModel struct {
	cfg     *config.Config
	client  *api.WhooingClient
	data    map[string]interface{}
	loading bool
	err     error
}

// newUserInfoSubModel은 새로운 사용자 정보 모델을 생성
func newUserInfoSubModel(cfg *config.Config) *userInfoSubModel {
	return &userInfoSubModel{
		cfg:     cfg,
		client:  NewClient(cfg),
		loading: true,
	}
}

func (m *userInfoSubModel) Init() tea.Cmd {
	return m.fetchUserInfo()
}

// fetchUserInfo는 GetUser() raw bytes를 재활용하여 데이터 로드
func (m *userInfoSubModel) fetchUserInfo() tea.Cmd {
	return func() tea.Msg {
		// CLI와 동일한 GetUser() 호출 (raw []byte 반환)
		raw, err := m.client.GetUser()
		if err != nil {
			return userInfoErrMsg{err: err}
		}

		// TUI 측에서 results만 파싱
		var resp struct {
			Code    int                    `json:"code"`
			Results map[string]interface{} `json:"results"`
		}
		if err := json.Unmarshal(raw, &resp); err != nil {
			return userInfoErrMsg{err: fmt.Errorf("응답 파싱 실패: %w", err)}
		}
		if resp.Code != 200 {
			return userInfoErrMsg{err: fmt.Errorf("API 오류 (code=%d)", resp.Code)}
		}

		return userInfoLoadedMsg{data: resp.Results}
	}
}

type userInfoLoadedMsg struct {
	data map[string]interface{}
}

type userInfoErrMsg struct {
	err error
}

func (m *userInfoSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "Q", "esc", "enter":
			return m, func() tea.Msg { return backToMenuMsg{} }
		}

	case userInfoLoadedMsg:
		m.data = msg.data
		m.loading = false

	case userInfoErrMsg:
		m.err = msg.err
		m.loading = false
	}

	return m, nil
}

func (m *userInfoSubModel) View() string {
	if m.loading {
		return titleStyle.Render("사용자 정보") + "\n\n사용자 정보를 불러오는 중...\n"
	}

	if m.err != nil {
		return titleStyle.Render("사용자 정보") + "\n\n" +
			errorStyle.Render("[오류] "+m.err.Error()) + "\n\n" +
			helpStyle.Render("[Enter] 메뉴로 돌아가기")
	}

	return m.renderUserInfo()
}

func (m *userInfoSubModel) renderUserInfo() string {
	labelStyle := lipgloss.NewStyle().Width(16).Foreground(lipgloss.Color("8"))
	valueStyle := lipgloss.NewStyle().Bold(true)

	var content string
	content += titleStyle.Render("사용자 정보") + "\n\n"

	// 표시할 필드 목록 (순서 보장)
	fields := []struct {
		label string
		key   string
		fmt   func(interface{}) string
	}{
		{"사용자명", "username", fmtString},
		{"언어", "language", fmtString},
		{"국가", "country", fmtString},
		{"타임존", "timezone", fmtString},
		{"통화", "currency", fmtString},
		{"레벨", "level", fmtString},
		{"마일리지", "mileage", fmtNumber},
		{"마지막 접속", "last_login_timestamp", fmtTimestamp},
		{"가입일", "created_timestamp", fmtTimestamp},
		{"유료 만료일", "expire", fmtTimestamp},
	}

	for _, f := range fields {
		val, ok := m.data[f.key]
		if !ok {
			continue
		}
		content += labelStyle.Render(f.label) + valueStyle.Render(f.fmt(val)) + "\n"
	}

	content += "\n" + helpStyle.Render("[Enter/Esc/q] 메뉴로 돌아가기")
	return content
}

// 포맷 헬퍼

func fmtString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func fmtNumber(v interface{}) string {
	if n, ok := v.(float64); ok {
		return FormatMoney(n)
	}
	return fmt.Sprintf("%v", v)
}

func fmtTimestamp(v interface{}) string {
	if n, ok := v.(float64); ok {
		ts := int64(n)
		if ts <= 0 {
			return "-"
		}
		return time.Unix(ts, 0).Format("2006-01-02 15:04")
	}
	return fmt.Sprintf("%v", v)
}
