// api/whooing.go
// Whooing API 클라이언트

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"whoo-cli/config"
)

const (
	apiBaseURL = "https://whooing.com/api"
)

// WhooingClient는 후잉 API 클라이언트
type WhooingClient struct {
	config         *config.Config
	httpClient     *http.Client
	cache          TTLCache
	lastRestOfAPI  int64 // atomic 접근용
	bucketMu       sync.Mutex
	bucketTokens   int       // 현재 토큰 수 (최대 bucketMax)
	bucketRefillAt time.Time // 다음 리필 시각
}

// 토큰버킷 설정: 분당 20회, 버스트 5회
const (
	bucketMax      = 5
	bucketPerMin   = 20
	bucketInterval = time.Minute / bucketPerMin // 3초마다 1토큰 충전
)

// APIResponse는 후잉 API 공통 응답 구조
type APIResponse struct {
	Code            int             `json:"code"`
	Message         string          `json:"message"`
	Error           json.RawMessage `json:"error"`
	ErrorParameters json.RawMessage `json:"error_parameters"`
	RestOfAPI       int             `json:"rest_of_api"`
	Results         json.RawMessage `json:"results"`
}

// LastRestOfAPI는 마지막으로 수신된 rest_of_api 값 반환
func (c *WhooingClient) LastRestOfAPI() int {
	return int(atomic.LoadInt64(&c.lastRestOfAPI))
}

// InvalidateCache는 지정 키들의 캐시를 즉시 무효화
func (c *WhooingClient) InvalidateCache(keys ...string) {
	c.cache.Invalidate(keys...)
}

// SectionUI는 섹션 UI 설정
type SectionUI struct {
	BudgetLong string `json:"budgetLong"` // "y" 이면 장기목표 연동 모드
}

// Section은 후잉 섹션(가계부) 정보
type Section struct {
	SectionID string    `json:"section_id"`
	Title     string    `json:"title"`
	Currency  string    `json:"currency"`
	Isolation string    `json:"isolation"`
	UI        SectionUI `json:"ui"`
}

// Account는 계정 정보 (json_array 응답용)
type Account struct {
	AccountID string `json:"account_id"`
	Name      string `json:"name"`
	Type      string `json:"type"` // assets, liabilities, capital, expenses, income
}

// AccountDetail은 개별 계정 상세 정보 (json 객체 응답용)
// accounts.json 응답: {"assets": {"id1": {"title": "...", ...}}, ...}
// 거래 입력과 BS 계정명 표시에 필요한 필드만 정의
// open_date 등은 number 타입일 수 있어 interface{}로 처리
type AccountDetail struct {
	Title         string      `json:"title"`
	Type          string      `json:"type"`
	Category      string      `json:"category"`
	OpenDate      interface{} `json:"open_date"`
	CloseDate     interface{} `json:"close_date"`
	Memo          string      `json:"memo"`
	OpeningAmount float64     `json:"opening_amount"`
}

// AccountsMap은 계정 타입별로 그룹핑된 계정 목록
// accounts.json 응답 형태: {"assets": {"id": {...}}, "liabilities": {"id": {...}}, ...}
type AccountsMap struct {
	Assets      map[string]AccountDetail `json:"assets"`
	Liabilities map[string]AccountDetail `json:"liabilities"`
	Capital     map[string]AccountDetail `json:"capital"`
	Expenses    map[string]AccountDetail `json:"expenses"`
	Income      map[string]AccountDetail `json:"income"`
}

// GetAccountsByType은 특정 계정 타입의 계정 목록을 ID→Detail 맵으로 반환
func (am *AccountsMap) GetAccountsByType(accountType string) map[string]AccountDetail {
	switch accountType {
	case "assets":
		return am.Assets
	case "liabilities":
		return am.Liabilities
	case "capital":
		return am.Capital
	case "expenses":
		return am.Expenses
	case "income":
		return am.Income
	default:
		return nil
	}
}

// GetTitle은 계정 ID로 계정명을 찾아 반환
func (am *AccountsMap) GetTitle(accountType, accountID string) string {
	accounts := am.GetAccountsByType(accountType)
	if accounts == nil {
		return accountID
	}
	if detail, ok := accounts[accountID]; ok {
		return detail.Title
	}
	return accountID
}

// Entry는 거래 내역
// 실제 API 응답: entry_id는 int, entry_date는 숫자 또는 "YYYYMMDD.NNNN" 문자열
type Entry struct {
	EntryID    int            `json:"entry_id"`
	EntryDate  FlexibleString `json:"entry_date"` // "20260310.0007" 형식 (뒤 소수점은 정렬용)
	LAccount   string         `json:"l_account"`  // 왼쪽
	LAccountID string         `json:"l_account_id"`
	RAccount   string         `json:"r_account"` // 오른쪽
	RAccountID string         `json:"r_account_id"`
	Money      float64        `json:"money"`
	Item       string         `json:"item"`
	Memo       string         `json:"memo"`
}

