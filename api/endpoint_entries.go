// api/endpoint_entries.go
// Entries 고급 기능 엔드포인트 (일괄, 검색, 흐름, 외부)

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ─── 데이터 모델 ──────────────────────────────────────────────

// EntryInput은 일괄입력용 거래 단건 구조체
type EntryInput struct {
	EntryDate     int    `json:"entry_date,omitempty"`
	LAccount      string `json:"l_account"`
	LAccountID    string `json:"l_account_id"`
	RAccount      string `json:"r_account"`
	RAccountID    string `json:"r_account_id"`
	Item          string `json:"item,omitempty"`
	Money         int64  `json:"money"`
	Memo          string `json:"memo,omitempty"`
	AttachmentIDs string `json:"attachment_ids,omitempty"`
}

// EntryChanges는 복수 거래 일괄 수정 시 변경할 필드 (비어있는 필드는 미전송)
type EntryChanges struct {
	EntryDate  int
	LAccount   string
	LAccountID string
	RAccount   string
	RAccountID string
	Item       string
	Money      int64
	Memo       string
}

// EntrySearch는 고급 검색 파라미터
type EntrySearch struct {
	SectionID string
	StartDate int
	EndDate   int
	Max       string // entry_date 커서 ("20260101.0034")
	Limit     int    // 최대 100

	Account    string // 좌우 공통 계정
	AccountID  string // 좌우 공통 항목 (Account 필수)
	LAccount   string
	LAccountID string
	RAccount   string
	RAccountID string

	Item      string // 와일드카드(*) + (detail) 지원
	Memo      string // 공백=AND, ! prefix=제외
	MoneyFrom int64
	MoneyTo   int64

	SortColumn string // entry_date|item|money|total|l_account_id|r_account_id
	SortOrder  string // desc|asc
}

// FlowQuery는 flow/changes 엔드포인트 공통 파라미터 (entries GET과 동일)
type FlowQuery struct {
	SectionID  string
	StartDate  int
	EndDate    int
	Account    string
	AccountID  string
	LAccount   string
	LAccountID string
	RAccount   string
	RAccountID string
	Item       string
	Memo       string
	RowsType   string // day|month|year (changes 엔드포인트용)
}

// ─── 헬퍼 ─────────────────────────────────────────────────────

func entrySearchToValues(q EntrySearch) url.Values {
	p := url.Values{}
	p.Set("section_id", q.SectionID)
	if q.StartDate > 0 {
		p.Set("start_date", fmt.Sprintf("%d", q.StartDate))
	}
	if q.EndDate > 0 {
		p.Set("end_date", fmt.Sprintf("%d", q.EndDate))
	}
	if q.Max != "" {
		p.Set("max", q.Max)
	}
	if q.Limit > 0 {
		p.Set("limit", fmt.Sprintf("%d", q.Limit))
	}
	if q.Account != "" {
		p.Set("account", q.Account)
	}
	if q.AccountID != "" {
		p.Set("account_id", q.AccountID)
	}
	if q.LAccount != "" {
		p.Set("l_account", q.LAccount)
	}
	if q.LAccountID != "" {
		p.Set("l_account_id", q.LAccountID)
	}
	if q.RAccount != "" {
		p.Set("r_account", q.RAccount)
	}
	if q.RAccountID != "" {
		p.Set("r_account_id", q.RAccountID)
	}
	if q.Item != "" {
		p.Set("item", q.Item)
	}
	if q.Memo != "" {
		p.Set("memo", q.Memo)
	}
	if q.MoneyFrom > 0 {
		p.Set("money_from", fmt.Sprintf("%d", q.MoneyFrom))
	}
	if q.MoneyTo > 0 {
		p.Set("money_to", fmt.Sprintf("%d", q.MoneyTo))
	}
	if q.SortColumn != "" {
		p.Set("sort_column", q.SortColumn)
	}
	if q.SortOrder != "" {
		p.Set("sort_order", q.SortOrder)
	}
	return p
}

