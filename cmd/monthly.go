// cmd/monthly.go
// monthly CLI 커맨드 — 월별입력 관리

package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"whoo-cli/api"
	"whoo-cli/config"
)

// RunMonthly는 monthly CLI 커맨드 실행
func RunMonthly(cfg *config.Config, args []string) {
	RequireAuth(cfg)
	RequireSection(cfg)

	if len(args) == 0 {
		runMonthlyList(cfg, nil)
		return
	}

	switch args[0] {
	case "list", "ls":
		runMonthlyList(cfg, args[1:])
	case "add":
		runMonthlyAdd(cfg, args[1:])
	case "edit":
		runMonthlyEdit(cfg, args[1:])
	case "delete", "del", "rm":
		runMonthlyDelete(cfg, args[1:])
	case "sort":
		runMonthlySort(cfg, args[1:])
	case "pay", "use":
		runMonthlyPay(cfg, args[1:])
	case "help", "--help", "-h":
		showMonthlyHelp()
	default:
		showMonthlyHelp()
	}
}

// runMonthlyList는 월별입력 목록 조회
func runMonthlyList(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("monthly list", flag.ExitOnError)
	slot := fs.String("slot", "", "슬롯 (slot1/slot2/slot3, 기본: 전체)")
	if err := fs.Parse(args); err != nil {
		return
	}

	client := NewClient(cfg)
	var data []byte
	var err error

	if *slot == "" {
		data, err = client.GetMonthlyItems(cfg.SectionID)
	} else {
		data, err = client.GetMonthlyItemsSlot(cfg.SectionID, *slot)
	}
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runMonthlyAdd는 월별입력 항목 생성
func runMonthlyAdd(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("monthly add", flag.ExitOnError)
	slot := fs.String("slot", "slot1", "슬롯 (slot1/slot2/slot3)")
	item := fs.String("item", "", "아이템명 (필수)")
	money := fs.Int64("money", 0, "금액 (0=미지정)")
	lAccount := fs.String("l-account", "", "왼쪽 계정 타입 (필수)")
	lID := fs.String("l-id", "", "왼쪽 계정 ID (필수)")
	rAccount := fs.String("r-account", "", "오른쪽 계정 타입 (필수)")
	rID := fs.String("r-id", "", "오른쪽 계정 ID (필수)")
	payDate := fs.Int("pay-date", 0, "결제일 1~31 (필수)")
	skipHoliday := fs.String("skip-holiday", "none", "공휴일 처리 before|after|none")
	if err := fs.Parse(args); err != nil {
		return
	}

	if *item == "" || *lAccount == "" || *lID == "" || *rAccount == "" || *rID == "" {
		PrintError("--item, --l-account, --l-id, --r-account, --r-id 필수")
		os.Exit(1)
	}
	if *payDate < 1 || *payDate > 31 {
		PrintError("--pay-date는 1~31 사이 값 필요")
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.CreateMonthlyItem(cfg.SectionID, *slot, api.MonthlyItemInput{
		FrequentItemInput: api.FrequentItemInput{
			Item:       *item,
			Money:      *money,
			LAccount:   *lAccount,
			LAccountID: *lID,
			RAccount:   *rAccount,
			RAccountID: *rID,
		},
		PayDate:     *payDate,
		SkipHoliday: *skipHoliday,
	})
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runMonthlyEdit는 월별입력 항목 수정
func runMonthlyEdit(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("monthly edit", flag.ExitOnError)
	slot := fs.String("slot", "slot1", "슬롯 (slot1/slot2/slot3)")
	item := fs.String("item", "", "아이템명")
	money := fs.Int64("money", 0, "금액")
	lAccount := fs.String("l-account", "", "왼쪽 계정 타입")
	lID := fs.String("l-id", "", "왼쪽 계정 ID")
	rAccount := fs.String("r-account", "", "오른쪽 계정 타입")
	rID := fs.String("r-id", "", "오른쪽 계정 ID")
	payDate := fs.Int("pay-date", 0, "결제일 1~31")
	skipHoliday := fs.String("skip-holiday", "", "공휴일 처리 before|after|none")
	if err := fs.Parse(args); err != nil {
		return
	}

	if fs.NArg() < 1 {
		PrintError("item_id 필요: monthly edit [--slot slot1] <item_id> [옵션]")
		os.Exit(1)
	}
	itemID := fs.Arg(0)

	client := NewClient(cfg)
	data, err := client.UpdateMonthlyItem(cfg.SectionID, *slot, itemID, api.MonthlyItemInput{
		FrequentItemInput: api.FrequentItemInput{
			Item:       *item,
			Money:      *money,
			LAccount:   *lAccount,
			LAccountID: *lID,
			RAccount:   *rAccount,
			RAccountID: *rID,
		},
		PayDate:     *payDate,
		SkipHoliday: *skipHoliday,
	})
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runMonthlyDelete는 월별입력 항목 삭제
func runMonthlyDelete(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("monthly delete", flag.ExitOnError)
	slot := fs.String("slot", "slot1", "슬롯 (slot1/slot2/slot3)")
	if err := fs.Parse(args); err != nil {
		return
	}

	if fs.NArg() < 1 {
		PrintError("item_id 필요: monthly delete [--slot slot1] <item_id>")
		os.Exit(1)
	}
	itemID := fs.Arg(0)

	client := NewClient(cfg)
	data, err := client.DeleteMonthlyItem(cfg.SectionID, *slot, itemID)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runMonthlySort는 월별입력 항목 순서 변경
func runMonthlySort(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("monthly sort", flag.ExitOnError)
	slot := fs.String("slot", "slot1", "슬롯 (slot1/slot2/slot3)")
	if err := fs.Parse(args); err != nil {
		return
	}

	if fs.NArg() < 1 {
		PrintError("ID 목록 필요: monthly sort [--slot slot1] <id1,id2,id3>")
		os.Exit(1)
	}
	ids := strings.Split(fs.Arg(0), ",")

	client := NewClient(cfg)
	data, err := client.SortMonthlyItems(cfg.SectionID, *slot, ids)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runMonthlyPay는 월별입력을 실제 거래로 확정(결제 입력)
// 공식 API에 pay 엔드포인트는 없고 CreateEntry로 처리한다.
func runMonthlyPay(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("monthly pay", flag.ExitOnError)
	slot := fs.String("slot", "slot1", "슬롯 (slot1/slot2/slot3)")
	money := fs.Int64("money", 0, "금액 덮어쓰기 (0=항목 기본값)")
	memo := fs.String("memo", "", "메모")
	date := fs.String("date", "", "거래일 YYYYMMDD (기본: due_date, 없으면 오늘)")
	if err := fs.Parse(args); err != nil {
		return
	}
	if fs.NArg() < 1 {
		PrintError("item_id 필요: monthly pay [--slot slot1] <item_id> [옵션]")
		os.Exit(1)
	}
	itemID := fs.Arg(0)

	client := NewClient(cfg)
	raw, err := client.GetMonthlyItemsSlot(cfg.SectionID, *slot)
	if err != nil {
		PrintError("월별입력 조회 실패: %v", err)
		os.Exit(1)
	}
	mi := findMonthlyItemByID(raw, itemID)
	if mi == nil {
		PrintError("item_id '%s'를 찾을 수 없습니다", itemID)
		os.Exit(1)
	}

	useMoney := *money
	if useMoney == 0 {
		useMoney = mi.Money
	}
	useDate := *date
	if useDate == "" {
		useDate = mi.DueDate
	}
	if len(useDate) != 8 {
		useDate = time.Now().Format("20060102")
	}

	entry, err := client.CreateEntry(
		cfg.SectionID, useDate,
		mi.LAccount, mi.LAccountID, mi.RAccount, mi.RAccountID,
		mi.Item, *memo, float64(useMoney),
	)
	if err != nil {
		PrintError("거래 생성 실패: %v", err)
		os.Exit(1)
	}
	out, _ := marshalPretty(entry)
	fmt.Println(string(out))
}

type monthlyItemInfo struct {
	Item       string
	Money      int64
	LAccount   string
	LAccountID string
	RAccount   string
	RAccountID string
	DueDate    string
}

func findMonthlyItemByID(raw []byte, itemID string) *monthlyItemInfo {
	for _, m := range extractWhooingItemMaps(raw) {
		if whooingItemID(m) != itemID {
			continue
		}
		return &monthlyItemInfo{
			Item:       mapString(m, "item"),
			Money:      mapInt64(m, "money"),
			LAccount:   mapString(m, "l_account"),
			LAccountID: mapString(m, "l_account_id"),
			RAccount:   mapString(m, "r_account"),
			RAccountID: mapString(m, "r_account_id"),
			DueDate:    mapString(m, "due_date"),
		}
	}
	return nil
}

func showMonthlyHelp() {
	fmt.Print(`월별입력 관리

사용법:
  whoo monthly [서브커맨드] [옵션]

서브커맨드:
  list   [--slot slot1]                          월별입력 목록 조회
  add    --slot slot1 --item <명> --l-account <타입> --l-id <id>
         --r-account <타입> --r-id <id> --pay-date <1~31>
         [--money <금액>] [--skip-holiday before|after|none]
                                                  월별입력 항목 추가
  edit   [--slot slot1] <item_id> [옵션]          월별입력 항목 수정
  delete [--slot slot1] <item_id>                월별입력 항목 삭제
  sort   [--slot slot1] <id1,id2,id3>            순서 변경
  pay    [--slot slot1] <item_id> [--money N]
         [--memo <메모>] [--date YYYYMMDD]        결제 입력 (entries 생성)
         (단축: use)

예시:
  whoo monthly list
  whoo monthly add --item "통신비" --money 79200 --pay-date 27 \
    --skip-holiday after --l-account expenses --l-id x1 \
    --r-account liabilities --r-id x3
  whoo monthly pay m113769
  whoo monthly edit m1 --money 85000
  whoo monthly delete m1
`)
}