// FlexibleString은 JSON string 또는 number를 모두 문자열로 수용한다
// Whooing API는 entry_date 등을 number(20110812.0001) 또는 string으로 반환할 수 있다
type FlexibleString string

// UnmarshalJSON은 string/number/null을 모두 수용
func (s *FlexibleString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*s = ""
		return nil
	}
	if data[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		*s = FlexibleString(str)
		return nil
	}
	// number (정수/소수)
	var num json.Number
	if err := json.Unmarshal(data, &num); err != nil {
		return fmt.Errorf("FlexibleString: %w", err)
	}
	*s = FlexibleString(num.String())
	return nil
}

// String은 FlexibleString을 일반 문자열로 반환
func (s FlexibleString) String() string {
	return string(s)
}

// DateOnly는 entry_date에서 날짜 부분(YYYYMMDD)만 반환
func (e *Entry) DateOnly() string {
	ed := string(e.EntryDate)
	if idx := strings.Index(ed, "."); idx > 0 {
		return ed[:idx]
	}
	return ed
}

// entriesResponse는 거래 내역 API 응답의 results 구조
// 실제: {"rows": [...], "reports": [...]}
type entriesResponse struct {
	Rows []Entry `json:"rows"`
}

// BSResponse는 자산부채(Balance Sheet) API 응답의 results 구조
// 실제: {"assets": {"total": N, "accounts": [...]}, "liabilities": {"total": N, "accounts": [...]}}
type BSResponse struct {
	Assets      BSGroup `json:"assets"`
	Liabilities BSGroup `json:"liabilities"`
}

// BSGroup은 BS 응답의 계정 타입별 그룹
type BSGroup struct {
	Total    float64     `json:"total"`
	Accounts []BSAccount `json:"accounts"`
}

// BSAccount는 BS 응답의 개별 계정 항목
type BSAccount struct {
	AccountID string  `json:"account_id"`
	Money     float64 `json:"money"`
}

// NewWhooingClient는 새 API 클라이언트 생성
func NewWhooingClient(cfg *config.Config) *WhooingClient {
	return &WhooingClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		bucketTokens:   bucketMax,
		bucketRefillAt: time.Now().Add(bucketInterval),
	}
}

// buildAPIKey는 X-API-KEY 헤더 값 생성
// app_id/secret 미설정 시 빈 헤더 대신 에러를 반환해 원인 불명 400을 막는다
func (c *WhooingClient) buildAPIKey() (string, error) {
	appID, err := config.GetAppID()
	if err != nil {
		return "", err
	}
	if appID == "" {
		return "", fmt.Errorf("app_id가 비어 있습니다")
	}
	signiture := c.config.ComputeSigniture()
	if signiture == "" {
		// secret 미설정 또는 token_secret 없음
		if _, err := config.GetAppSecret(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("signiture 계산 실패: token_secret을 확인하세요")
	}
	timestamp := time.Now().UnixMilli()
	return fmt.Sprintf("app_id=%s,token=%s,signiture=%s,timestamp=%d",
		appID, c.config.Token, signiture, timestamp), nil
}

// acquireToken은 토큰버킷에서 토큰 1개를 소비. 필요 시 대기
func (c *WhooingClient) acquireToken() {
	c.bucketMu.Lock()
	defer c.bucketMu.Unlock()

	now := time.Now()
	// 경과 시간만큼 토큰 충전
	if now.After(c.bucketRefillAt) {
		elapsed := now.Sub(c.bucketRefillAt)
		added := int(elapsed/bucketInterval) + 1
		c.bucketTokens += added
		if c.bucketTokens > bucketMax {
			c.bucketTokens = bucketMax
		}
		c.bucketRefillAt = now.Add(bucketInterval)
	}

	if c.bucketTokens <= 0 {
		// 토큰이 없으면 다음 리필까지 대기
		wait := time.Until(c.bucketRefillAt)
		if wait > 0 {
			time.Sleep(wait)
		}
		c.bucketTokens = 1
		c.bucketRefillAt = time.Now().Add(bucketInterval)
	}
	c.bucketTokens--
}

// doRequest는 공통 HTTP 요청 처리 (토큰버킷 + 재시도 포함)
func (c *WhooingClient) doRequest(method, endpoint string, params url.Values) ([]byte, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		c.acquireToken()

		body, statusCode, err := c.doRequestOnce(method, endpoint, params)
		if err == nil {
			return body, nil
		}

		// APIError 분기 처리
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			switch {
			case apiErr.IsTokenExpired(): // 405: 재시도 금지, 즉시 반환
				return nil, apiErr
			case apiErr.IsDailyLimit(): // 402: 재시도 금지, 즉시 반환
				return nil, apiErr
			case apiErr.IsRateLimit(): // 429: 3초 대기 후 재시도
				time.Sleep(3 * time.Second)
				lastErr = err
				continue
			}
		}

		// HTTP 500: 3초 대기 후 1회 재시도
		if statusCode == http.StatusInternalServerError && attempt == 0 {
			time.Sleep(3 * time.Second)
			lastErr = err
			continue
		}

		return nil, err
	}

	return nil, fmt.Errorf("재시도 초과: %w", lastErr)
}

