package cmd

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"whoo-cli/api"
	"whoo-cli/config"
)

type stubEntryClient struct {
	t      *testing.T
	load   func(string) (*api.AccountsMap, error)
	create func(string, api.Entry) (*api.Entry, error)
	update func(string, int, map[string]string) (*api.Entry, error)
}

func (s *stubEntryClient) GetAccountsMap(section string) (*api.AccountsMap, error) {
	s.t.Helper()
	if s.load == nil {
		s.t.Fatal("unexpected accounts request")
		return nil, nil
	}
	return s.load(section)
}

func (s *stubEntryClient) CreateEntry(
	section, date, left, leftID, right, rightID, item, memo string,
	money float64,
) (*api.Entry, error) {
	s.t.Helper()
	if s.create == nil {
		s.t.Fatal("unexpected create request")
		return nil, nil
	}
	return s.create(section, api.Entry{
		EntryDate: api.FlexibleString(date), LAccount: left, LAccountID: leftID,
		RAccount: right, RAccountID: rightID, Item: item, Memo: memo, Money: money,
	})
}

func (s *stubEntryClient) UpdateEntry(section string, id int, fields map[string]string) (*api.Entry, error) {
	s.t.Helper()
	if s.update == nil {
		s.t.Fatal("unexpected update request")
		return nil, nil
	}
	return s.update(section, id, fields)
}

func entryTestOriginal() api.Entry {
	return api.Entry{
		EntryID: 42, EntryDate: "20261029.0001", LAccount: "expenses", LAccountID: "e1",
		RAccount: "assets", RAccountID: "a1", Money: 1200, Item: "커피", Memo: "기존 메모",
	}
}

func TestEntryCreateFlow(t *testing.T) {
	t.Parallel()
	calls := 0
	m := newEntrySubModel(&config.Config{SectionID: "s1"})
	m.date = "20261029"
	m.client = &stubEntryClient{
		t: t,
		load: func(section string) (*api.AccountsMap, error) {
			if section != "s1" {
				t.Fatalf("accounts section = %q", section)
			}
			return entryTestAccounts(), nil
		},
		create: func(section string, entry api.Entry) (*api.Entry, error) {
			calls++
			want := entryTestOriginal()
			want.EntryID, want.EntryDate, want.Memo = 0, "20261029", "점심 & 커피"
			if section != "s1" || !reflect.DeepEqual(entry, want) {
				t.Fatalf("create(%q) = %+v, want %+v", section, entry, want)
			}
			entry.EntryID = 42
			return &entry, nil
		},
	}
	m.Update(m.Init()())
	enter := tea.KeyMsg{Type: tea.KeyEnter}
	keys := []tea.KeyMsg{
		enter,
		{Type: tea.KeyRunes, Runes: []rune{'4'}}, {Type: tea.KeyRunes, Runes: []rune{'1'}},
		{Type: tea.KeyRunes, Runes: []rune{'1'}}, {Type: tea.KeyRunes, Runes: []rune{'1'}},
		{Type: tea.KeyRunes, Runes: []rune("1,200")}, enter,
		{Type: tea.KeyRunes, Runes: []rune("커피")}, enter,
		{Type: tea.KeyRunes, Runes: []rune("점심 & 커피")}, enter,
	}
	for _, key := range keys {
		m.Update(key)
	}
	if m.step != entryStepConfirm || calls != 0 || !strings.Contains(m.View(), "등록하시겠습니까") {
		t.Fatalf("form did not reach confirmation: step = %d, calls = %d", m.step, calls)
	}
	_, save := m.Update(enter)
	if save == nil {
		t.Fatal("missing save command")
	}
	m.Update(save())
	if calls != 1 || !m.done || m.saving || !strings.Contains(m.View(), "등록되었습니다") {
		t.Fatalf("save completion not shown: calls = %d, view = %s", calls, m.View())
	}
	_, leave := m.Update(enter)
	if leave == nil {
		t.Fatal("completion did not return to menu")
	}
	if _, ok := leave().(backToMenuMsg); !ok {
		t.Fatal("creation returned to wrong screen")
	}
}

