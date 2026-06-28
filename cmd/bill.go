// cmd/bill.go
// bill CLI 커맨드 — 신용카드 청구내역 조회

package cmd

import (
	"flag"
	"fmt"
	"os"

	"whoo-cli/api"
	"whoo-cli/config"
)

// RunBill은 bill CLI 커맨드 실행
func RunBill(cfg *config.Config, args []string) {
	RequireAuth(cfg)
	RequireSection(cfg)

	fs := flag.NewFlagSet("bill", flag.ExitOnError)
	accountID := fs.String("account-id", "", "카드 항목 ID (생략 시 전체)")
	from := fs.Int("from", 0, "시작 연월 YYYYMM (예: 202601)")
	to := fs.Int("to", 0, "종료 연월 YYYYMM (예: 202612)")
	if err := fs.Parse(args); err != nil {
		return
	}

	if (*from == 0) != (*to == 0) {
		fmt.Fprintln(os.Stderr, "[오류] --from 과 --to 는 함께 지정해야 합니다")
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.GetBill(api.CardQuery{
		SectionID: cfg.SectionID,
		StartYM:   *from,
		EndYM:     *to,
		AccountID: *accountID,
	})
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

func showBillHelp() {
	fmt.Println("사용법: whoo bill [옵션]")
	fmt.Println()
	fmt.Println("신용카드 청구내역 조회")
	fmt.Println()
	fmt.Println("옵션:")
	fmt.Println("  --account-id    카드 항목 ID (생략 시 전체)")
	fmt.Println("  --from          시작 연월 YYYYMM (예: 202601)")
	fmt.Println("  --to            종료 연월 YYYYMM (예: 202612)")
}