// doRequestOnce는 단일 HTTP 요청 실행. HTTP 상태코드도 함께 반환
func (c *WhooingClient) doRequestOnce(method, endpoint string, params url.Values) ([]byte, int, error) {
	reqURL := fmt.Sprintf("%s%s", apiBaseURL, endpoint)
	if (method == http.MethodGet || method == http.MethodDelete) && params != nil {
		reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
	}

	var bodyReader io.Reader
	if (method == http.MethodPost || method == http.MethodPut) && params != nil {
		bodyReader = strings.NewReader(params.Encode())
	}

	req, err := http.NewRequest(method, reqURL, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("요청 생성 실패: %w", err)
	}

	apiKey, err := c.buildAPIKey()
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Accept", "application/json")
	if method == http.MethodPost || method == http.MethodPut {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("API 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("응답 읽기 실패: %w", err)
	}

	// HTTP 오류 상태코드를 APIError로 변환
	switch resp.StatusCode {
	case http.StatusOK:
		// 정상
	case 429:
		return nil, 429, &APIError{Code: 429, Message: "분당 요청 한도 초과", Endpoint: endpoint}
	case http.StatusPaymentRequired: // 402
		return nil, 402, &APIError{Code: 402, Message: "일일 API 한도 초과", Endpoint: endpoint}
	case http.StatusMethodNotAllowed: // 405
		return nil, 405, &APIError{Code: 405, Message: "토큰이 만료되었습니다. 재인증이 필요합니다", Endpoint: endpoint}
	default:
		return nil, resp.StatusCode, &APIError{
			Code:     resp.StatusCode,
			Message:  fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
			Endpoint: endpoint,
		}
	}

	return body, http.StatusOK, nil
}

// parseResponse는 공통 응답 파싱 및 에러 체크
// client가 nil이 아니면 rest_of_api 값을 클라이언트에 저장
func parseResponse(data []byte, target interface{}) error {
	return parseResponseWithClient(nil, data, target)
}

// parseResponseWithClient는 클라이언트에 rest_of_api를 기록하는 응답 파싱
func parseResponseWithClient(c *WhooingClient, data []byte, target interface{}) error {
	var apiResp APIResponse
	if err := json.Unmarshal(data, &apiResp); err != nil {
		return fmt.Errorf("응답 파싱 실패: %w", err)
	}

	// rest_of_api 저장
	if c != nil {
		atomic.StoreInt64(&c.lastRestOfAPI, int64(apiResp.RestOfAPI))
	}

	if apiResp.Code != 200 && apiResp.Code != 204 {
		errMsg := apiResp.Message
		if errMsg == "" {
			errMsg = string(apiResp.ErrorParameters)
		}
		if errMsg == "" || errMsg == "null" {
			errMsg = string(apiResp.Error)
		}
		if errMsg == "" || errMsg == "null" {
			errMsg = "알 수 없는 오류"
		}
		return &APIError{
			Code:       apiResp.Code,
			Message:    errMsg,
			Parameters: apiResp.ErrorParameters,
		}
	}

	if target != nil && apiResp.Results != nil {
		if err := json.Unmarshal(apiResp.Results, target); err != nil {
			return fmt.Errorf("결과 파싱 실패: %w", err)
		}
	}

	return nil
}

// parseEntryArrayResponse는 거래 배열 응답을 파싱하여 첫 번째 거래 반환
func (c *WhooingClient) parseEntryArrayResponse(data []byte) (*Entry, error) {
	var entries []Entry
	if err := parseResponseWithClient(c, data, &entries); err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("응답에 거래 데이터가 없습니다")
	}
	return &entries[0], nil
}

