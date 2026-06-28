// api/endpoint_budget.go
// Budget / Budget Goal / Goal 엔드포인트

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ─── 데이터 모델 ──────────────────────────────────────────────

// BudgetLine은 예산/실적/잔여 집계값
type BudgetLine struct {
	Budget  int64 `json:"budget"`
	Money   int64 `json:"money"`
	Remains int64 `json:"remains"`
}

// BudgetLineWithID는 계정 ID를 포함한 예산/실적/잔여
type BudgetLineWithID struct {
	AccountID string `json:"account_id"`
	Budget    int64  `json:"budget"`
	Money     int64  `json:"money"`
	Remains   int64  `json:"remains"`
}

// BudgetMisc는 예산 부가 정보
type BudgetMisc struct {
	DailyRemains  int64      `json:"daily_remains"`
	WeeklyRemains int64      `json:"weekly_remains"`
	Standard      int64      `json:"standard"`
	Possibility   float64    `json:"possibility"`
	Today         BudgetLine `json:"today"`
}

// BudgetAggregate는 집계 데이터
// accounts 는 map[account_id]BudgetLineWithID 형태
type BudgetAggregate struct {
	Total         BudgetLine                  `json:"total"`
	TotalSteady   BudgetLine                  `json:"total_steady"`
	TotalFloating BudgetLine                  `json:"total_floating"`
	Misc          BudgetMisc                  `json:"misc"`
	Accounts      map[string]BudgetLineWithID `json:"accounts"`
}

// BudgetMonthRow는 월별 집계 행
type BudgetMonthRow struct {
	Date          int                         `json:"date"`
	Total         BudgetLine                  `json:"total"`
	TotalSteady   BudgetLine                  `json:"total_steady"`
	TotalFloating BudgetLine                  `json:"total_floating"`
	Misc          BudgetMisc                  `json:"misc"`
	Accounts      map[string]BudgetLineWithID `json:"accounts"`
}

// BudgetResponse는 budget API 응답의 results 구조
type BudgetResponse struct {
	Aggregate BudgetAggregate           `json:"aggregate"`
	RowsType  string                    `json:"rows_type"`
	Rows      map[string]BudgetMonthRow `json:"rows"`
}

// BudgetGoalResponse는 budget_goal API 응답의 results 구조
type BudgetGoalResponse struct {
	SetID        int       `json:"set_id"`
	LastModified int64     `json:"last_modified"`
	BaseYM       int       `json:"base_ym"`
	GoalYM       int       `json:"goal_ym"`
	BaseMoney    int64     `json:"base_money"`
	GoalMoney    int64     `json:"goal_money"`
	BaseIncome   int64     `json:"base_income"`
	BaseExpenses int64     `json:"base_expenses"`
	EachMonths   [][]int64 `json:"each_months"`
	SplitType    string    `json:"split_type"`
}

// BudgetGoalParams는 budget_goal PUT 파라미터
type BudgetGoalParams struct {
	BaseYM       int
	GoalYM       int
	GoalMoney    int64
	BaseMoney    int64
	BaseIncome   int64
	BaseExpenses int64
	EachMonths   [][]int64 // [[수입 12개월], [지출 12개월]]
	SplitType    string    // auto | equal | manual
}

// GoalMap은 goal API 응답의 results 구조 (YYYYMM → 목표 자본)
type GoalMap map[string]int64

// ─── Budget API 메서드 ────────────────────────────────────────

// GetBudget는 월별 계정 예산 대비 실적 조회
// GET /api/budget/:account.json
func (c *WhooingClient) GetBudget(sectionID, account string, startYM, endYM int) (*BudgetResponse, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("start_ym", strconv.Itoa(startYM))
	params.Set("end_ym", strconv.Itoa(endYM))

	data, err := c.doRequest(http.MethodGet, fmt.Sprintf("/budget/%s.json", account), params)
	if err != nil {
		return nil, err
	}

	var resp BudgetResponse
	if err := parseResponseWithClient(c, data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateBudget는 특정 월의 항목별 예산 수정
// PUT /api/budget/:account.json
// accountBudgets: map[account_id]budget_amount
func (c *WhooingClient) UpdateBudget(sectionID, account string, targetYM int, accountBudgets map[string]int64) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("target_ym", strconv.Itoa(targetYM))
	for id, amount := range accountBudgets {
		params.Set(id, strconv.FormatInt(amount, 10))
	}
	return c.doRequest(http.MethodPut, fmt.Sprintf("/budget/%s.json", account), params)
}

// UpdateBudgetBasicTotal는 장기목표용 월별 총액 일괄 수정
// PUT /api/budget/:account/basic_total.json
// monthly: 12개 월 총액 ([0]=1월 ... [11]=12월)
func (c *WhooingClient) UpdateBudgetBasicTotal(sectionID, account string, startYM, endYM int, monthly [12]int64) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("start_ym", strconv.Itoa(startYM))
	params.Set("end_ym", strconv.Itoa(endYM))
	for i, v := range monthly {
		params.Set(strconv.Itoa(i+1), strconv.FormatInt(v, 10))
	}
	return c.doRequest(http.MethodPut, fmt.Sprintf("/budget/%s/basic_total.json", account), params)
}

