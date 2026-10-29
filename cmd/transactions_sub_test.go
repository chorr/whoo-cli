package cmd

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"whoo-cli/api"
	"whoo-cli/config"
)

type stubTransactionsClient struct {
	t        *testing.T
	entries  func(string, string, string, int, string) ([]api.Entry, error)
	accounts func(string) (*api.AccountsMap, error)
	delete   func(string, int) error
}

func (s *stubTransactionsClient) GetEntries(section, start, end string, limit int, cursor string) ([]api.Entry, error) {
	s.t.Helper()
	if s.entries == nil {
		s.t.Fatal("unexpected entries request")
		return nil, nil
	}
	return s.entries(section, start, end, limit, cursor)
}

func (s *stubTransactionsClient) GetAccountsMap(section string) (*api.AccountsMap, error) {
	s.t.Helper()
	if s.accounts == nil {
		s.t.Fatal("unexpected accounts request")
		return nil, nil
	}
	return s.accounts(section)
}

func (s *stubTransactionsClient) DeleteEntry(section string, id int) error {
	s.t.Helper()
	if s.delete == nil {
		s.t.Fatal("unexpected delete request")
		return nil
	}
	return s.delete(section, id)
}

func transactionTestPage(count int) []api.Entry {
	entries := make([]api.Entry, count)
	for i := range entries {
		entries[i] = api.Entry{
			EntryID: 100 - i, EntryDate: api.FlexibleString(fmt.Sprintf("20261029.%04d", 100-i)),
			LAccount: "expenses", LAccountID: "e1", RAccount: "assets", RAccountID: "a1", Money: 1200, Item: "커피",
		}
	}
	return entries
}

func TestTransactionsLoad(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"full page", "empty page", "entries failure", "accounts failure"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			sentinel := errors.New("load failed")
			m := newTransactionsSubModel(&config.Config{SectionID: "s1"})
			m.startDate, m.endDate = "20261001", "20261029"
			rows := transactionTestPage(pageSize)
			if name == "empty page" {
				rows = []api.Entry{}
			}
			accountsCalls := 0
			m.client = &stubTransactionsClient{
				t: t,
				entries: func(section, start, end string, limit int, cursor string) ([]api.Entry, error) {
					if section != "s1" || start != "20261001" || end != "20261029" || limit != pageSize || cursor != "" {
						t.Fatalf("unexpected initial query: %q, %q, %q, %d, %q", section, start, end, limit, cursor)
					}
					if name == "entries failure" {
						return nil, sentinel
					}
					return rows, nil
				},
				accounts: func(section string) (*api.AccountsMap, error) {
					accountsCalls++
					if name == "entries failure" || section != "s1" {
						t.Fatal("unexpected accounts request")
					}
					if name == "accounts failure" {
						return nil, sentinel
					}
					return entryTestAccounts(), nil
				},
			}
			m.Update(m.Init()())
			if m.loading || m.loadingMore {
				t.Fatal("load did not clear loading state")
			}
			if strings.HasSuffix(name, "failure") {
				if !errors.Is(m.err, sentinel) || !strings.Contains(m.View(), sentinel.Error()) {
					t.Fatalf("load failure lost: %v, view = %s", m.err, m.View())
				}
				if name == "entries failure" && accountsCalls != 0 {
					t.Fatal("accounts queried after entries failure")
				}
				return
			}
			if m.err != nil || !m.tableReady || !reflect.DeepEqual(m.entries, rows) || m.hasMore != (len(rows) == pageSize) {
				t.Fatalf("initial result: rows = %d, hasMore = %v, error = %v", len(m.entries), m.hasMore, m.err)
			}
			if len(rows) > 0 && (!strings.Contains(m.View(), "식비") || !strings.Contains(m.View(), "현금")) {
				t.Fatal("account titles missing from table")
			}
		})
	}
}

func TestTransactionsPagination(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"last partial page", "empty last page", "failed page"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := newTransactionsSubModel(&config.Config{SectionID: "s1"})
			first := transactionTestPage(pageSize)
			m.Update(transactionsLoadedMsg{entries: first, accountsMap: entryTestAccounts(), hasMore: true})
			m.table.SetCursor(5)
			calls := 0
			sentinel := errors.New("next page failed")
			more := []api.Entry{}
			if name == "last partial page" {
				more = []api.Entry{{EntryID: 20, EntryDate: "20261028.0002"}, {EntryID: 19, EntryDate: "20261028.0001"}}
			}
			m.client = &stubTransactionsClient{
				t: t,
				entries: func(section, start, end string, limit int, cursor string) ([]api.Entry, error) {
					calls++
					wantCursor := first[len(first)-1].EntryDate.String()
					if section != "s1" || limit != pageSize || cursor != wantCursor {
						t.Fatalf("next query section = %q, limit = %d, cursor = %q; want %q", section, limit, cursor, wantCursor)
					}
					if name == "failed page" {
						return nil, sentinel
					}
					return more, nil
				},
			}
			_, fetch := m.Update(tea.KeyMsg{Type: tea.KeyTab})
			if fetch == nil || !m.loadingMore {
				t.Fatal("Tab did not request more entries")
			}
			if _, duplicate := m.Update(tea.KeyMsg{Type: tea.KeyTab}); duplicate != nil {
				t.Fatal("Tab issued duplicate page request")
			}
			m.Update(fetch())
			if calls != 1 || m.loadingMore || m.table.Cursor() != 5 {
				t.Fatalf("pagination state: calls = %d, loadingMore = %v, cursor = %d", calls, m.loadingMore, m.table.Cursor())
			}
			if name == "failed page" {
				if !errors.Is(m.err, sentinel) || !reflect.DeepEqual(m.entries, first) {
					t.Fatal("failed page replaced existing entries or lost error")
				}
				return
			}
			want := append(append([]api.Entry{}, first...), more...)
			if !reflect.DeepEqual(m.entries, want) || m.hasMore {
				t.Fatal("last page not appended correctly")
			}
			if _, fetch := m.Update(tea.KeyMsg{Type: tea.KeyTab}); fetch != nil {
				t.Fatal("request issued after last page")
			}
		})
	}
}

