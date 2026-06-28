// api/endpoint_card.go
// Bill / Checkcard 엔드포인트

package api

import (
	"fmt"
	"net/http"
	"net/url"
)

// CardQuery는 bill/checkcard 조회 파라미터
type CardQuery struct {
	SectionID string
	StartYM   int    // YYYYMM (예: 202601), 0이면 생략
	EndYM     int    // YYYYMM (예: 202612), 0이면 생략
	AccountID string // 빈 문자열이면 전체 조회
}

// ─── 데이터 모델 ──────────────────────────────────────────────

// BillAggregateAccount는 bill aggregate 내 카드별 집계
type BillAggregateAccount struct {
	AccountID    string `json:"account_id"`
	Money        int64  `json:"money"`
	StartUseDate int    `json:"start_use_date"`
	EndUseDate   int    `json:"end_use_date"`
	PayDate      int    `json:"pay_date"`       // 결제일 (1~31, 일 단위)
	PayAccountID string `json:"pay_account_id"` // 결제 자산 항목 ID
}

// BillAccountRow는 월별 행 내 카드별 금액 행
type BillAccountRow struct {
	AccountID string `json:"account_id"`
	Money     int64  `json:"money"`
}

// BillMonthRow는 월별 행
type BillMonthRow struct {
	Total    int64            `json:"total"`
	Accounts []BillAccountRow `json:"accounts"`
}

// BillResponse는 bill 전체 조회(섹션) 응답 results 구조체
// GET /bill.json — aggregate.accounts 배열 형태
type BillResponse struct {
	Aggregate struct {
		Total    int64                  `json:"total"`
		Accounts []BillAggregateAccount `json:"accounts"`
	} `json:"aggregate"`
	RowsType string                  `json:"rows_type"`
	Rows     map[string]BillMonthRow `json:"rows"`
}

// BillSingleResponse는 bill 단일 카드 조회 응답 results 구조체
// GET /bill/:account_id.json — aggregate가 단일 객체 형태
type BillSingleResponse struct {
	Aggregate BillAggregateAccount `json:"aggregate"`
	RowsType  string               `json:"rows_type"`
	Rows      map[string]struct {
		AccountID    string `json:"account_id"`
		Money        int64  `json:"money"`
		StartUseDate int    `json:"start_use_date"`
		EndUseDate   int    `json:"end_use_date"`
		PayDate      int    `json:"pay_date"`
		PayAccountID string `json:"pay_account_id"`
	} `json:"rows"`
}

// CheckcardAggregateAccount는 checkcard aggregate 내 카드별 집계
type CheckcardAggregateAccount struct {
	AccountID string `json:"account_id"`
	Money     int64  `json:"money"`
}

// CheckcardRow는 월별 행
type CheckcardRow struct {
	Total    int64                       `json:"total"`
	Accounts []CheckcardAggregateAccount `json:"accounts"`
}

// CheckcardResponse는 checkcard 엔드포인트 응답 results 구조체
type CheckcardResponse struct {
	Aggregate struct {
		Total    int64                       `json:"total"`
		Accounts []CheckcardAggregateAccount `json:"accounts"`
	} `json:"aggregate"`
	RowsType string                    `json:"rows_type"`
	Rows     map[string]CheckcardRow   `json:"rows"`
}

// ─── 헬퍼 ─────────────────────────────────────────────────────

func cardQueryToValues(q CardQuery) url.Values {
	p := url.Values{}
	p.Set("section_id", q.SectionID)
	if q.StartYM > 0 {
		p.Set("start_date", fmt.Sprintf("%d", q.StartYM))
	}
	if q.EndYM > 0 {
		p.Set("end_date", fmt.Sprintf("%d", q.EndYM))
	}
	return p
}

// ─── Bill ─────────────────────────────────────────────────────

// GetBill은 신용카드 청구 내역 조회
// GET /api/bill.json 또는 /api/bill/:account_id.json
func (c *WhooingClient) GetBill(q CardQuery) ([]byte, error) {
	var path string
	if q.AccountID != "" {
		path = fmt.Sprintf("/bill/%s.json", q.AccountID)
	} else {
		path = "/bill.json"
	}
	return c.doRequest(http.MethodGet, path, cardQueryToValues(q))
}

// ─── Checkcard ────────────────────────────────────────────────

// GetCheckcard는 체크카드 사용 내역 조회
// GET /api/checkcard.json 또는 /api/checkcard/:account_id.json
func (c *WhooingClient) GetCheckcard(q CardQuery) ([]byte, error) {
	var path string
	if q.AccountID != "" {
		path = fmt.Sprintf("/checkcard/%s.json", q.AccountID)
	} else {
		path = "/checkcard.json"
	}
	return c.doRequest(http.MethodGet, path, cardQueryToValues(q))
}
