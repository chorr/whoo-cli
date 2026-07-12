// cmd/flow_result.go
// 흐름 분석 결과 파싱 및 TUI 렌더 헬퍼

package cmd

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"whoo-cli/api"
)

// flow 결과 계정 타입 표시 순서
var flowAccountTypeOrder = []string{
	"assets", "liabilities", "capital", "income", "expenses",
}

// ─── 파싱 구조 ────────────────────────────────────────────────

type flowMoneyTriple struct {
	ID     string
	From   float64
	To     float64
	Margin float64
}

type flowGroupView struct {
	TypeCode string
	Total    flowMoneyTriple
	Accounts []flowMoneyTriple
}

type changesView struct {
	In      float64
	Out     float64
	Rows    []changesRow
	RowsTyp string
}

type changesRow struct {
	Date  string
	Money float64
}

// parseFlowGroups는 flow_of_account / flow_of_account_id 결과를 파싱한다.
// accounts 필드는 map(id→obj) 또는 배열을 모두 수용한다.
func parseFlowGroups(raw []byte) ([]flowGroupView, error) {
	var wrapper map[string]interface{}
	if err := parseJSONResponse(raw, &wrapper); err != nil {
		return nil, err
	}
	results, ok := wrapper["results"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("results 형식 오류")
	}

	var groups []flowGroupView
	// 문서/실제 응답 키 순회 (알려진 타입 우선, 그 외 키도 처리)
	seen := map[string]bool{}
	keys := append([]string{}, flowAccountTypeOrder...)
	for k := range results {
		if !seen[k] && k != "count" {
			// 이미 order에 있으면 스킵
			found := false
			for _, o := range flowAccountTypeOrder {
				if o == k {
					found = true
					break
				}
			}
			if !found {
				keys = append(keys, k)
			}
		}
	}
	for _, code := range keys {
		seen[code] = true
		gRaw, ok := results[code].(map[string]interface{})
		if !ok {
			continue
		}
		g := flowGroupView{TypeCode: code}
		if t, ok := gRaw["total"].(map[string]interface{}); ok {
			g.Total = parseTriple(t, "")
		}
		// accounts: map 또는 array
		switch accs := gRaw["accounts"].(type) {
		case map[string]interface{}:
			ids := make([]string, 0, len(accs))
			for id := range accs {
				ids = append(ids, id)
			}
			sort.Strings(ids)
			for _, id := range ids {
				if am, ok := accs[id].(map[string]interface{}); ok {
					tr := parseTriple(am, id)
					if tr.ID == "" {
						tr.ID = id
					}
					g.Accounts = append(g.Accounts, tr)
				}
			}
		case []interface{}:
			for _, elem := range accs {
				if am, ok := elem.(map[string]interface{}); ok {
					id := mapString(am, "account_id")
					g.Accounts = append(g.Accounts, parseTriple(am, id))
				}
			}
		}
		// 전부 그룹은 생략
		if g.Total.From == 0 && g.Total.To == 0 && g.Total.Margin == 0 && len(g.Accounts) == 0 {
			continue
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func parseTriple(m map[string]interface{}, fallbackID string) flowMoneyTriple {
	t := flowMoneyTriple{ID: fallbackID}
	if id := mapString(m, "account_id"); id != "" {
		t.ID = id
	}
	t.From = mapFloat(m, "from")
	t.To = mapFloat(m, "to")
	t.Margin = mapFloat(m, "margin")
	return t
}

func mapFloat(m map[string]interface{}, key string) float64 {
	if n, ok := m[key].(float64); ok {
		return n
	}
	return 0
}

// parseChangesView는 changes_of_* 결과를 파싱한다.
// rows는 map(date→money) 또는 [{date,money}] 배열을 수용한다.
func parseChangesView(raw []byte) (*changesView, error) {
	var wrapper map[string]interface{}
	if err := parseJSONResponse(raw, &wrapper); err != nil {
		return nil, err
	}
	results, ok := wrapper["results"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("results 형식 오류")
	}
	v := &changesView{}
	if agg, ok := results["aggregate"].(map[string]interface{}); ok {
		v.In = mapFloat(agg, "in")
		v.Out = mapFloat(agg, "out")
	}
	v.RowsTyp = mapString(results, "rows_type")
	switch rows := results["rows"].(type) {
	case map[string]interface{}:
		dates := make([]string, 0, len(rows))
		for d := range rows {
			dates = append(dates, d)
		}
		sort.Strings(dates)
		for _, d := range dates {
			money := 0.0
			switch mv := rows[d].(type) {
			case float64:
				money = mv
			case map[string]interface{}:
				money = mapFloat(mv, "money")
			}
			v.Rows = append(v.Rows, changesRow{Date: d, Money: money})
		}
	case []interface{}:
		for _, elem := range rows {
			if rm, ok := elem.(map[string]interface{}); ok {
				v.Rows = append(v.Rows, changesRow{
					Date:  mapString(rm, "date"),
					Money: mapFloat(rm, "money"),
				})
			}
		}
	}
	return v, nil
}

// ─── 렌더 ────────────────────────────────────────────────────

func resolveAccountTitle(am *api.AccountsMap, id string) string {
	if am == nil || id == "" {
		return id
	}
	for _, t := range flowAccountTypeOrder {
		if title := am.GetTitle(t, id); title != "" && title != id {
			return title
		}
	}
	return id
}

// renderFlowGroupsLines는 계정 흐름 테이블을 줄 단위로 생성
func renderFlowGroupsLines(groups []flowGroupView, am *api.AccountsMap) []string {
	var lines []string
	if len(groups) == 0 {
		return []string{"  (표시할 데이터가 없습니다)"}
	}
	// 헤더
	lines = append(lines, fmt.Sprintf("  %-14s  %12s  %12s  %12s", "항목", "From", "To", "Margin"))
	lines = append(lines, "  "+strings.Repeat("-", 56))

	for _, g := range groups {
		typeName := FormatAccount(g.TypeCode)
		if typeName == g.TypeCode {
			typeName = g.TypeCode
		}
		lines = append(lines, "")
		lines = append(lines, headerStyle.Render(fmt.Sprintf("  [%s]", typeName)))
		lines = append(lines, formatTripleLine("  합계", g.Total))

		// margin 절댓값 큰 순, 0은 뒤로
		accs := append([]flowMoneyTriple(nil), g.Accounts...)
		sort.SliceStable(accs, func(i, j int) bool {
			ai, aj := math.Abs(accs[i].Margin), math.Abs(accs[j].Margin)
			if ai != aj {
				return ai > aj
			}
			return accs[i].ID < accs[j].ID
		})
		shown := 0
		for _, a := range accs {
			if a.From == 0 && a.To == 0 && a.Margin == 0 {
				continue // 무변동 항목 숨김
			}
			title := resolveAccountTitle(am, a.ID)
			label := truncateRunes(title, 12)
			if a.ID != "" {
				label = fmt.Sprintf("%s(%s)", truncateRunes(title, 8), a.ID)
				label = truncateRunes(label, 14)
			}
			lines = append(lines, formatTripleLine("  "+label, a))
			shown++
		}
		if shown == 0 {
			lines = append(lines, "    (세부 항목 없음)")
		}
	}
	return lines
}

func formatTripleLine(label string, t flowMoneyTriple) string {
	return fmt.Sprintf("%-16s  %12s  %12s  %12s",
		label,
		FormatMoney(t.From),
		FormatMoney(t.To),
		FormatMoney(t.Margin),
	)
}

// renderChangesLines는 일일 변동 테이블 줄 생성
func renderChangesLines(v *changesView) []string {
	if v == nil {
		return []string{"  (표시할 데이터가 없습니다)"}
	}
	var lines []string
	lines = append(lines, fmt.Sprintf("  합계  In %s   Out %s   순증감 %s",
		FormatMoney(v.In),
		FormatMoney(v.Out),
		FormatMoney(v.In-v.Out),
	))
	if v.RowsTyp != "" {
		lines = append(lines, helpStyle.Render(fmt.Sprintf("  단위: %s", v.RowsTyp)))
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %-12s  %14s  %s", "날짜", "금액", "추이"))
	lines = append(lines, "  "+strings.Repeat("-", 40))

	maxAbs := 0.0
	for _, r := range v.Rows {
		if a := math.Abs(r.Money); a > maxAbs {
			maxAbs = a
		}
	}
	for _, r := range v.Rows {
		date := r.Date
		if len(date) == 8 {
			date = FormatDate(date)
		}
		bar := moneyBar(r.Money, maxAbs, 16)
		lines = append(lines, fmt.Sprintf("  %-12s  %14s  %s", date, FormatMoney(r.Money), bar))
	}
	if len(v.Rows) == 0 {
		lines = append(lines, "  (일별 데이터 없음)")
	}
	return lines
}

func moneyBar(v, maxAbs float64, width int) string {
	if maxAbs == 0 || width <= 0 {
		return ""
	}
	n := int(math.Round(math.Abs(v) / maxAbs * float64(width)))
	if n < 0 {
		n = 0
	}
	if n > width {
		n = width
	}
	if v > 0 {
		return strings.Repeat("+", n)
	}
	if v < 0 {
		return strings.Repeat("-", n)
	}
	return ""
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	if max <= 1 {
		return string(r[:max])
	}
	return string(r[:max-1]) + "…"
}

// sliceViewport는 스크롤 오프셋 기준 화면 줄 슬라이스
func sliceViewport(lines []string, offset, maxVisible int) (view []string, newOffset int) {
	if maxVisible <= 0 {
		maxVisible = 18
	}
	if len(lines) == 0 {
		return nil, 0
	}
	if offset < 0 {
		offset = 0
	}
	maxOff := len(lines) - maxVisible
	if maxOff < 0 {
		maxOff = 0
	}
	if offset > maxOff {
		offset = maxOff
	}
	end := offset + maxVisible
	if end > len(lines) {
		end = len(lines)
	}
	return lines[offset:end], offset
}
