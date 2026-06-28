// api/endpoint_frequent.go
// Frequent Items / Monthly Items 엔드포인트

package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ─── 데이터 모델 ──────────────────────────────────────────────

// FrequentItemInput은 자주입력 항목 생성/수정 파라미터
type FrequentItemInput struct {
	Item       string
	Money      int64
	LAccount   string
	LAccountID string
	RAccount   string
	RAccountID string
}

// MonthlyItemInput은 월별입력 항목 생성/수정 파라미터
type MonthlyItemInput struct {
	FrequentItemInput
	PayDate     int    // 결제일 1~31
	SkipHoliday string // before|after|none
}

// ─── Frequent Items ───────────────────────────────────────────

// GetFrequentItems는 전체 슬롯의 자주입력 목록 조회
// GET /api/frequent_items.json?section_id=...
func (c *WhooingClient) GetFrequentItems(sectionID string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	return c.doRequest(http.MethodGet, "/frequent_items.json", params)
}

// GetFrequentItemsSlot은 특정 슬롯의 자주입력 목록 조회
// GET /api/frequent_items/:slot.json?section_id=...
func (c *WhooingClient) GetFrequentItemsSlot(sectionID, slot string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	return c.doRequest(http.MethodGet, fmt.Sprintf("/frequent_items/%s.json", slot), params)
}

// CreateFrequentItem은 자주입력 항목 생성
// POST /api/frequent_items/:slot.json
func (c *WhooingClient) CreateFrequentItem(sectionID, slot string, p FrequentItemInput) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("item", p.Item)
	if p.Money > 0 {
		params.Set("money", fmt.Sprintf("%d", p.Money))
	}
	params.Set("l_account", p.LAccount)
	params.Set("l_account_id", p.LAccountID)
	params.Set("r_account", p.RAccount)
	params.Set("r_account_id", p.RAccountID)
	return c.doRequest(http.MethodPost, fmt.Sprintf("/frequent_items/%s.json", slot), params)
}

// UpdateFrequentItem은 자주입력 항목 수정
// PUT /api/frequent_items/:slot/:item_id.json
func (c *WhooingClient) UpdateFrequentItem(sectionID, slot, itemID string, p FrequentItemInput) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("item", p.Item)
	if p.Money > 0 {
		params.Set("money", fmt.Sprintf("%d", p.Money))
	}
	params.Set("l_account", p.LAccount)
	params.Set("l_account_id", p.LAccountID)
	params.Set("r_account", p.RAccount)
	params.Set("r_account_id", p.RAccountID)
	return c.doRequest(http.MethodPut, fmt.Sprintf("/frequent_items/%s/%s.json", slot, itemID), params)
}

// DeleteFrequentItem은 자주입력 항목 삭제
// DELETE /api/frequent_items/:slot/:item_id/:section_id.json
func (c *WhooingClient) DeleteFrequentItem(sectionID, slot, itemID string) ([]byte, error) {
	return c.doRequest(http.MethodDelete,
		fmt.Sprintf("/frequent_items/%s/%s/%s.json", slot, itemID, sectionID), nil)
}

// SortFrequentItems는 자주입력 항목 순서 변경
// PUT /api/frequent_items/:slot/sort.json
func (c *WhooingClient) SortFrequentItems(sectionID, slot string, ids []string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("item_ids", strings.Join(ids, ","))
	return c.doRequest(http.MethodPut, fmt.Sprintf("/frequent_items/%s/sort.json", slot), params)
}

// ─── Monthly Items ────────────────────────────────────────────

// GetMonthlyItems는 전체 슬롯의 월별입력 목록 조회
// GET /api/monthly_items.json?section_id=...
func (c *WhooingClient) GetMonthlyItems(sectionID string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	return c.doRequest(http.MethodGet, "/monthly_items.json", params)
}

// GetMonthlyItemsSlot은 특정 슬롯의 월별입력 목록 조회
// GET /api/monthly_items/:slot.json?section_id=...
func (c *WhooingClient) GetMonthlyItemsSlot(sectionID, slot string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	return c.doRequest(http.MethodGet, fmt.Sprintf("/monthly_items/%s.json", slot), params)
}

// CreateMonthlyItem은 월별입력 항목 생성
// POST /api/monthly_items/:slot.json
func (c *WhooingClient) CreateMonthlyItem(sectionID, slot string, p MonthlyItemInput) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("item", p.Item)
	if p.Money > 0 {
		params.Set("money", fmt.Sprintf("%d", p.Money))
	}
	params.Set("l_account", p.LAccount)
	params.Set("l_account_id", p.LAccountID)
	params.Set("r_account", p.RAccount)
	params.Set("r_account_id", p.RAccountID)
	if p.PayDate > 0 {
		params.Set("pay_date", fmt.Sprintf("%d", p.PayDate))
	}
	if p.SkipHoliday != "" {
		params.Set("skip_holiday", p.SkipHoliday)
	}
	return c.doRequest(http.MethodPost, fmt.Sprintf("/monthly_items/%s.json", slot), params)
}

// UpdateMonthlyItem은 월별입력 항목 수정
// PUT /api/monthly_items/:slot/:item_id.json
func (c *WhooingClient) UpdateMonthlyItem(sectionID, slot, itemID string, p MonthlyItemInput) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("item", p.Item)
	if p.Money > 0 {
		params.Set("money", fmt.Sprintf("%d", p.Money))
	}
	params.Set("l_account", p.LAccount)
	params.Set("l_account_id", p.LAccountID)
	params.Set("r_account", p.RAccount)
	params.Set("r_account_id", p.RAccountID)
	if p.PayDate > 0 {
		params.Set("pay_date", fmt.Sprintf("%d", p.PayDate))
	}
	if p.SkipHoliday != "" {
		params.Set("skip_holiday", p.SkipHoliday)
	}
	return c.doRequest(http.MethodPut, fmt.Sprintf("/monthly_items/%s/%s.json", slot, itemID), params)
}

// DeleteMonthlyItem은 월별입력 항목 삭제
// DELETE /api/monthly_items/:slot/:item_id/:section_id.json
func (c *WhooingClient) DeleteMonthlyItem(sectionID, slot, itemID string) ([]byte, error) {
	return c.doRequest(http.MethodDelete,
		fmt.Sprintf("/monthly_items/%s/%s/%s.json", slot, itemID, sectionID), nil)
}

// SortMonthlyItems는 월별입력 항목 순서 변경
// PUT /api/monthly_items/:slot/sort.json
func (c *WhooingClient) SortMonthlyItems(sectionID, slot string, ids []string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("item_ids", strings.Join(ids, ","))
	return c.doRequest(http.MethodPut, fmt.Sprintf("/monthly_items/%s/sort.json", slot), params)
}