func TestEntryEditFlow(t *testing.T) {
	t.Parallel()
	for _, changed := range []bool{false, true} {
		name := "unchanged skips API"
		if changed {
			name = "only changed fields"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			original := entryTestOriginal()
			m := newEntrySubModelForEdit(&config.Config{SectionID: "s1"}, original, entryTestAccounts())
			calls := 0
			m.client = &stubEntryClient{
				t: t,
				update: func(section string, id int, fields map[string]string) (*api.Entry, error) {
					calls++
					want := map[string]string{"money": "2300", "memo": ""}
					if section != "s1" || id != 42 || !reflect.DeepEqual(fields, want) {
						t.Fatalf("update(%q, %d, %v), want %v", section, id, fields, want)
					}
					return &original, nil
				},
			}
			if m.Init() != nil {
				t.Fatal("edit with loaded accounts should not load again")
			}
			if changed {
				m.money, m.memo = "2300", ""
			}
			m.step = entryStepConfirm
			_, save := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if save == nil {
				t.Fatal("missing edit command")
			}
			m.Update(save())
			wantCalls := 0
			if changed {
				wantCalls = 1
			}
			if calls != wantCalls || !m.done || m.saving || !strings.Contains(m.View(), "수정되었습니다") {
				t.Fatalf("edit completion: calls = %d, view = %s", calls, m.View())
			}
			if !strings.Contains(m.View(), "거래내역으로 돌아가기") {
				t.Fatal("edit completion advertises the wrong destination")
			}
			_, leave := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if leave == nil {
				t.Fatal("completion did not return to transactions")
			}
			if _, ok := leave().(backToTransactionsMsg); !ok {
				t.Fatal("edit returned to wrong screen")
			}
		})
	}
}

func TestEntrySaveFailure(t *testing.T) {
	t.Parallel()
	for _, edit := range []bool{false, true} {
		name := "create"
		if edit {
			name = "update"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			sentinel := errors.New("server rejected entry")
			m := entryTestModel()
			m.editMode, m.step = edit, entryStepConfirm
			original := entryTestOriginal()
			m.editEntry = &original
			m.money = "2300"
			m.client = &stubEntryClient{
				t:      t,
				create: func(string, api.Entry) (*api.Entry, error) { return nil, sentinel },
				update: func(string, int, map[string]string) (*api.Entry, error) { return nil, sentinel },
			}
			_, save := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			m.Update(save())
			if !errors.Is(m.err, sentinel) || m.done || m.saving || m.money != "2300" {
				t.Fatalf("save failure lost state: err = %v, done = %v, money = %q", m.err, m.done, m.money)
			}
			if !strings.Contains(m.View(), sentinel.Error()) || strings.Contains(m.View(), "[완료]") {
				t.Fatalf("failure displayed as success: %s", m.View())
			}
			_, leave := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if leave == nil || !m.cancelled {
				t.Fatal("failure screen did not provide exit")
			}
			_, transactions := leave().(backToTransactionsMsg)
			if transactions != edit {
				t.Fatal("failure returned to wrong screen")
			}
		})
	}
}

func TestEntryCancelBeforeSave(t *testing.T) {
	t.Parallel()
	for _, key := range []tea.KeyMsg{{Type: tea.KeyEsc}, {Type: tea.KeyRunes, Runes: []rune{'n'}}} {
		t.Run(key.String(), func(t *testing.T) {
			t.Parallel()
			m := entryTestModel()
			m.step = entryStepConfirm
			m.client = &stubEntryClient{t: t}
			_, cancel := m.Update(key)
			if cancel == nil || !m.cancelled || m.done || m.saving {
				t.Fatal("cancellation did not leave unsaved form")
			}
			if _, ok := cancel().(backToMenuMsg); !ok {
				t.Fatal("cancel triggered a save instead of navigation")
			}
		})
	}
}

func entryTestAccounts() *api.AccountsMap {
	return &api.AccountsMap{
		Expenses: map[string]api.AccountDetail{"e1": {Title: "식비"}},
		Assets:   map[string]api.AccountDetail{"a1": {Title: "현금"}},
	}
}

