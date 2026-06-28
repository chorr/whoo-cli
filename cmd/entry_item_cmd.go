// cmd/entry_item_cmd.go
// 거래 아이템 명령어 파서 (//n 할부, **n 반복, (detail) 메모, (n%) 수수료)

package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ItemCommand는 파싱된 아이템 명령어 구조
type ItemCommand struct {
	Base   string  // 기본 아이템명
	Detail string  // 괄호 메모 (예: "서울 강남")
	Repeat int     // **n → n (반복 입력 횟수)
	Split  int     // //n → n (할부 개월 수)
	Fee    float64 // (n%) 수수료율 (0이면 미적용)
}

// 파서에서 사용하는 정규식
var (
	reSplit  = regexp.MustCompile(`//(\d+)`)
	reRepeat = regexp.MustCompile(`\*\*(\d+)`)
	reFee    = regexp.MustCompile(`\((\d+(?:\.\d+)?)%\)`)
	reDetail = regexp.MustCompile(`\(([^%][^)]*)\)`)
)

// ParseItemCommand는 아이템 문자열을 파싱하여 ItemCommand 반환
// 규칙:
//   - "아이템//3"       → Split=3
//   - "아이템**12"      → Repeat=12
//   - "노트북//3(28%)"  → Split=3, Fee=28.0
//   - "아이템(서울)"    → Detail="서울"
//   - 명령어는 POST 전용. PUT에서는 파서 무시
func ParseItemCommand(s string) (*ItemCommand, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("아이템이 비어있습니다")
	}

	cmd := &ItemCommand{}
	rest := s

	// 1. 수수료율 추출: (n%) — Split 이후에 나타나는 경우
	if m := reFee.FindStringIndex(rest); m != nil {
		feeStr := reFee.FindStringSubmatch(rest)[1]
		fee, err := strconv.ParseFloat(feeStr, 64)
		if err != nil {
			return nil, fmt.Errorf("수수료율 파싱 실패: %w", err)
		}
		cmd.Fee = fee
		rest = rest[:m[0]] + rest[m[1]:]
	}

	// 2. 할부 //n 추출
	if m := reSplit.FindStringSubmatch(rest); m != nil {
		n, err := strconv.Atoi(m[1])
		if err != nil || n <= 1 {
			return nil, fmt.Errorf("할부 개월 수는 2 이상이어야 합니다")
		}
		cmd.Split = n
		rest = reSplit.ReplaceAllString(rest, "")
	}

	// 3. 반복 **n 추출
	if m := reRepeat.FindStringSubmatch(rest); m != nil {
		n, err := strconv.Atoi(m[1])
		if err != nil || n <= 1 {
			return nil, fmt.Errorf("반복 횟수는 2 이상이어야 합니다")
		}
		cmd.Repeat = n
		rest = reRepeat.ReplaceAllString(rest, "")
	}

	// 4. 할부와 반복은 동시 사용 불가
	if cmd.Split > 0 && cmd.Repeat > 0 {
		return nil, fmt.Errorf("할부(//)와 반복(**)은 동시에 사용할 수 없습니다")
	}

	// 5. 수수료는 할부에만 적용
	if cmd.Fee > 0 && cmd.Split == 0 {
		return nil, fmt.Errorf("수수료율(%%)은 할부(//) 명령어에만 사용할 수 있습니다")
	}

	// 6. 괄호 메모 추출: (내용) — % 없는 괄호
	if m := reDetail.FindStringSubmatch(rest); m != nil {
		cmd.Detail = strings.TrimSpace(m[1])
		rest = reDetail.ReplaceAllString(rest, "")
	}

	// 7. 나머지가 기본 아이템명
	cmd.Base = strings.TrimSpace(rest)
	if cmd.Base == "" {
		return nil, fmt.Errorf("아이템명이 비어있습니다")
	}

	return cmd, nil
}

// String은 ItemCommand를 API 전송용 아이템 문자열로 재직렬화
// 기본명(메모) + 명령어 형태
func (c *ItemCommand) String() string {
	var b strings.Builder
	b.WriteString(c.Base)
	if c.Detail != "" {
		b.WriteString(fmt.Sprintf("(%s)", c.Detail))
	}
	if c.Split > 0 {
		b.WriteString(fmt.Sprintf("//%d", c.Split))
		if c.Fee > 0 {
			// 정수면 정수로, 소수점 있으면 소수점 표시
			if c.Fee == float64(int(c.Fee)) {
				b.WriteString(fmt.Sprintf("(%d%%)", int(c.Fee)))
			} else {
				b.WriteString(fmt.Sprintf("(%.2f%%)", c.Fee))
			}
		}
	}
	if c.Repeat > 0 {
		b.WriteString(fmt.Sprintf("**%d", c.Repeat))
	}
	return b.String()
}

// Preview는 ItemCommand를 사람이 읽기 쉬운 미리보기 문자열로 반환
func (c *ItemCommand) Preview() string {
	var parts []string
	parts = append(parts, c.Base)
	if c.Detail != "" {
		parts = append(parts, fmt.Sprintf("메모: %s", c.Detail))
	}
	if c.Split > 0 {
		s := fmt.Sprintf("%d개월 할부", c.Split)
		if c.Fee > 0 {
			s += fmt.Sprintf(" (수수료 %.1f%%)", c.Fee)
		}
		parts = append(parts, s)
	}
	if c.Repeat > 0 {
		parts = append(parts, fmt.Sprintf("%d회 반복", c.Repeat))
	}
	return strings.Join(parts, " / ")
}
