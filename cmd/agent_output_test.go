package cmd

import (
	"testing"

	"whoo-cli/api"
)

func TestFlattenAccounts(t *testing.T) {
	accounts := &api.AccountsMap{
		Assets: map[string]api.AccountDetail{
			"x10": {Title: "열 번째", Category: "normal"},
			"x2":  {Title: "두 번째", Category: "normal"},
		},
		Expenses: map[string]api.AccountDetail{
			"x3": {Title: "비용", Category: "floating"},
		},
	}

	rows := flattenAccounts(accounts, "s1", "")
	if len(rows) != 3 {
		t.Fatalf("len(rows) = %d, want 3", len(rows))
	}
	if rows[0].AccountID != "x2" || rows[1].AccountID != "x10" {
		t.Errorf("asset order = %s, %s; want x2, x10", rows[0].AccountID, rows[1].AccountID)
	}
	for _, row := range rows {
		if row.SectionID != "s1" {
			t.Errorf("section_id = %q, want s1", row.SectionID)
		}
	}
}

func TestFindAccountType(t *testing.T) {
	accounts := &api.AccountsMap{
		Assets: map[string]api.AccountDetail{"x2": {Title: "통장"}},
	}
	accountType, ok := findAccountType(accounts, "x2")
	if !ok || accountType != "assets" {
		t.Errorf("findAccountType() = %q, %v; want assets, true", accountType, ok)
	}
}

func TestParseFlatEntries(t *testing.T) {
	data := []byte(`{
		"code": 200,
		"results": {
			"rows": [{
				"entry_id": 1,
				"entry_date": 20260916.0001,
				"l_account": "expenses",
				"l_account_id": "x1",
				"r_account": "assets",
				"r_account_id": "x2",
				"money": 1000,
				"item": "커피",
				"memo": ""
			}]
		}
	}`)

	rows, err := parseFlatEntries(data, "s1")
	if err != nil {
		t.Fatalf("parseFlatEntries() error = %v", err)
	}
	if len(rows) != 1 || rows[0].SectionID != "s1" || rows[0].RAccountID != "x2" {
		t.Errorf("rows = %+v", rows)
	}
}
