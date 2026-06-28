// api/params.go
// 공통 파라미터 빌더

package api

import (
	"fmt"
	"net/url"
)

// Params는 API 요청 파라미터 빌더
type Params struct {
	v url.Values
}

// NewParams는 빈 파라미터 빌더 생성
func NewParams() *Params {
	return &Params{v: url.Values{}}
}

// Section은 section_id 파라미터 설정
func (p *Params) Section(sectionID string) *Params {
	p.v.Set("section_id", sectionID)
	return p
}

// DateRange는 start_date / end_date 파라미터 설정 (YYYYMMDD 형식)
func (p *Params) DateRange(start, end string) *Params {
	p.v.Set("start_date", start)
	p.v.Set("end_date", end)
	return p
}

// Month는 month 파라미터 설정 (YYYYMM 형식)
func (p *Params) Month(yyyymm string) *Params {
	p.v.Set("month", yyyymm)
	return p
}

// Limit는 limit 파라미터 설정
func (p *Params) Limit(n int) *Params {
	if n > 0 {
		p.v.Set("limit", fmt.Sprintf("%d", n))
	}
	return p
}

// Str은 임의 문자열 파라미터 설정
func (p *Params) Str(key, value string) *Params {
	p.v.Set(key, value)
	return p
}

// Values는 내부 url.Values 반환
func (p *Params) Values() url.Values {
	return p.v
}
