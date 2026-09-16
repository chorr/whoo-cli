// api/endpoint_report.go
// Report / Report Summary 통합 보고서 엔드포인트
// 공식 문서 기준 bs.json, pl.json, daily_pl.json, zigzag*.json, mountain.json을 대체한다

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// ─── 데이터 모델 ──────────────────────────────────────────────

// ReportQuery는 report/report_summary 공통 파라미터
type ReportQuery struct {
	SectionID string
	AccountID string // 특정 항목만 조회 (예: x123)
	StartDate string // Ymd 또는 Ym (예: 20240101, 202401)
	EndDate   string
	RowsType  string // day|month|quarter|year|none (생략 시 month)
	Item      string // 항목명 필터 (와일드카드 * 지원)
}

// ReportGroup은 report 응답의 계정 타입별 {total, accounts} 구조
type ReportGroup struct {
	Total    float64            `json:"total"`
	Accounts map[string]float64 `json:"accounts"`
}

func reportQueryToValues(q ReportQuery) url.Values {
	p := url.Values{}
	p.Set("section_id", q.SectionID)
	if q.AccountID != "" {
		p.Set("account_id", q.AccountID)
	}
	if q.StartDate != "" {
		p.Set("start_date", q.StartDate)
	}
	if q.EndDate != "" {
		p.Set("end_date", q.EndDate)
	}
	if q.RowsType != "" {
		p.Set("rows_type", q.RowsType)
	}
	if q.Item != "" {
		p.Set("item", q.Item)
	}
	return p
}

// ─── Report API 메서드 ────────────────────────────────────────

// GetReport는 전체 계정 통합 보고서 조회 (raw JSON)
// GET /api/report.json
func (c *WhooingClient) GetReport(q ReportQuery) ([]byte, error) {
	return c.doRequest(http.MethodGet, "/report.json", reportQueryToValues(q))
}

// GetReportByAccount는 특정 계정 타입 보고서 조회 (raw JSON)
// GET /api/report/:account.json
// account: assets|liabilities|expenses|income|all (콤마로 복수 지정 가능)
func (c *WhooingClient) GetReportByAccount(account string, q ReportQuery) ([]byte, error) {
	accounts := strings.Split(account, ",")
	if len(accounts) > 1 {
		responses := make([][]byte, 0, len(accounts))
		for _, accountType := range accounts {
			accountType = strings.TrimSpace(accountType)
			if accountType == "" {
				continue
			}
			data, err := c.doRequest(
				http.MethodGet,
				fmt.Sprintf("/report/%s.json", accountType),
				reportQueryToValues(q),
			)
			if err != nil {
				return nil, err
			}
			responses = append(responses, data)
		}
		return c.mergeReportResponses(responses)
	}
	return c.doRequest(http.MethodGet, fmt.Sprintf("/report/%s.json", account), reportQueryToValues(q))
}

// mergeReportResponses는 WAF에서 콤마 경로를 거부하는 환경을 피하기 위해
// 계정별로 조회한 report 응답을 기존 다중 계정 응답 형태로 합친다.
func (c *WhooingClient) mergeReportResponses(responses [][]byte) ([]byte, error) {
	var root map[string]interface{}
	mergedResults := map[string]interface{}{}

	for i, data := range responses {
		var results map[string]interface{}
		if err := parseResponseWithClient(c, data, &results); err != nil {
			return nil, err
		}
		deepMergeJSON(mergedResults, results)

		if i == 0 {
			if err := json.Unmarshal(data, &root); err != nil {
				return nil, fmt.Errorf("report 응답 파싱 실패: %w", err)
			}
		}
	}

	if root == nil {
		return nil, fmt.Errorf("합칠 report 응답이 없습니다")
	}
	addDerivedReportValues(mergedResults)
	root["results"] = mergedResults
	root["rest_of_api"] = c.LastRestOfAPI()
	return json.Marshal(root)
}

func deepMergeJSON(dst, src map[string]interface{}) {
	for key, srcValue := range src {
		srcMap, srcOK := srcValue.(map[string]interface{})
		dstMap, dstOK := dst[key].(map[string]interface{})
		if srcOK && dstOK {
			deepMergeJSON(dstMap, srcMap)
			continue
		}
		dst[key] = srcValue
	}
}