func TestTransactionsDeleteFlow(t *testing.T) {
	t.Parallel()
	for _, fail := range []bool{false, true} {
		name := "success refreshes list"
		if fail {
			name = "failure keeps list"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			m := newTransactionsSubModel(&config.Config{SectionID: "s1"})
			rows := transactionTestPage(2)
			m.Update(transactionsLoadedMsg{entries: rows, accountsMap: entryTestAccounts()})
			m.table.SetCursor(1)
			sentinel := errors.New("delete failed")
			deleteCalls, refreshCalls := 0, 0
			m.client = &stubTransactionsClient{
				t: t,
				delete: func(section string, id int) error {
					deleteCalls++
					if section != "s1" || id != rows[1].EntryID {
						t.Fatalf("wrong selected entry: section = %q, id = %d", section, id)
					}
					if fail {
						return sentinel
					}
					return nil
				},
				entries: func(string, string, string, int, string) ([]api.Entry, error) {
					refreshCalls++
					return rows[:1], nil
				},
				accounts: func(string) (*api.AccountsMap, error) { return entryTestAccounts(), nil },
			}
			m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
			if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil || !m.confirmDelete {
				t.Fatal("Enter must not confirm deletion")
			}
			_, remove := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
			if remove == nil {
				t.Fatal("Y did not emit delete command")
			}
			_, refresh := m.Update(remove())
			if deleteCalls != 1 || m.deleting {
				t.Fatal("delete did not complete")
			}
			if fail {
				if refresh != nil || refreshCalls != 0 || !errors.Is(m.deleteErr, sentinel) || len(m.entries) != 2 {
					t.Fatalf("delete failure state lost: err = %v, rows = %d", m.deleteErr, len(m.entries))
				}
				if !strings.Contains(m.View(), sentinel.Error()) {
					t.Fatal("delete error not shown")
				}
				if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}); cmd != nil || m.deleteErr != nil {
					t.Fatal("error dismissal triggered edit action")
				}
				return
			}
			if refresh == nil || !m.loading || m.tableReady {
				t.Fatal("deleted entry remains actionable while refreshing")
			}
			if _, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}); cmd != nil || m.confirmDelete {
				t.Fatal("delete reopened on stale list")
			}
			m.Update(refresh())
			refreshed := refreshCalls == 1 && !m.loading && m.tableReady
			if !refreshed || len(m.entries) != 1 || m.entries[0].EntryID != rows[0].EntryID {
				t.Fatal("successful delete did not refresh entries")
			}
		})
	}
}

func TestTransactionsPreventsActionsDuringDelete(t *testing.T) {
	t.Parallel()
	m := newTransactionsSubModel(&config.Config{SectionID: "s1"})
	m.Update(transactionsLoadedMsg{
		entries: []api.Entry{{EntryID: 42, Item: "커피"}}, accountsMap: entryTestAccounts(),
	})
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	_, first := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if first == nil || !m.deleting {
		t.Fatal("delete confirmation did not emit command")
	}
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'d'}}, {Type: tea.KeyRunes, Runes: []rune{'y'}},
		{Type: tea.KeyRunes, Runes: []rune{'e'}}, {Type: tea.KeyEsc},
	} {
		if _, next := m.Update(key); next != nil || m.confirmDelete {
			t.Fatalf("key %q emitted an action during pending delete", key.String())
		}
	}
}

func TestCalcDateRangeAt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		now    string
		preset int
		start  string
		end    string
	}{
		{name: "current month", now: "2026-03-31", preset: 0, start: "20260301", end: "20260331"},
		{name: "month end", now: "2026-03-31", preset: 1, start: "20260201", end: "20260228"},
		{name: "leap year", now: "2024-03-31", preset: 1, start: "20240201", end: "20240229"},
		{name: "year boundary", now: "2026-01-31", preset: 1, start: "20251201", end: "20251231"},
		{name: "three months", now: "2026-03-31", preset: 2, start: "20251201", end: "20260331"},
		{name: "custom range", now: "2026-03-31", preset: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			now, err := time.Parse("2006-01-02", tt.now)
			if err != nil {
				t.Fatal(err)
			}
			start, end := calcDateRangeAt(tt.preset, now)
			if start != tt.start || end != tt.end {
				t.Fatalf("range = %s~%s, want %s~%s", start, end, tt.start, tt.end)
			}
		})
	}
}
