package api

import (
	"encoding/json"
	"testing"
)

func TestFlexibleStringUnmarshal(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
	}{
		{"string", `"20260310.0007"`, "20260310.0007"},
		{"number", `20110812.0001`, "20110812.0001"},
		{"int", `20110812`, "20110812"},
		{"null", `null`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var s FlexibleString
			if err := json.Unmarshal([]byte(tc.raw), &s); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if string(s) != tc.want {
				t.Fatalf("got %q want %q", s, tc.want)
			}
		})
	}
}

func TestEntryDateOnlyFromNumber(t *testing.T) {
	raw := []byte(`{
		"entry_id": 1,
		"entry_date": 20110812.0001,
		"l_account": "expenses",
		"l_account_id": "x1",
		"r_account": "assets",
		"r_account_id": "x2",
		"money": 1000,
		"item": "테스트",
		"memo": ""
	}`)
	var e Entry
	if err := json.Unmarshal(raw, &e); err != nil {
		t.Fatalf("entry unmarshal: %v", err)
	}
	if e.DateOnly() != "20110812" {
		t.Fatalf("DateOnly=%q", e.DateOnly())
	}
}

func TestParseEntryObjectResponse(t *testing.T) {
	c := &WhooingClient{}
	raw := []byte(`{
		"code": 200,
		"message": "",
		"results": {
			"entry_id": 1352827,
			"entry_date": 20110812.0001,
			"l_account": "expenses",
			"l_account_id": "x20",
			"r_account": "assets",
			"r_account_id": "x4",
			"item": "후원",
			"money": 10000,
			"memo": "memo"
		}
	}`)
	entry, err := c.parseEntryObjectResponse(raw)
	if err != nil {
		t.Fatalf("parseEntryObjectResponse: %v", err)
	}
	if entry.EntryID != 1352827 {
		t.Fatalf("entry_id=%d", entry.EntryID)
	}
	if entry.DateOnly() != "20110812" {
		t.Fatalf("DateOnly=%q", entry.DateOnly())
	}
}

func TestParseEntryArrayResponseStillWorks(t *testing.T) {
	c := &WhooingClient{}
	raw := []byte(`{
		"code": 200,
		"results": [{
			"entry_id": 1,
			"entry_date": "20260101.0001",
			"l_account": "expenses",
			"l_account_id": "x1",
			"r_account": "assets",
			"r_account_id": "x2",
			"item": "a",
			"money": 100,
			"memo": ""
		}]
	}`)
	entry, err := c.parseEntryArrayResponse(raw)
	if err != nil {
		t.Fatalf("parseEntryArrayResponse: %v", err)
	}
	if entry.EntryID != 1 {
		t.Fatalf("entry_id=%d", entry.EntryID)
	}
}