// addDerivedReportValues는 분리 조회 시 서버가 계산하지 못하는 자본/순이익을 보완한다.
func addDerivedReportValues(node map[string]interface{}) {
	for _, value := range node {
		if child, ok := value.(map[string]interface{}); ok {
			addDerivedReportValues(child)
		}
	}

	if _, exists := node["capital"]; !exists {
		if assets, aok := reportTotal(node["assets"]); aok {
			if liabilities, lok := reportTotal(node["liabilities"]); lok {
				node["capital"] = map[string]interface{}{"total": assets - liabilities}
			}
		}
	}
	if _, exists := node["net_income"]; !exists {
		if income, iok := reportTotal(node["income"]); iok {
			if expenses, eok := reportTotal(node["expenses"]); eok {
				node["net_income"] = income - expenses
			}
		}
	}
}

func reportTotal(value interface{}) (float64, bool) {
	group, ok := value.(map[string]interface{})
	if !ok {
		return 0, false
	}
	total, ok := group["total"].(float64)
	return total, ok
}

type reportBalanceResults struct {
	Assets      *ReportGroup `json:"assets"`
	Liabilities *ReportGroup `json:"liabilities"`
	Aggregate   struct {
		Assets      *ReportGroup `json:"assets"`
		Liabilities *ReportGroup `json:"liabilities"`
	} `json:"aggregate"`
}

// parseReportAsBS는 rows_type=none의 직접 그룹과 aggregate 그룹을 모두 수용한다.
func (c *WhooingClient) parseReportAsBS(data []byte) (*BSResponse, error) {
	var results reportBalanceResults
	if err := parseResponseWithClient(c, data, &results); err != nil {
		return nil, err
	}
	assets := results.Assets
	liabilities := results.Liabilities
	if assets == nil {
		assets = results.Aggregate.Assets
	}
	if liabilities == nil {
		liabilities = results.Aggregate.Liabilities
	}
	if assets == nil || liabilities == nil {
		return nil, fmt.Errorf("report 응답에 자산/부채 잔액이 없습니다")
	}
	return &BSResponse{
		Assets:      reportGroupToBSGroup(*assets),
		Liabilities: reportGroupToBSGroup(*liabilities),
	}, nil
}

func reportGroupToBSGroup(group ReportGroup) BSGroup {
	accounts := make([]BSAccount, 0, len(group.Accounts))
	for accountID, money := range group.Accounts {
		accounts = append(accounts, BSAccount{AccountID: accountID, Money: money})
	}
	sort.Slice(accounts, func(i, j int) bool {
		return reportAccountIDLess(accounts[i].AccountID, accounts[j].AccountID)
	})
	return BSGroup{Total: group.Total, Accounts: accounts}
}

func reportAccountIDLess(a, b string) bool {
	number := func(accountID string) (int, bool) {
		digits := strings.TrimLeftFunc(accountID, func(r rune) bool {
			return r < '0' || r > '9'
		})
		value, err := strconv.Atoi(digits)
		return value, digits != "" && err == nil
	}
	aNumber, aOK := number(a)
	bNumber, bOK := number(b)
	if aOK && bOK && aNumber != bNumber {
		return aNumber < bNumber
	}
	return a < b
}

// GetReportSummary는 기간별 손익/자산 요약 조회 (raw JSON, flat 숫자 응답)
// GET /api/report_summary.json 또는 /api/report_summary/:account.json
// account가 빈 문자열이면 기본(expenses,income) 요약
func (c *WhooingClient) GetReportSummary(account string, q ReportQuery) ([]byte, error) {
	endpoint := "/report_summary.json"
	if account != "" {
		endpoint = fmt.Sprintf("/report_summary/%s.json", account)
	}
	return c.doRequest(http.MethodGet, endpoint, reportQueryToValues(q))
}

// ─── 유연한 results 파싱 헬퍼 ─────────────────────────────────

// flexibleGoalRows는 goal 응답 results를 배열/객체 모두 수용해 파싱
// 공식 문서: [{"date": "201101", "money": N}] / 과거 구현: {"201101": N}
func flexibleGoalRows(raw json.RawMessage) (GoalMap, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "" || trimmed == "null" {
		return GoalMap{}, nil
	}
	result := GoalMap{}
	if strings.HasPrefix(trimmed, "[") {
		var rows []struct {
			Date  FlexibleString `json:"date"`
			Money float64        `json:"money"`
		}
		if err := json.Unmarshal(raw, &rows); err != nil {
			return nil, fmt.Errorf("goal 배열 파싱 실패: %w", err)
		}
		for _, r := range rows {
			result[r.Date.String()] = int64(r.Money)
		}
		return result, nil
	}
	var m map[string]float64
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("goal 객체 파싱 실패: %w", err)
	}
	for ym, money := range m {
		result[ym] = int64(money)
	}
	return result, nil
}