// DeleteBudget는 기간 내 예산 리셋
// DELETE /api/budget/:account.json
func (c *WhooingClient) DeleteBudget(sectionID, account string, startYM, endYM int) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("start_ym", strconv.Itoa(startYM))
	params.Set("end_ym", strconv.Itoa(endYM))
	return c.doRequest(http.MethodDelete, fmt.Sprintf("/budget/%s.json", account), params)
}

// ─── Budget Goal API 메서드 ───────────────────────────────────

// GetBudgetGoal는 장기 예산목표 조회
// GET /api/budget_goal.json
func (c *WhooingClient) GetBudgetGoal(sectionID string) (*BudgetGoalResponse, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)

	data, err := c.doRequest(http.MethodGet, "/budget_goal.json", params)
	if err != nil {
		return nil, err
	}

	var resp BudgetGoalResponse
	if err := parseResponseWithClient(c, data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateBudgetGoal는 장기 예산목표 설정/수정
// PUT /api/budget_goal.json
func (c *WhooingClient) UpdateBudgetGoal(sectionID string, p BudgetGoalParams) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("base_ym", strconv.Itoa(p.BaseYM))
	params.Set("goal_ym", strconv.Itoa(p.GoalYM))
	params.Set("goal_money", strconv.FormatInt(p.GoalMoney, 10))
	if p.BaseMoney != 0 {
		params.Set("base_money", strconv.FormatInt(p.BaseMoney, 10))
	}
	if p.BaseIncome != 0 {
		params.Set("base_income", strconv.FormatInt(p.BaseIncome, 10))
	}
	if p.BaseExpenses != 0 {
		params.Set("base_expenses", strconv.FormatInt(p.BaseExpenses, 10))
	}
	if p.SplitType != "" {
		params.Set("split_type", p.SplitType)
	}
	if len(p.EachMonths) > 0 {
		b, err := json.Marshal(p.EachMonths)
		if err == nil {
			params.Set("each_months", string(b))
		}
	}
	return c.doRequest(http.MethodPut, "/budget_goal.json", params)
}

// DeleteBudgetGoal는 장기 예산목표 + 모든 goal/budget 초기화 (복구 불가)
// DELETE /api/budget_goal.json
func (c *WhooingClient) DeleteBudgetGoal(sectionID string) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	return c.doRequest(http.MethodDelete, "/budget_goal.json", params)
}

// ─── Goal API 메서드 ──────────────────────────────────────────

// GetGoal는 월별 자본 도달 목표 조회
// GET /api/goal.json
func (c *WhooingClient) GetGoal(sectionID string, startYM, endYM int) (GoalMap, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	params.Set("start_ym", strconv.Itoa(startYM))
	params.Set("end_ym", strconv.Itoa(endYM))

	data, err := c.doRequest(http.MethodGet, "/goal.json", params)
	if err != nil {
		return nil, err
	}

	var resp GoalMap
	if err := parseResponseWithClient(c, data, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdateGoal는 월별 자본 목표 수정
// PUT /api/goal.json
// monthlyGoals: map[YYYYMM]money
func (c *WhooingClient) UpdateGoal(sectionID string, monthlyGoals map[int]int64) ([]byte, error) {
	params := url.Values{}
	params.Set("section_id", sectionID)
	for ym, money := range monthlyGoals {
		params.Set(strconv.Itoa(ym), strconv.FormatInt(money, 10))
	}
	return c.doRequest(http.MethodPut, "/goal.json", params)
}

// ─── 헬퍼 ────────────────────────────────────────────────────

// ParseBudgetAccountCSV는 "id=amount,id2=amount2" 형식을 파싱
func ParseBudgetAccountCSV(s string) (map[string]int64, error) {
	result := make(map[string]int64)
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("잘못된 형식: %q (id=amount 필요)", pair)
		}
		amount, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("금액 파싱 오류 %q: %w", parts[1], err)
		}
		result[strings.TrimSpace(parts[0])] = amount
	}
	return result, nil
}
