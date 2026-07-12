package cmd

import (
	"testing"
)

func TestExtractWhooingItemMaps_MapShape(t *testing.T) {
	// 실제 monthly API 형태: slot1이 배열이 아니라 item_id 맵
	raw := []byte(`{
		"code": 200,
		"results": {
			"count": 2,
			"slot1": {
				"m113769": {
					"item_id": "m113769",
					"item": "실손보험",
					"money": 25776,
					"l_account": "expenses",
					"l_account_id": "x57",
					"r_account": "assets",
					"r_account_id": "x131",
					"skip_holiday": "none",
					"pay_date": 13,
					"due_date": "20260713",
					"d_day": 1,
					"paid_date": ""
				},
				"m208775": {
					"item_id": "m208775",
					"item": "운전자보험",
					"money": 12391,
					"l_account": "expenses",
					"l_account_id": "x51",
					"r_account": "assets",
					"r_account_id": "x131",
					"pay_date": 13,
					"due_date": "20260713",
					"d_day": 1,
					"paid_date": ""
				}
			}
		}
	}`)
	items := parseMonthlyItemsFromRaw(raw)
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	ids := map[string]bool{}
	for _, it := range items {
		ids[it.ID] = true
		if it.Item == "" {
			t.Fatal("empty item name")
		}
		if it.PayDate != 13 {
			t.Fatalf("pay_date=%d", it.PayDate)
		}
	}
	if !ids["m113769"] || !ids["m208775"] {
		t.Fatalf("ids=%v", ids)
	}
}

func TestExtractWhooingItemMaps_ArrayShape(t *testing.T) {
	// 문서 예시 형태 (배열)
	raw := []byte(`{
		"code": 200,
		"results": {
			"count": 1,
			"slot1": [
				{
					"item_id": "m4",
					"item": "통신비",
					"money": 79200,
					"l_account": "expenses",
					"l_account_id": "x12",
					"r_account": "assets",
					"r_account_id": "x5",
					"skip_holiday": "after",
					"pay_date": 27,
					"due_date": "20120327",
					"d_day": 1,
					"paid_date": "20120227"
				}
			]
		}
	}`)
	items := parseMonthlyItemsFromRaw(raw)
	if len(items) != 1 {
		t.Fatalf("got %d items", len(items))
	}
	if items[0].ID != "m4" || items[0].Item != "통신비" {
		t.Fatalf("%+v", items[0])
	}
	if items[0].PaidDate != "20120227" {
		t.Fatalf("paid_date=%q", items[0].PaidDate)
	}
}

func TestParseFrequentItemsMapShape(t *testing.T) {
	// 슬롯 단건: results가 바로 id→객체 맵
	raw := []byte(`{
		"code": 200,
		"results": {
			"f6": {
				"item_id": "f6",
				"item": "생필품",
				"money": 0,
				"l_account": "expenses",
				"l_account_id": "x53",
				"r_account": "liabilities",
				"r_account_id": "x21"
			}
		}
	}`)
	items := parseFrequentItemsFromRaw(raw)
	if len(items) != 1 || items[0].ID != "f6" {
		t.Fatalf("%+v", items)
	}
	fi := findFrequentItemByID(raw, "f6")
	if fi == nil || fi.Item != "생필품" {
		t.Fatalf("%+v", fi)
	}
}

func TestOldParserWouldMissMapShape(t *testing.T) {
	// 회귀: item_id 맵 응답에서 0건이 되면 안 됨
	raw := []byte(`{"code":200,"results":{"slot1":{"m1":{"item_id":"m1","item":"a","money":1,"pay_date":1,"l_account":"expenses","l_account_id":"x1","r_account":"assets","r_account_id":"x2"}}}}`)
	if n := len(parseMonthlyItemsFromRaw(raw)); n != 1 {
		t.Fatalf("map shape parse count=%d", n)
	}
}
