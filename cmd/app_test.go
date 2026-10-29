package cmd

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"whoo-cli/config"
)

func TestAppInitialState(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cfg  config.Config
		want appState
	}{
		{name: "unauthenticated", want: stateAuth},
		{name: "partial authentication", cfg: config.Config{Token: "t"}, want: stateAuth},
		{name: "choose section", cfg: config.Config{Token: "t", TokenSecret: "s"}, want: stateSection},
		{
			name: "ready", cfg: config.Config{Token: "t", TokenSecret: "s", SectionID: "s1"},
			want: stateMenu,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := newAppModel(&tt.cfg)
			m.Init()
			if m.state != tt.want || m.View() == "" {
				t.Fatalf("initial state = %v, want %v; view = %q", m.state, tt.want, m.View())
			}
		})
	}
}

func TestAppAuthenticationTransitions(t *testing.T) {
	t.Parallel()
	m := newAppModel(&config.Config{})
	m.Init()
	authenticated := &config.Config{Token: "t", TokenSecret: "s"}
	m.Update(authCompleteMsg{cfg: authenticated})
	if m.state != stateSection || m.cfg != authenticated || m.authModel != nil || m.sectionModel == nil {
		t.Fatal("authentication completion did not switch to section selection")
	}
	selected := &config.Config{Token: "t", TokenSecret: "s", SectionID: "s1"}
	m.Update(sectionSelectedMsg{cfg: selected})
	if m.state != stateMenu || m.cfg != selected || m.sectionModel != nil || m.menuModel == nil {
		t.Fatal("section selection did not switch to menu")
	}
}

func TestAppMenuNavigation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		shortcut rune
		index    int
		state    appState
	}{
		{name: "transactions", shortcut: '1', index: 0, state: stateTransactions},
		{name: "entry", shortcut: '2', index: 1, state: stateEntry},
		{name: "balance", shortcut: '3', index: 2, state: stateBalance},
		{name: "frequent", shortcut: '4', index: 3, state: stateFrequent},
		{name: "monthly", shortcut: '5', index: 4, state: stateMonthly},
		{name: "accounts", shortcut: '6', index: 5, state: stateAccountManage},
		{name: "flow", shortcut: '7', index: 6, state: stateFlow},
		{name: "card", shortcut: '8', index: 7, state: stateCard},
		{name: "budget", shortcut: '9', index: 8, state: stateBudget},
		{name: "sections", shortcut: 'a', index: 9, state: stateSectionHub},
		{name: "user", shortcut: 'b', index: 10, state: stateUserInfo},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := newAppModel(&config.Config{Token: "t", TokenSecret: "s", SectionID: "s1"})
			m.Init()
			_, selectCmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.shortcut}})
			if selectCmd == nil {
				t.Fatal("shortcut did not emit menu selection")
			}
			msg, ok := selectCmd().(menuSelectionMsg)
			if !ok || msg.selection != tt.index {
				t.Fatalf("selection = %+v, want %d", msg, tt.index)
			}
			// 반환된 API 명령은 실행하지 않고 화면 전환 메시지만 전달한다.
			m.Update(msg)
			if m.state != tt.state || m.menuModel != nil {
				t.Fatalf("state = %v, want %v; previous menu retained = %v", m.state, tt.state, m.menuModel != nil)
			}
			m.Update(backToMenuMsg{})
			if m.state != stateMenu || m.menuModel.cursorIndex() != tt.index {
				t.Fatalf("return to menu lost cursor %d", tt.index)
			}
		})
	}
}

func TestAppQuit(t *testing.T) {
	t.Parallel()
	m := newAppModel(&config.Config{})
	m.Init()
	_, quit := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if quit == nil {
		t.Fatal("ctrl+c did not emit quit command")
	}
	if _, ok := quit().(tea.QuitMsg); !ok || m.state != stateExit || m.View() != "" {
		t.Fatal("ctrl+c did not exit and clear the view")
	}
}
