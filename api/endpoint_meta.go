// api/endpoint_meta.go
// Sections / Accounts CRUD 엔드포인트

package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ─── 데이터 모델 ──────────────────────────────────────────────

// SectionCreateParams는 섹션 생성 파라미터
type SectionCreateParams struct {
	Title         string
	Currency      string
	Memo          string
	SkinID        int
	DecimalPlaces int
	DateFormat    string
	StartYear     int
	TemplateID    int
}

// SectionUpdateParams는 섹션 수정 파라미터
// UISettings: ui[key]=value 스타일로 직렬화됨
type SectionUpdateParams struct {
	Title         string
	Currency      string
	Memo          string
	SkinID        int
	DecimalPlaces int
	DateFormat    string
	UISettings    map[string]string // ui[budgetLong]=y 등
}

// AccountCreateParams는 항목 생성 파라미터
type AccountCreateParams struct {
	SectionID    string
	Title        string
	Type         string // account | group
	OpenDate     int    // YYYYMMDD
	CloseDate    int    // 29991231 = 종료없음
	Memo         string
	Category     string // normal|client|creditcard|checkcard|steady|floating
	OptUseDate   string // 신용카드: pp1~p31
	OptPayDate   int    // 신용카드: 1~31
	OptPayAccID  string // 신용카드: 대금결제 자산항목 id
}

// AccountUpdateParams는 항목 수정 파라미터 (전체 필드 전송 필수)
type AccountUpdateParams struct {
	SectionID   string
	Title       string
	OpenDate    int
	CloseDate   int
	Memo        string
	Category    string
	OptUseDate  string
	OptPayDate  int
	OptPayAccID string
}

// AccountExistsResult는 항목 거래 존재 여부 확인 결과
type AccountExistsResult struct {
	Count     int    `json:"count"`
	MinDate   int    `json:"minDate"`
	MaxDate   int    `json:"maxDate"`
	Balance   int64  `json:"balance"`
	LastOne   string `json:"last_one"`
	CloseDate int    `json:"close_date"`
}

// ─── Sections CRUD ────────────────────────────────────────────

// CreateSection은 섹션 신규 생성
// POST /api/sections.json
func (c *WhooingClient) CreateSection(p SectionCreateParams) ([]byte, error) {
	params := url.Values{}
	params.Set("title", p.Title)
	params.Set("currency", p.Currency)
	if p.Memo != "" {
		params.Set("memo", p.Memo)
	}
	if p.SkinID > 0 {
		params.Set("skin_id", fmt.Sprintf("%d", p.SkinID))
	}
	if p.DecimalPlaces > 0 {
		params.Set("decimal_places", fmt.Sprintf("%d", p.DecimalPlaces))
	}
	if p.DateFormat != "" {
		params.Set("date_format", p.DateFormat)
	}
	if p.StartYear > 0 {
		params.Set("start_year", fmt.Sprintf("%d", p.StartYear))
	}
	if p.TemplateID > 0 {
		params.Set("template_id", fmt.Sprintf("%d", p.TemplateID))
	}
	return c.doRequest(http.MethodPost, "/sections.json", params)
}

// UpdateSection은 섹션 정보 수정
// PUT /api/sections/:section_id.json
func (c *WhooingClient) UpdateSection(sectionID string, p SectionUpdateParams) ([]byte, error) {
	params := url.Values{}
	params.Set("title", p.Title)
	params.Set("currency", p.Currency)
	params.Set("memo", p.Memo)
	if p.SkinID >= 0 {
		params.Set("skin_id", fmt.Sprintf("%d", p.SkinID))
	}
	if p.DecimalPlaces >= 0 {
		params.Set("decimal_places", fmt.Sprintf("%d", p.DecimalPlaces))
	}
	if p.DateFormat != "" {
		params.Set("date_format", p.DateFormat)
	}
	// UI 설정: ui[key]=value 직렬화
	for k, v := range p.UISettings {
		params.Set(fmt.Sprintf("ui[%s]", k), v)
	}
	// 섹션 변경 시 accounts 캐시 무효화
	c.InvalidateCache(fmt.Sprintf("accounts:%s", sectionID), CacheKeySections)
	return c.doRequest(http.MethodPut, fmt.Sprintf("/sections/%s.json", sectionID), params)
}

