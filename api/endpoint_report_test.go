package api

import (
	"encoding/json"
	"testing"
)

func TestMergeReportResponses(t *testing.T) {
	client := &WhooingClient{}
	assets := []byte(`{
		"code": 200,
		"rest_of_api": 99,
		"results": {"assets": {"total": 3000, "accounts": {"x1": 3000}}}
	}`)
	liabilities := []byte(`{
		"code": 200,
		"rest_of_api": 98,
		"results": {"liabilities": {"total": 800, "accounts": {"x2": 800}}}
	}`)

	data, err := client.mergeReportResponses([][]byte{assets, liabilities})
	if err != nil {
		t.Fatalf("mergeReportResponses() error = %v", err)
	}

	var response struct {
		RestOfAPI int `json:"rest_of_api"`
		Results   struct {
			Assets      ReportGroup `json:"assets"`
			Liabilities ReportGroup `json:"liabilities"`
			Capital     ReportGroup `json:"capital"`
		} `json:"results"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if response.RestOfAPI != 98 {
		t.Errorf("rest_of_api = %d, want 98", response.RestOfAPI)
	}
	if response.Results.Assets.Total != 3000 || response.Results.Liabilities.Total != 800 {
		t.Errorf("merged totals = assets %.0f, liabilities %.0f", response.Results.Assets.Total, response.Results.Liabilities.Total)
	}
	if response.Results.Capital.Total != 2200 {
		t.Errorf("capital total = %.0f, want 2200", response.Results.Capital.Total)
	}

	bs, err := client.parseReportAsBS(data)
	if err != nil {
		t.Fatalf("parseReportAsBS() error = %v", err)
	}
	if bs.Assets.Total != 3000 || len(bs.Assets.Accounts) != 1 {
		t.Errorf("BS assets = %+v", bs.Assets)
	}
}
