// cmd/inout.go
// inout CLI 커맨드 — 자금증감 조회 (raw JSON)

package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"whoo-cli/api"
	"whoo-cli/config"
)

// RunInOut는 inout CLI 커맨드 실행
func RunInOut(cfg *config.Config, args []string) {
	if wantsHelp(args) {
		showInOutHelp()
		return
	}

	RequireAuth(cfg)
	RequireSection(cfg)

	fs := flag.NewFlagSet("inout", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = showInOutHelp
	account := fs.String("account", "", "계정 필터 (assets | liabilities)")
	accountID := fs.String("account-id", "", "계정 항목 ID (--account 필수)")
	from := fs.String("from", "", "시작일 YYYYMMDD (기본: 이번달 1일)")
	to := fs.String("to", "", "종료일 YYYYMMDD (기본: 오늘)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	now := time.Now()
	if *from == "" {
		*from = fmt.Sprintf("%d%02d01", now.Year(), now.Month())
	}
	if *to == "" {
		*to = now.Format("20060102")
	}

	client := NewClient(cfg)
	data, err := client.GetInOutRaw(api.InOutQuery{
		SectionID: cfg.SectionID,
		StartDate: *from,
		EndDate:   *to,
		Account:   *account,
		AccountID: *accountID,
	})
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

func showInOutHelp() {
	fmt.Println("사용법: whoo inout [플래그]")
	fmt.Println()
	fmt.Println("기간 내 자산/부채 자금증감(in/out/margin)을 JSON으로 출력합니다.")
	fmt.Println("항목 메타만 보려면 whoo accounts, 잔액은 whoo bs 를 사용하세요.")
	fmt.Println()
	fmt.Println("플래그:")
	fmt.Println("  --account      계정 필터 (assets | liabilities)")
	fmt.Println("  --account-id   계정 항목 ID (--account 필수)")
	fmt.Println("  --from         시작일 YYYYMMDD (기본: 이번달 1일)")
	fmt.Println("  --to           종료일 YYYYMMDD (기본: 오늘)")
	fmt.Println("  -h, --help     도움말")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whoo inout")
	fmt.Println("  whoo inout --from 20260801 --to 20260813")
	fmt.Println("  whoo inout --account assets")
}