func flowQueryToValues(q FlowQuery) url.Values {
	p := url.Values{}
	p.Set("section_id", q.SectionID)
	if q.StartDate > 0 {
		p.Set("start_date", fmt.Sprintf("%d", q.StartDate))
	}
	if q.EndDate > 0 {
		p.Set("end_date", fmt.Sprintf("%d", q.EndDate))
	}
	if q.Account != "" {
		p.Set("account", q.Account)
	}
	if q.AccountID != "" {
		p.Set("account_id", q.AccountID)
	}
	if q.LAccount != "" {
		p.Set("l_account", q.LAccount)
	}
	if q.LAccountID != "" {
		p.Set("l_account_id", q.LAccountID)
	}
	if q.RAccount != "" {
		p.Set("r_account", q.RAccount)
	}
	if q.RAccountID != "" {
		p.Set("r_account_id", q.RAccountID)
	}
	if q.Item != "" {
		p.Set("item", q.Item)
	}
	if q.Memo != "" {
		p.Set("memo", q.Memo)
	}
	if q.RowsType != "" {
		p.Set("rows_type", q.RowsType)
	}
	return p
}

// ─── 일괄 CRUD ────────────────────────────────────────────────

// CreateEntriesBatch는 최대 300건 일괄 입력
// POST /api/entries.json (entries 파라미터에 JSON 배열 문자열 전송)
// 300건 초과 시 자동 분할하여 순차 전송
func (c *WhooingClient) CreateEntriesBatch(sectionID string, rows []EntryInput) ([]byte, error) {
	const batchSize = 300

	var lastData []byte
	for i := 0; i < len(rows); i += batchSize {
		end := i + batchSize
		if end > len(rows) {
			end = len(rows)
		}
		chunk := rows[i:end]

		entriesJSON, err := json.Marshal(chunk)
		if err != nil {
			return nil, fmt.Errorf("일괄입력 JSON 직렬화 실패: %w", err)
		}

		params := url.Values{}
		params.Set("section_id", sectionID)
		params.Set("data_type", "json")
		params.Set("entries", string(entriesJSON))

		data, err := c.doRequest(http.MethodPost, "/entries.json", params)
		if err != nil {
			return nil, fmt.Errorf("일괄입력 실패 (chunk %d~%d): %w", i, end-1, err)
		}
		lastData = data
	}
	return lastData, nil
}

// UpdateEntriesBatch는 최대 100건 복수 수정
// PUT /api/entries/:entry_ids/:section_id.json
func (c *WhooingClient) UpdateEntriesBatch(sectionID string, ids []int64, changes EntryChanges) ([]byte, error) {
	idStrs := make([]string, len(ids))
	for i, id := range ids {
		idStrs[i] = fmt.Sprintf("%d", id)
	}
	endpoint := fmt.Sprintf("/entries/%s/%s.json", strings.Join(idStrs, ","), sectionID)

	params := url.Values{}
	if changes.EntryDate > 0 {
		params.Set("entry_date", fmt.Sprintf("%d", changes.EntryDate))
	}
	if changes.LAccount != "" {
		params.Set("l_account", changes.LAccount)
		params.Set("l_account_id", changes.LAccountID)
	}
	if changes.RAccount != "" {
		params.Set("r_account", changes.RAccount)
		params.Set("r_account_id", changes.RAccountID)
	}
	if changes.Item != "" {
		params.Set("item", changes.Item)
	}
	if changes.Money > 0 {
		params.Set("money", fmt.Sprintf("%d", changes.Money))
	}
	if changes.Memo != "" {
		params.Set("memo", changes.Memo)
	}

	return c.doRequest(http.MethodPut, endpoint, params)
}

// DeleteEntries는 최대 100건 복수 삭제
// DELETE /api/entries/:entry_ids/:section_id.json
func (c *WhooingClient) DeleteEntries(sectionID string, ids []int64) ([]byte, error) {
	idStrs := make([]string, len(ids))
	for i, id := range ids {
		idStrs[i] = fmt.Sprintf("%d", id)
	}
	endpoint := fmt.Sprintf("/entries/%s/%s.json", strings.Join(idStrs, ","), sectionID)
	return c.doRequest(http.MethodDelete, endpoint, nil)
}

