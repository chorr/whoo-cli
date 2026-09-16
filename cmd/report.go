// cmd/report.go
// report CLI 커맨드 — 통합 보고서 조회 (JSON 출력)
// 레거시 bs/pl/daily_pl/zigzag/mountain API를 대체하는 report / report_summary 사용

package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"whoo-cli/api"
	"whoo-cli/config"
)

// 유효한 계정 타입 (콤마로 복수 지정 가능)
var reportAccounts = map[string]bool{
	"assets": true, "liabilities": true, "expenses": true, "income": true, "all": true,
}

// RunReport는 report CLI 커맨드 실행
func RunReport(cfg *config.Config, args []string) {
	if wantsHelp(args) {
		showReportHelp()
		return
	}
	RequireAuth(cfg)
	RequireSection(cfg)

	summary := false
	if len(args) > 0 && args[0] == "summary" {
		summary = true
		args = args[1:]
	}

	// 첫 positional 인자가 계정 타입이면 분리 (assets,liabilities 등)
	account := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		account = args[0]
		args = args[1:]
		for _, a := range strings.Split(account, ",") {
			if !reportAccounts[a] {
				PrintError("유효하지 않은 계정 타입: %s (assets|liabilities|expenses|income|all)", a)
				os.Exit(1)
			}
		}
	}

	fs := flag.NewFlagSet("report", flag.ExitOnError)
	from := fs.String("from", "", "시작 날짜 YYYYMMDD 또는 YYYYMM")
	to := fs.String("to", "", "종료 날짜 YYYYMMDD 또는 YYYYMM")
	rowsType := fs.String("rows-type", "", "행 단위 day|month|quarter|year|none (기본: month)")
	accountID := fs.String("account-id", "", "특정 항목만 조회 (예: x12)")
	item := fs.String("item", "", "아이템 필터 (* 와일드카드)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	q := api.ReportQuery{
		SectionID: cfg.SectionID,
		AccountID: *accountID,
		StartDate: *from,
		EndDate:   *to,
		RowsType:  *rowsType,
		Item:      *item,
	}

	client := NewClient(cfg)
	var data []byte
	var err error
	switch {
	case summary:
		data, err = client.GetReportSummary(account, q)
	case account != "":
		data, err = client.GetReportByAccount(account, q)
	default:
		data, err = client.GetReport(q)
	}
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

func showReportHelp() {
	fmt.Println("사용법: whoo report [summary] [계정] [플래그]")
	fmt.Println()
	fmt.Println("통합 보고서를 JSON으로 출력합니다.")
	fmt.Println("(레거시 bs/pl/daily_pl/zigzag/mountain API를 대체합니다)")
	fmt.Println("콤마로 지정한 계정은 계정별 API를 호출한 뒤 같은 JSON 구조로 병합합니다.")
	fmt.Println("현재 잔액만 필요하면 403 영향을 받지 않는 whoo bs를 사용하세요.")
	fmt.Println()
	fmt.Println("커맨드:")
	fmt.Println("  (없음)               전체 계정 보고서 (rows + aggregate)")
	fmt.Println("  <계정>               특정 계정 보고서 (콤마로 복수 지정)")
	fmt.Println("  summary [계정]       기간 요약 (flat 숫자, 기본: expenses,income)")
	fmt.Println()
	fmt.Println("계정: assets | liabilities | expenses | income | all")
	fmt.Println()
	fmt.Println("플래그:")
	fmt.Println("  --from        시작 날짜 YYYYMMDD 또는 YYYYMM")
	fmt.Println("  --to          종료 날짜 YYYYMMDD 또는 YYYYMM")
	fmt.Println("  --rows-type   행 단위 day|month|quarter|year|none (기본: month)")
	fmt.Println("  --account-id  특정 항목만 조회 (예: x12)")
	fmt.Println("  --item        아이템 필터 (* 와일드카드)")
	fmt.Println("  --section     섹션 ID (전역 플래그)")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whoo report --from 202601 --to 202612")
	fmt.Println("  whoo report expenses,income --from 20260101 --to 20260131 --rows-type day")
	fmt.Println("  whoo report assets,liabilities --from 20260815 --to 20260815 --rows-type none")
	fmt.Println("  whoo report summary --from 202601 --to 202606")
	fmt.Println("  whoo report expenses --account-id x12 --item '커피*'")
}
