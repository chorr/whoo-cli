// cmd/inout.go
// inout CLI 커맨드 — 자금증감 조회

package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"whoo-cli/api"
	"whoo-cli/config"
)

// RunInOut는 inout CLI 커맨드 실행
func RunInOut(cfg *config.Config, args []string) {
	RequireAuth(cfg)
	RequireSection(cfg)

	fs := flag.NewFlagSet("inout", flag.ExitOnError)
	account := fs.String("account", "", "계정 필터 (assets | liabilities)")
	accountID := fs.String("account-id", "", "계정 항목 ID (--account 필수)")
	from := fs.String("from", "", "시작일 YYYYMMDD (기본: 이번달 1일)")
	to := fs.String("to", "", "종료일 YYYYMMDD (기본: 오늘)")
	fs.Parse(args)

	// 날짜 기본값: 이번달
	now := time.Now()
	if *from == "" {
		*from = fmt.Sprintf("%d%02d01", now.Year(), now.Month())
	}
	if *to == "" {
		*to = now.Format("20060102")
	}

	client := NewClient(cfg)
	resp, err := client.GetInOut(api.InOutQuery{
		SectionID: cfg.SectionID,
		StartDate: *from,
		EndDate:   *to,
		Account:   *account,
		AccountID: *accountID,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "[오류] %v\n", err)
		os.Exit(1)
	}

	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[오류] JSON 직렬화 실패: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