// ─── 고급 검색 ────────────────────────────────────────────────

// SearchEntries는 고급 필터 조건으로 거래 검색
// GET /api/entries.json
func (c *WhooingClient) SearchEntries(q EntrySearch) ([]byte, error) {
	return c.doRequest(http.MethodGet, "/entries.json", entrySearchToValues(q))
}

// ─── 흐름/변동 분석 ───────────────────────────────────────────

// FlowOfAccount는 특정 계정과 모든 계정/항목의 상대적 증감 조회
// GET /api/entries/flow_of_account.json
func (c *WhooingClient) FlowOfAccount(q FlowQuery) ([]byte, error) {
	return c.doRequest(http.MethodGet, "/entries/flow_of_account.json", flowQueryToValues(q))
}

// FlowOfAccountID는 특정 항목과 모든 계정/항목의 상대적 증감 조회
// GET /api/entries/flow_of_account_id.json
func (c *WhooingClient) FlowOfAccountID(q FlowQuery) ([]byte, error) {
	return c.doRequest(http.MethodGet, "/entries/flow_of_account_id.json", flowQueryToValues(q))
}

// ChangesOfAccountID는 특정 항목의 일일 변동 내역 조회
// GET /api/entries/changes_of_account_id.json
func (c *WhooingClient) ChangesOfAccountID(q FlowQuery) ([]byte, error) {
	return c.doRequest(http.MethodGet, "/entries/changes_of_account_id.json", flowQueryToValues(q))
}

// ChangesOfClient는 특정 거래처의 일일 변동 내역 조회
// GET /api/entries/changes_of_client.json
func (c *WhooingClient) ChangesOfClient(q FlowQuery) ([]byte, error) {
	return c.doRequest(http.MethodGet, "/entries/changes_of_client.json", flowQueryToValues(q))
}

// ChangesOfItem은 특정 아이템의 일일 발생 내역 조회
// GET /api/entries/changes_of_item.json
func (c *WhooingClient) ChangesOfItem(q FlowQuery) ([]byte, error) {
	return c.doRequest(http.MethodGet, "/entries/changes_of_item.json", flowQueryToValues(q))
}

// ─── 집계 ─────────────────────────────────────────────────────

// AccountIDsOfAccount는 계정의 항목별 금액 집계 조회
// GET /api/entries/account_ids_of_account.json
func (c *WhooingClient) AccountIDsOfAccount(q FlowQuery) ([]byte, error) {
	return c.doRequest(http.MethodGet, "/entries/account_ids_of_account.json", flowQueryToValues(q))
}

// ClientsOfAccountID는 항목의 거래처별 금액 집계 조회
// GET /api/entries/clients_of_account_id.json
func (c *WhooingClient) ClientsOfAccountID(q FlowQuery) ([]byte, error) {
	return c.doRequest(http.MethodGet, "/entries/clients_of_account_id.json", flowQueryToValues(q))
}

// ItemsOfAccountID는 항목의 아이템별 금액 집계 조회
// GET /api/entries/items_of_account_id.json
func (c *WhooingClient) ItemsOfAccountID(q FlowQuery) ([]byte, error) {
	return c.doRequest(http.MethodGet, "/entries/items_of_account_id.json", flowQueryToValues(q))
}

// ─── 외부 데이터 파싱 ─────────────────────────────────────────

// ParseOutside는 SMS 등 외부 데이터를 파싱하여 거래 자동 입력
// POST /api/entries/outside.json
func (c *WhooingClient) ParseOutside(sectionID, rows string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("rows", rows)
	return c.doRequest(http.MethodPost, "/entries/outside.json", params)
}

// ReportOutside는 인식되지 않은 외부 데이터 소스를 보고
// POST /api/entries/outside_report.json
func (c *WhooingClient) ReportOutside(source string) ([]byte, error) {
	params := url.Values{}
	params.Set("source", source)
	return c.doRequest(http.MethodPost, "/entries/outside_report.json", params)
}
