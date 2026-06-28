// api/endpoint_inout.go
// In/Out (자금증감) 엔드포인트

package api

import (
	"fmt"
	"net/http"
	"net/url"
)

// ─── 데이터 모델 ──────────────────────────────────────────────

// InOutQuery는 in_out 조회 파라미터
type InOutQuery struct {
	SectionID string
	StartDate string // YYYYMMDD, 빈 문자열이면 생략
	EndDate   string // YYYYMMDD, 빈 문자열이면 생략
	Account   string // "" | "assets" | "liabilities"
	AccountID string // 비어있으면 해당 계정 전체 조회
}

// InOutResponse는 전체 또는 계정별 in_out 응답 구조
type InOutResponse struct {
	Assets      InOutGroup `json:"assets"`
	Liabilities InOutGroup `json:"liabilities"`
}

// InOutGroup은 in_out 응답의 계정 타입별 그룹
type InOutGroup struct {
	Total    InOutValues    `json:"total"`
	Accounts []InOutAccount `json:"accounts"`
}

// InOutAccount는 개별 계정 항목의 증감 데이터
type InOutAccount struct {
	AccountID string `json:"account_id"`
	In        int64  `json:"in"`
	Out       int64  `json:"out"`
	Margin    int64  `json:"margin"`
}

// InOutValues는 in/out/margin 집계값
type InOutValues struct {
	In     int64 `json:"in"`
	Out    int64 `json:"out"`
	Margin int64 `json:"margin"`
}

// ─── API 메서드 ───────────────────────────────────────────────

// GetInOut는 자금증감 조회
// Account/AccountID 유무에 따라 URL 경로를 조립:
//   - "" / ""           → /in_out.json
//   - "assets" / ""     → /in_out/assets.json
//   - "assets" / "x1"   → /in_out/assets/x1.json
func (c *WhooingClient) GetInOut(q InOutQuery) (*InOutResponse, error) {
	endpoint := buildInOutEndpoint(q)

	params := url.Values{}
	params.Set("section_id", q.SectionID)
	if q.StartDate != "" {
		params.Set("start_date", q.StartDate)
	}
	if q.EndDate != "" {
		params.Set("end_date", q.EndDate)
	}

	data, err := c.doRequest(http.MethodGet, endpoint, params)
	if err != nil {
		return nil, err
	}

	// 단일 항목 응답(AccountID 지정)은 InOutValues 구조이므로
	// Assets.Total에 담아 공통 구조로 변환
	if q.AccountID != "" {
		var single InOutValues
		if err := parseResponseWithClient(c, data, &single); err != nil {
			return nil, err
		}
		return &InOutResponse{
			Assets: InOutGroup{
				Total: single,
				Accounts: []InOutAccount{{
					AccountID: q.AccountID,
					In:        single.In,
					Out:       single.Out,
					Margin:    single.Margin,
				}},
			},
		}, nil
	}

	// 계정 지정(Account != "") 응답: 해당 계정 그룹만 반환
	// 전체(Account == "") 응답: assets + liabilities 포함
	var resp InOutResponse
	if err := parseResponseWithClient(c, data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// buildInOutEndpoint는 InOutQuery 조합에 따라 API 경로 생성
func buildInOutEndpoint(q InOutQuery) string {
	if q.Account == "" {
		return "/in_out.json"
	}
	if q.AccountID == "" {
		return fmt.Sprintf("/in_out/%s.json", q.Account)
	}
	return fmt.Sprintf("/in_out/%s/%s.json", q.Account, q.AccountID)
}