// parseEntryObjectResponse는 results가 Entry 객체인 응답을 파싱 (수정 API)
func (c *WhooingClient) parseEntryObjectResponse(data []byte) (*Entry, error) {
	var entry Entry
	if err := parseResponseWithClient(c, data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// GetSections는 섹션(가계부) 목록 조회
// GET /api/sections.json_array
// isolation="y"인 섹션은 제외 (비공개 섹션)
func (c *WhooingClient) GetSections() ([]Section, error) {
	// 캐시 확인
	if cached, ok := c.cache.Get(CacheKeySections); ok {
		return cached.([]Section), nil
	}

	data, err := c.doRequest(http.MethodGet, "/sections.json_array", nil)
	if err != nil {
		return nil, err
	}

	var allSections []Section
	if err := parseResponseWithClient(c, data, &allSections); err != nil {
		return nil, err
	}

	var sections []Section
	for _, section := range allSections {
		if section.Isolation != "y" {
			sections = append(sections, section)
		}
	}

	c.cache.Set(CacheKeySections, sections, 5*time.Minute)
	return sections, nil
}

// GetAccounts는 계정 목록 조회 (배열 형태)
// GET /api/accounts.json_array?section_id={id}
func (c *WhooingClient) GetAccounts(sectionID string) ([]Account, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)

	data, err := c.doRequest(http.MethodGet, "/accounts.json_array", params)
	if err != nil {
		return nil, err
	}

	var accounts []Account
	if err := parseResponseWithClient(c, data, &accounts); err != nil {
		return nil, err
	}

	return accounts, nil
}

// GetAccountsMap은 계정 목록을 타입별로 그룹핑하여 조회
// GET /api/accounts.json?section_id={id}
// 응답: {"assets": {"id": {"title": ...}}, "liabilities": {...}, ...}
func (c *WhooingClient) GetAccountsMap(sectionID string) (*AccountsMap, error) {
	cacheKey := fmt.Sprintf("accounts:%s", sectionID)
	if cached, ok := c.cache.Get(cacheKey); ok {
		return cached.(*AccountsMap), nil
	}

	params := url.Values{}
	params.Set("section_id", sectionID)

	data, err := c.doRequest(http.MethodGet, "/accounts.json", params)
	if err != nil {
		return nil, err
	}

	var accountsMap AccountsMap
	if err := parseResponseWithClient(c, data, &accountsMap); err != nil {
		return nil, err
	}

	c.cache.Set(cacheKey, &accountsMap, 5*time.Minute)
	return &accountsMap, nil
}

// GetEntries는 거래 내역 조회
// GET /api/entries.json_array?section_id={id}&start_date={YYYYMMDD}&end_date={YYYYMMDD}
// limit: 조회 건수 (0이면 파라미터 미전송, API 기본값 20 사용)
// cursor: max 파라미터용 전체 entry_date 값 (빈 문자열이면 미사용)
func (c *WhooingClient) GetEntries(sectionID, startDate, endDate string, limit int, cursor string) ([]Entry, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("start_date", startDate)
	params.Set("end_date", endDate)
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	if cursor != "" {
		params.Set("max", cursor)
	}

	data, err := c.doRequest(http.MethodGet, "/entries.json_array", params)
	if err != nil {
		return nil, err
	}

	var resp entriesResponse
	if err := parseResponseWithClient(c, data, &resp); err != nil {
		return nil, err
	}

	return resp.Rows, nil
}

// GetBS는 자산부채(Balance Sheet) 조회
// GET /api/bs.json_array?section_id={id}&end_date={YYYYMMDD}
// 실제 응답 results: {"assets": {"total": N, "accounts": [...]}, "liabilities": {...}}
func (c *WhooingClient) GetBS(sectionID, endDate string) (*BSResponse, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("end_date", endDate)

	data, err := c.doRequest(http.MethodGet, "/bs.json_array", params)
	if err != nil {
		return nil, err
	}

	var resp BSResponse
	if err := parseResponseWithClient(c, data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateEntry는 거래 입력
// POST /api/entries.json (form-encoded body)
func (c *WhooingClient) CreateEntry(sectionID, entryDate, lAccount, lAccountID, rAccount, rAccountID, item, memo string, money float64) (*Entry, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("entry_date", entryDate)
	params.Set("l_account", lAccount)
	params.Set("l_account_id", lAccountID)
	params.Set("r_account", rAccount)
	params.Set("r_account_id", rAccountID)
	params.Set("money", fmt.Sprintf("%.0f", money))
	params.Set("item", item)
	if memo != "" {
		params.Set("memo", memo)
	}

	data, err := c.doRequest(http.MethodPost, "/entries.json", params)
	if err != nil {
		return nil, err
	}

	return c.parseEntryArrayResponse(data)
}

// UpdateEntry는 거래 수정 (변경된 필드만 전송)
func (c *WhooingClient) UpdateEntry(sectionID string, entryID int, fields map[string]string) (*Entry, error) {
	endpoint := fmt.Sprintf("/entries/%d.json", entryID)
	params := url.Values{}
	params.Set("section_id", sectionID)
	for k, v := range fields {
		params.Set(k, v)
	}

	data, err := c.doRequest(http.MethodPut, endpoint, params)
	if err != nil {
		return nil, err
	}

	// PUT 응답 results는 단일 객체 (POST 생성은 배열)
	return c.parseEntryObjectResponse(data)
}

// DeleteEntry는 거래 삭제
// DELETE /api/entries/{entry_id}/{section_id}.json
func (c *WhooingClient) DeleteEntry(sectionID string, entryID int) error {
	endpoint := fmt.Sprintf("/entries/%d/%s.json", entryID, sectionID)
	data, err := c.doRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}
	return parseResponseWithClient(c, data, nil)
}

// GetUser는 유저 정보 조회 (raw JSON 반환)
// GET /api/user.json
func (c *WhooingClient) GetUser() ([]byte, error) {
	return c.doRequest(http.MethodGet, "/user.json", nil)
}

// GetUserLogs는 유저 로그 리스트 조회 (raw JSON 반환)
// GET /api/user_logs.json
func (c *WhooingClient) GetUserLogs() ([]byte, error) {
	return c.doRequest(http.MethodGet, "/user_logs.json", nil)
}

// GetSectionsAll은 전체 섹션 목록 조회 (raw JSON 반환)
// GET /api/sections.json
func (c *WhooingClient) GetSectionsAll() ([]byte, error) {
	return c.doRequest(http.MethodGet, "/sections.json", nil)
}

// GetSection은 특정 섹션 조회 (raw JSON 반환)
// GET /api/sections/:section_id.json
func (c *WhooingClient) GetSection(sectionID string) ([]byte, error) {
	return c.doRequest(http.MethodGet, fmt.Sprintf("/sections/%s.json", sectionID), nil)
}

// GetSectionDefault는 기본 섹션 조회 (raw JSON 반환)
// GET /api/sections/default.json
func (c *WhooingClient) GetSectionDefault() ([]byte, error) {
	return c.doRequest(http.MethodGet, "/sections/default.json", nil)
}

// --- Accounts CLI용 (raw JSON 반환) ---

// GetAccountsList는 전체 항목 목록 조회
// GET /api/accounts.json?section_id=...
func (c *WhooingClient) GetAccountsList(sectionID string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	return c.doRequest(http.MethodGet, "/accounts.json", params)
}

// GetAccountsByType은 계정별 항목 목록 조회
// GET /api/accounts/:account.json?section_id=...
func (c *WhooingClient) GetAccountsByType(sectionID, account string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	return c.doRequest(http.MethodGet, fmt.Sprintf("/accounts/%s.json", account), params)
}

// GetAccountByID는 특정 항목 상세 조회
// GET /api/accounts/:account/:account_id.json?section_id=...
func (c *WhooingClient) GetAccountByID(sectionID, account, accountID string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	return c.doRequest(http.MethodGet, fmt.Sprintf("/accounts/%s/%s.json", account, accountID), params)
}

// --- Entries CLI용 (raw JSON 반환) ---

// GetEntriesSearch는 거래내역 조회
// GET /api/entries.json?section_id=...&start_date=...&end_date=...
func (c *WhooingClient) GetEntriesSearch(sectionID, startDate, endDate string, limit int) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("start_date", startDate)
	params.Set("end_date", endDate)
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	return c.doRequest(http.MethodGet, "/entries.json", params)
}

// GetEntryDetail은 특정 거래 조회
// GET /api/entries/:entry_id.json?section_id=...
func (c *WhooingClient) GetEntryDetail(sectionID, entryID string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	return c.doRequest(http.MethodGet, fmt.Sprintf("/entries/%s.json", entryID), params)
}

// GetLatestEntries는 최근 거래내역 조회
// GET /api/entries/latest.json?section_id=...
func (c *WhooingClient) GetLatestEntries(sectionID string, limit int) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	return c.doRequest(http.MethodGet, "/entries/latest.json", params)
}

// GetLatestItems는 최근 아이템 목록 조회 (Suggest용)
// GET /api/entries/latest_items.json?section_id=...
func (c *WhooingClient) GetLatestItems(sectionID string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	return c.doRequest(http.MethodGet, "/entries/latest_items.json", params)
}

