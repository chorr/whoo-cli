// cmd/entries.go
// entries CLI 커맨드 — 거래내역 관리 (서브커맨드 + 플래그 구조)

package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"whooing-cli/config"
)

// RunEntries는 entries CLI 커맨드 실행
func RunEntries(cfg *config.Config, args []string) {
	RequireAuth(cfg)
	RequireSection(cfg)

	// 서브커맨드 없으면 기본 목록 조회
	if len(args) == 0 {
		runEntriesList(cfg, args)
		return
	}

	switch args[0] {
	case "latest":
		runEntriesLatest(cfg, args[1:])
	case "latest_items":
		runEntriesLatestItems(cfg)
	case "help", "--help", "-h":
		showEntriesHelp()
	default:
		// 플래그(-로 시작)면 목록 조회, 아니면 entry_id로 간주
		if strings.HasPrefix(args[0], "-") {
			runEntriesList(cfg, args)
		} else {
			runEntriesGet(cfg, args[0])
		}
	}
}

// runEntriesList는 거래내역 목록 조회 (flag 파싱)
func runEntriesList(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("entries", flag.ContinueOnError)
	from := fs.String("from", "", "시작 날짜 (YYYYMMDD)")
	to := fs.String("to", "", "종료 날짜 (YYYYMMDD)")
	limit := fs.Int("limit", 0, "조회 수 제한")

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	// 기본값: 이번 달
	now := time.Now()
	if *from == "" {
		*from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("20060102")
	}
	if *to == "" {
		*to = now.Format("20060102")
	}

	client := NewClient(cfg)
	data, err := client.GetEntriesSearch(cfg.SectionID, *from, *to, *limit)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runEntriesGet은 특정 거래 조회
func runEntriesGet(cfg *config.Config, entryID string) {
	client := NewClient(cfg)
	data, err := client.GetEntryDetail(cfg.SectionID, entryID)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runEntriesLatest는 최근 거래내역 조회
func runEntriesLatest(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("entries latest", flag.ContinueOnError)
	limit := fs.Int("limit", 0, "조회 수 제한")

	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.GetLatestEntries(cfg.SectionID, *limit)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runEntriesLatestItems는 최근 아이템 목록 조회
func runEntriesLatestItems(cfg *config.Config) {
	client := NewClient(cfg)
	data, err := client.GetLatestItems(cfg.SectionID)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// showEntriesHelp는 entries 서브커맨드 도움말 출력
func showEntriesHelp() {
	fmt.Println("사용법: whooing entries [command] [flags]")
	fmt.Println()
	fmt.Println("커맨드:")
	fmt.Println("  (없음)         거래내역 조회 (기본: 이번 달)")
	fmt.Println("  <entry_id>     특정 거래 조회")
	fmt.Println("  latest         최근 거래내역")
	fmt.Println("  latest_items   최근 아이템 목록")
	fmt.Println("  help           도움말 표시")
	fmt.Println()
	fmt.Println("플래그:")
	fmt.Println("  --from YYYYMMDD   시작 날짜 (기본: 이번 달 1일)")
	fmt.Println("  --to YYYYMMDD     종료 날짜 (기본: 오늘)")
	fmt.Println("  --limit N         조회 수 제한")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whooing entries                              이번 달 거래내역")
	fmt.Println("  whooing entries --from 20260301 --to 20260310  기간 지정")
	fmt.Println("  whooing entries --limit 5                    최근 5건")
	fmt.Println("  whooing entries 1352827                      특정 거래 조회")
	fmt.Println("  whooing entries latest                       최근 거래내역")
	fmt.Println("  whooing entries latest --limit 10            최근 10건")
	fmt.Println("  whooing entries latest_items                 최근 아이템 목록")
}
