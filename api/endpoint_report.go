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

// reportAggregate는 report 응답 aggregate 중 BS 계정 부분
type reportAggregate struct {
	Assets      ReportGroup `json:"assets"`
	Liabilities ReportGroup `json:"liabilities"`
}

// reportResults는 report 응답의 results 구조 (BS 변환용 최소 필드)
type reportResults struct {
	Aggregate reportAggregate `json:"aggregate"`
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
	return c.doRequest(http.MethodGet, fmt.Sprintf("/report/%s.json", account), reportQueryToValues(q))
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

// ─── BS 변환 헬퍼 ────────────────────────────────────────────

// reportGroupToBSGroup은 report의 {total, accounts map}을 BSGroup으로 변환
// 항목 순서는 account_id 자연 정렬 (x1, x2, ..., x10)
func reportGroupToBSGroup(g ReportGroup) BSGroup {
	accounts := make([]BSAccount, 0, len(g.Accounts))
	for id, money := range g.Accounts {
		accounts = append(accounts, BSAccount{AccountID: id, Money: money})
	}
	sort.Slice(accounts, func(i, j int) bool {
		return accountIDLess(accounts[i].AccountID, accounts[j].AccountID)
	})
	return BSGroup{Total: g.Total, Accounts: accounts}
}

// accountIDLess는 "x12" 형태의 항목 ID를 숫자 기준으로 비교
func accountIDLess(a, b string) bool {
	na, aok := accountIDNumber(a)
	nb, bok := accountIDNumber(b)
	if aok && bok {
		if na != nb {
			return na < nb
		}
		return a < b
	}
	return a < b
}

// accountIDNumber는 항목 ID의 숫자 부분을 추출 ("x12" → 12)
func accountIDNumber(id string) (int, bool) {
	trimmed := strings.TrimLeftFunc(id, func(r rune) bool {
		return r < '0' || r > '9'
	})
	if trimmed == "" {
		return 0, false
	}
	n, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, false
	}
	return n, true
}

// parseReportAsBS는 report 응답 raw JSON을 BSResponse로 변환
func (c *WhooingClient) parseReportAsBS(data []byte) (*BSResponse, error) {
	var results reportResults
	if err := parseResponseWithClient(c, data, &results); err != nil {
		return nil, err
	}
	return &BSResponse{
		Assets:      reportGroupToBSGroup(results.Aggregate.Assets),
		Liabilities: reportGroupToBSGroup(results.Aggregate.Liabilities),
	}, nil
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
