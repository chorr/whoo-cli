package cmd

import (
	"testing"
)

func TestParseCardTableRowsSingle(t *testing.T) {
	raw := []byte(`{
		"code": 200,
		"results": {
			"rows_type": "month",
			"rows": {
				"202607": {
					"account_id": "x21",
					"money": 4588656,
					"start_use_date": 20260601,
					"end_use_date": 20260630,
					"pay_date": 12
				},
				"202606": {
					"account_id": "x21",
					"money": 4446215,
					"start_use_date": 20260501,
					"end_use_date": 20260531,
					"pay_date": 12
				}
			}
		}
	}`)
	rows := parseCardTableRows(raw)
	if len(rows) != 2 {
		t.Fatalf("rows=%d", len(rows))
	}
	// 최신순
	if rows[0].YM != "202607" || rows[0].Money != 4588656 {
		t.Fatalf("%+v", rows[0])
	}
	if rows[0].StartUseDate != 20260601 || rows[0].EndUseDate != 20260630 {
		t.Fatalf("use dates %+v", rows[0])
	}
}

func TestParseCardTableRowsWithTotal(t *testing.T) {
	raw := []byte(`{
		"code": 200,
		"results": {
			"rows": {
				"202606": {
					"date": 202606,
					"total": 1000,
					"accounts": {
						"x21": {"account_id":"x21","money":600,"start_use_date":20260501,"end_use_date":20260531}
					}
				}
			}
		}
	}`)
	rows := parseCardTableRows(raw)
	if len(rows) != 1 || rows[0].Money != 1000 {
		t.Fatalf("%+v", rows)
	}
}

func TestParseCardDrillEntriesRowsWrapper(t *testing.T) {
	raw := []byte(`{
		"code": 200,
		"results": {
			"rows": [
				{
					"entry_id": 1,
					"entry_date": "20260615.0001",
					"l_account": "expenses",
					"l_account_id": "x50",
					"r_account": "liabilities",
					"r_account_id": "x21",
					"item": "커피",
					"money": 5000,
					"memo": ""
				},
				{
					"entry_id": 2,
					"entry_date": "20260616.0001",
					"l_account": "expenses",
					"l_account_id": "x50",
					"r_account": "liabilities",
					"r_account_id": "x99",
					"item": "다른카드",
					"money": 1000,
					"memo": ""
				}
			],
			"reports": []
		}
	}`)
	entries := parseCardDrillEntries(raw, "x21")
	if len(entries) != 1 {
		t.Fatalf("want 1 got %d", len(entries))
	}
	if entries[0].Item != "커피" || entries[0].Money != 5000 {
		t.Fatalf("%+v", entries[0])
	}
	if entries[0].Date != "20260615" {
		t.Fatalf("date=%s", entries[0].Date)
	}
}
