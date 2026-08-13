package api

import (
	"encoding/json"
	"testing"
)

func TestInOutAccounts_UnmarshalArray(t *testing.T) {
	raw := []byte(`[{"account_id":"x2","in":10,"out":3,"margin":7}]`)
	var a InOutAccounts
	if err := json.Unmarshal(raw, &a); err != nil {
		t.Fatal(err)
	}
	if len(a) != 1 || a[0].AccountID != "x2" || a[0].In != 10 {
		t.Fatalf("unexpected: %+v", a)
	}
}

func TestInOutAccounts_UnmarshalObject(t *testing.T) {
	// 실제 Whooing 응답: accounts가 배열이 아니라 객체
	raw := []byte(`{"x1":{"in":100,"out":40,"margin":60},"x3":{"in":0,"out":5,"margin":-5}}`)
	var a InOutAccounts
	if err := json.Unmarshal(raw, &a); err != nil {
		t.Fatal(err)
	}
	if len(a) != 2 {
		t.Fatalf("len=%d want 2", len(a))
	}
	if a[0].AccountID != "x1" || a[0].Margin != 60 {
		t.Fatalf("first: %+v", a[0])
	}
	if a[1].AccountID != "x3" || a[1].Out != 5 {
		t.Fatalf("second: %+v", a[1])
	}
}

func TestInOutResponse_UnmarshalObjectAccounts(t *testing.T) {
	raw := []byte(`{
		"assets": {
			"total": {"in":100,"out":40,"margin":60},
			"accounts": {"x1":{"in":100,"out":40,"margin":60}}
		},
		"liabilities": {
			"total": {"in":0,"out":0,"margin":0},
			"accounts": {}
		}
	}`)
	var resp InOutResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Assets.Total.In != 100 {
		t.Fatalf("total in=%d", resp.Assets.Total.In)
	}
	if len(resp.Assets.Accounts) != 1 || resp.Assets.Accounts[0].AccountID != "x1" {
		t.Fatalf("assets accounts: %+v", resp.Assets.Accounts)
	}
}
