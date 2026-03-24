// cmd/balance_sub.go
// 자산부채 현황 - bubbletea 서브 모델

package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"whooing-cli/api"
	"whooing-cli/config"
)

// balanceSubModel은 자산부채 조회 서브 모델
type balanceSubModel struct {
	cfg         *config.Config
	client      *api.WhooingClient
	bsResp      *api.BSResponse
	accountsMap *api.AccountsMap
	endDate     string
	loading     bool
	err         error
	width       int
	height      int
}

// newBalanceSubModel은 새로운 자산부채 모델을 생성
func newBalanceSubModel(cfg *config.Config) *balanceSubModel {
	return &balanceSubModel{
		cfg:     cfg,
		client:  NewClient(cfg),
		endDate: time.Now().Format("20060102"),
		loading: true,
	}
}

func (m *balanceSubModel) Init() tea.Cmd {
	return m.fetchBalance()
}

func (m *balanceSubModel) fetchBalance() tea.Cmd {
	return func() tea.Msg {
		// 자산부채 조회
		bsResp, err := m.client.GetBS(m.cfg.SectionID, m.endDate)
		if err != nil {
			return balanceErrMsg{err: err}
		}

		// 계정명 조회
		accountsMap, err := m.client.GetAccountsMap(m.cfg.SectionID)
		if err != nil {
			return balanceErrMsg{err: err}
		}

		return balanceLoadedMsg{
			bsResp:      bsResp,
			accountsMap: accountsMap,
		}
	}
}

type balanceLoadedMsg struct {
	bsResp      *api.BSResponse
	accountsMap *api.AccountsMap
}

type balanceErrMsg struct {
	err error
}

func (m *balanceSubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "Q", "esc":
			return m, func() tea.Msg { return backToMenuMsg{} }
		case "enter":
			return m, func() tea.Msg { return backToMenuMsg{} }
		}

	case balanceLoadedMsg:
		m.bsResp = msg.bsResp
		m.accountsMap = msg.accountsMap
		m.loading = false

	case balanceErrMsg:
		m.err = msg.err
		m.loading = false
	}

	return m, nil
}

func (m *balanceSubModel) View() string {
	if m.loading {
		return m.renderLoading()
	}

	if m.err != nil {
		return m.renderError()
	}

	return m.renderBalance()
}

func (m *balanceSubModel) renderLoading() string {
	return titleStyle.Render("자산부채 현황") + "\n\n자산부채 데이터를 불러오는 중...\n"
}

func (m *balanceSubModel) renderError() string {
	return titleStyle.Render("자산부채 현황") + "\n\n" +
		errorStyle.Render("[오류] "+m.err.Error()) + "\n\n" +
		helpStyle.Render("[Enter] 메뉴로 돌아가기")
}

func (m *balanceSubModel) renderBalance() string {
	var content string

	// 헤더
	content += titleStyle.Render("자산부채 현황") + "\n"
	content += fmt.Sprintf("(%s 기준)\n\n", FormatDate(m.endDate))

	// 데이터가 없는 경우
	if len(m.bsResp.Assets.Accounts) == 0 && len(m.bsResp.Liabilities.Accounts) == 0 {
		return content + "자산부채 데이터가 없습니다\n\n" +
			helpStyle.Render("[Enter] 메뉴로 돌아가기  [q] 종료")
	}

	// 테이블 작성
	var tableBuf strings.Builder
	w := tabwriter.NewWriter(&tableBuf, 0, 0, 2, ' ', 0)

	// 자산 출력
	if len(m.bsResp.Assets.Accounts) > 0 {
		fmt.Fprintln(w, "[자산]")
		m.printBSGroup(w, m.bsResp.Assets, "assets")
		fmt.Fprintf(w, "  소계\t%s\n", FormatMoney(m.bsResp.Assets.Total))
		fmt.Fprintln(w)
	}

	// 부채 출력
	if len(m.bsResp.Liabilities.Accounts) > 0 {
		fmt.Fprintln(w, "[부채]")
		m.printBSGroup(w, m.bsResp.Liabilities, "liabilities")
		fmt.Fprintf(w, "  소계\t%s\n", FormatMoney(m.bsResp.Liabilities.Total))
		fmt.Fprintln(w)
	}

	w.Flush()
	content += tableBuf.String()

	// 순자산
	netWorth := m.bsResp.Assets.Total - m.bsResp.Liabilities.Total
	content += fmt.Sprintf("순자산: %s\n", FormatMoney(netWorth))

	// 도움말
	content += "\n" + helpStyle.Render("[Enter] 메뉴로 돌아가기  [q] 종료")

	return content
}

func (m *balanceSubModel) printBSGroup(w *tabwriter.Writer, group api.BSGroup, accountType string) {
	for _, item := range group.Accounts {
		if item.Money == 0 {
			continue
		}
		title := m.accountsMap.GetTitle(accountType, item.AccountID)
		fmt.Fprintf(w, "  %s\t%s\n", title, FormatMoney(item.Money))
	}
}
