// cmd/bs.go
// bs CLI 커맨드 — 자산/부채 잔액(Balance Sheet) 조회

package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"whoo-cli/config"
)

// RunBS는 bs CLI 커맨드 실행
func RunBS(cfg *config.Config, args []string) {
	if wantsHelp(args) {
		showBSHelp()
		return
	}

	RequireAuth(cfg)
	RequireSection(cfg)

	fs := flag.NewFlagSet("bs", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = showBSHelp
	end := fs.String("end", "", "기준일 YYYYMMDD (기본: 오늘)")
	if err := fs.Parse(args); err != nil {
		PrintError("%v", err)
		showBSHelp()
		os.Exit(1)
	}

	if *end == "" {
		*end = time.Now().Format("20060102")
	}

	client := NewClient(cfg)
	data, err := client.GetBSRaw(cfg.SectionID, *end)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

func showBSHelp() {
	fmt.Println("사용법: whoo bs [플래그]")
	fmt.Println()
	fmt.Println("현재 섹션의 자산/부채 잔액을 JSON으로 출력합니다.")
	fmt.Println("(내부적으로 통합 보고서 API report/assets,liabilities를 사용합니다)")
	fmt.Println("항목 메타만 보려면 whoo accounts, 기간 증감은 whoo inout 을 사용하세요.")
	fmt.Println("기간별 보고서가 필요하면 whoo report 를 사용하세요.")
	fmt.Println()
	fmt.Println("플래그:")
	fmt.Println("  --end          기준일 YYYYMMDD (기본: 오늘)")
	fmt.Println("  -h, --help     도움말")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whoo bs")
	fmt.Println("  whoo bs --end 20260813")
}