// DeleteSections는 섹션 삭제 (복수 가능, 콤마 구분)
// DELETE /api/sections/:section_id.json (section_id에 "s1,s2" 형태 가능)
func (c *WhooingClient) DeleteSections(ids []string) ([]byte, error) {
	sectionID := strings.Join(ids, ",")
	c.InvalidateCache(CacheKeySections)
	return c.doRequest(http.MethodDelete, fmt.Sprintf("/sections/%s.json", sectionID), nil)
}

// SortSections는 섹션 순서 변경
// PUT /api/sections/sort.json
func (c *WhooingClient) SortSections(ids []string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_ids", strings.Join(ids, ","))
	c.InvalidateCache(CacheKeySections)
	return c.doRequest(http.MethodPut, "/sections/sort.json", params)
}

// ─── Accounts CRUD ────────────────────────────────────────────

// CreateAccount는 항목 신규 생성
// POST /api/accounts/:account.json
func (c *WhooingClient) CreateAccount(account string, p AccountCreateParams) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", p.SectionID)
	params.Set("title", p.Title)
	params.Set("type", p.Type)
	if p.OpenDate > 0 {
		params.Set("open_date", fmt.Sprintf("%d", p.OpenDate))
	}
	if p.CloseDate > 0 {
		params.Set("close_date", fmt.Sprintf("%d", p.CloseDate))
	}
	if p.Memo != "" {
		params.Set("memo", p.Memo)
	}
	if p.Category != "" {
		params.Set("category", p.Category)
	}
	if p.OptUseDate != "" {
		params.Set("opt_use_date", p.OptUseDate)
	}
	if p.OptPayDate > 0 {
		params.Set("opt_pay_date", fmt.Sprintf("%d", p.OptPayDate))
	}
	if p.OptPayAccID != "" {
		params.Set("opt_pay_account_id", p.OptPayAccID)
	}
	// 항목 변경 시 캐시 무효화
	c.InvalidateCache(fmt.Sprintf("accounts:%s", p.SectionID))
	return c.doRequest(http.MethodPost, fmt.Sprintf("/accounts/%s.json", account), params)
}

// UpdateAccount는 항목 정보 수정 (전체 필드 전송 필수)
// PUT /api/accounts/:account/:account_id.json
func (c *WhooingClient) UpdateAccount(account, accountID string, p AccountUpdateParams) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", p.SectionID)
	params.Set("title", p.Title)
	if p.OpenDate > 0 {
		params.Set("open_date", fmt.Sprintf("%d", p.OpenDate))
	}
	if p.CloseDate > 0 {
		params.Set("close_date", fmt.Sprintf("%d", p.CloseDate))
	}
	params.Set("memo", p.Memo)
	if p.Category != "" {
		params.Set("category", p.Category)
	}
	if p.OptUseDate != "" {
		params.Set("opt_use_date", p.OptUseDate)
	}
	if p.OptPayDate > 0 {
		params.Set("opt_pay_date", fmt.Sprintf("%d", p.OptPayDate))
	}
	if p.OptPayAccID != "" {
		params.Set("opt_pay_account_id", p.OptPayAccID)
	}
	c.InvalidateCache(fmt.Sprintf("accounts:%s", p.SectionID))
	return c.doRequest(http.MethodPut, fmt.Sprintf("/accounts/%s/%s.json", account, accountID), params)
}

// DeleteAccount는 항목 삭제
// DELETE /api/accounts/:account/:account_id/:section_id.json
func (c *WhooingClient) DeleteAccount(account, accountID, sectionID string) ([]byte, error) {
	c.InvalidateCache(fmt.Sprintf("accounts:%s", sectionID))
	return c.doRequest(http.MethodDelete, fmt.Sprintf("/accounts/%s/%s/%s.json", account, accountID, sectionID), nil)
}

// AccountExists는 항목에 거래 존재 여부 확인 (삭제 전 필수 호출)
// GET /api/accounts/:account/:account_id/exists.json?section_id=...
func (c *WhooingClient) AccountExists(account, accountID, sectionID string) (*AccountExistsResult, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	data, err := c.doRequest(http.MethodGet, fmt.Sprintf("/accounts/%s/%s/exists.json", account, accountID), params)
	if err != nil {
		return nil, err
	}
	var result AccountExistsResult
	if err := parseResponseWithClient(c, data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SortAccounts는 항목 순서 변경
// PUT /api/accounts/:account/sort.json
func (c *WhooingClient) SortAccounts(account, sectionID string, ids []string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("account_ids", strings.Join(ids, ","))
	c.InvalidateCache(fmt.Sprintf("accounts:%s", sectionID))
	return c.doRequest(http.MethodPut, fmt.Sprintf("/accounts/%s/sort.json", account), params)
}