func entryTestModel() *entrySubModel {
	m := newEntrySubModel(&config.Config{SectionID: "s1"})
	m.Update(accountsLoadedMsg{accountsMap: entryTestAccounts()})
	m.date = "20261029"
	m.lAccountType, m.lAccountID, m.lAccountName = "expenses", "e1", "식비"
	m.rAccountType, m.rAccountID, m.rAccountName = "assets", "a1", "현금"
	m.money, m.item, m.memo = "1200", "커피", "기존 메모"
	return m
}

func TestEntryPreventsDuplicateSubmission(t *testing.T) {
	t.Parallel()
	for _, edit := range []bool{false, true} {
		name := "create"
		if edit {
			name = "update"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := entryTestModel()
			m.editMode = edit
			m.step = entryStepConfirm
			_, first := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if first == nil {
				t.Fatal("confirmation did not emit save command")
			}
			for _, key := range []tea.KeyMsg{
				{Type: tea.KeyEnter}, {Type: tea.KeyRunes, Runes: []rune{'y'}},
				{Type: tea.KeyCtrlU}, {Type: tea.KeyEsc},
			} {
				if _, next := m.Update(key); next != nil || m.step != entryStepConfirm || m.cancelled {
					t.Fatalf("key %q changed pending save or emitted another command", key.String())
				}
			}
			if !strings.Contains(m.View(), "저장 중") {
				t.Fatal("pending save is not visible")
			}
		})
	}
}

func TestEntryMoneyValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty"},
		{name: "zero", input: "0"},
		{name: "negative", input: "-1"},
		{name: "not numeric", input: "abc"},
		{name: "NaN", input: "NaN"},
		{name: "positive infinity", input: "+Inf"},
		{name: "negative infinity", input: "-Inf"},
		{name: "comma separated", input: " 1,200 ", want: "1200"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := entryTestModel()
			m.step, m.textInput = entryStepMoney, tt.input
			m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if tt.want == "" {
				if m.step != entryStepMoney || m.inputHint == "" || m.money != "1200" {
					t.Fatalf("invalid money advanced form: step = %d, hint = %q, money = %q", m.step, m.inputHint, m.money)
				}
				return
			}
			if m.step != entryStepItem || m.money != tt.want || m.inputHint != "" {
				t.Fatalf("valid money rejected: step = %d, money = %q, hint = %q", m.step, m.money, m.inputHint)
			}
		})
	}
}

func TestEntryDateValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "too short", input: "202610"},
		{name: "invalid day", input: "20260229"},
		{name: "invalid month", input: "20261301"},
		{name: "non numeric", input: "abcdefgh"},
		{name: "leap day", input: "20240229", want: "20240229"},
		{name: "trimmed date", input: " 20261029 ", want: "20261029"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := entryTestModel()
			m.step, m.textInput = entryStepDate, tt.input
			m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			if tt.want == "" {
				if m.step != entryStepDate || m.date != "20261029" || m.inputHint == "" || m.err != nil {
					t.Fatal("invalid date changed form or became an API error")
				}
				return
			}
			if m.step != entryStepLAccountType || m.date != tt.want || m.inputHint != "" {
				t.Fatalf("valid date rejected: date = %q, step = %d, hint = %q", m.date, m.step, m.inputHint)
			}
		})
	}
}

func TestEntryAccountChoicesRespectDate(t *testing.T) {
	t.Parallel()
	m := entryTestModel()
	m.accountsMap.Expenses["e0"] = api.AccountDetail{Title: "종료된 항목", CloseDate: float64(20261028)}
	m.accountsMap.Expenses["e2"] = api.AccountDetail{
		Title: "당일 항목", OpenDate: float64(20261029), CloseDate: "20261029",
	}
	m.accountsMap.Expenses["e3"] = api.AccountDetail{Title: "미래 항목", OpenDate: "20261030"}
	m.step = entryStepLAccountType
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	want := []accountChoice{{id: "e1", title: "식비"}, {id: "e2", title: "당일 항목"}}
	if m.step != entryStepLAccountID || !reflect.DeepEqual(m.choices, want) {
		t.Fatalf("account choices = %+v, want %+v", m.choices, want)
	}
}
