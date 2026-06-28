// cmd/frequent.go
// frequent CLI 커맨드 — 자주입력 관리

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

// RunFrequent는 frequent CLI 커맨드 실행
func RunFrequent(cfg *config.Config, args []string) {
	RequireAuth(cfg)
	RequireSection(cfg)

	if len(args) == 0 {
		runFrequentList(cfg, nil)
		return
	}

	switch args[0] {
	case "list", "ls":
		runFrequentList(cfg, args[1:])
	case "add":
		runFrequentAdd(cfg, args[1:])
	case "edit":
		runFrequentEdit(cfg, args[1:])
	case "delete", "del", "rm":
		runFrequentDelete(cfg, args[1:])
	case "sort":
		runFrequentSort(cfg, args[1:])
	case "use":
		runFrequentUse(cfg, args[1:])
	case "help", "--help", "-h":
		showFrequentHelp()
	default:
		showFrequentHelp()
	}
}

// runFrequentList는 자주입력 목록 조회
func runFrequentList(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("frequent list", flag.ExitOnError)
	slot := fs.String("slot", "", "슬롯 (slot1/slot2/slot3, 기본: 전체)")
	if err := fs.Parse(args); err != nil {
		return
	}

	client := NewClient(cfg)
	var data []byte
	var err error

	if *slot == "" {
		data, err = client.GetFrequentItems(cfg.SectionID)
	} else {
		data, err = client.GetFrequentItemsSlot(cfg.SectionID, *slot)
	}
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runFrequentAdd는 자주입력 항목 생성
func runFrequentAdd(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("frequent add", flag.ExitOnError)
	slot := fs.String("slot", "slot1", "슬롯 (slot1/slot2/slot3)")
	item := fs.String("item", "", "아이템명 (필수)")
	money := fs.Int64("money", 0, "금액 (0=미지정)")
	lAccount := fs.String("l-account", "", "왼쪽 계정 타입 (필수)")
	lID := fs.String("l-id", "", "왼쪽 계정 ID (필수)")
	rAccount := fs.String("r-account", "", "오른쪽 계정 타입 (필수)")
	rID := fs.String("r-id", "", "오른쪽 계정 ID (필수)")
	if err := fs.Parse(args); err != nil {
		return
	}

	if *item == "" || *lAccount == "" || *lID == "" || *rAccount == "" || *rID == "" {
		PrintError("--item, --l-account, --l-id, --r-account, --r-id 필수")
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.CreateFrequentItem(cfg.SectionID, *slot, api.FrequentItemInput{
		Item:       *item,
		Money:      *money,
		LAccount:   *lAccount,
		LAccountID: *lID,
		RAccount:   *rAccount,
		RAccountID: *rID,
	})
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runFrequentEdit는 자주입력 항목 수정
func runFrequentEdit(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("frequent edit", flag.ExitOnError)
	slot := fs.String("slot", "slot1", "슬롯 (slot1/slot2/slot3)")
	item := fs.String("item", "", "아이템명")
	money := fs.Int64("money", 0, "금액")
	lAccount := fs.String("l-account", "", "왼쪽 계정 타입")
	lID := fs.String("l-id", "", "왼쪽 계정 ID")
	rAccount := fs.String("r-account", "", "오른쪽 계정 타입")
	rID := fs.String("r-id", "", "오른쪽 계정 ID")
	if err := fs.Parse(args); err != nil {
		return
	}

	if fs.NArg() < 1 {
		PrintError("item_id 필요: frequent edit [--slot slot1] <item_id> [옵션]")
		os.Exit(1)
	}
	itemID := fs.Arg(0)

	client := NewClient(cfg)
	data, err := client.UpdateFrequentItem(cfg.SectionID, *slot, itemID, api.FrequentItemInput{
		Item:       *item,
		Money:      *money,
		LAccount:   *lAccount,
		LAccountID: *lID,
		RAccount:   *rAccount,
		RAccountID: *rID,
	})
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runFrequentDelete는 자주입력 항목 삭제
func runFrequentDelete(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("frequent delete", flag.ExitOnError)
	slot := fs.String("slot", "slot1", "슬롯 (slot1/slot2/slot3)")
	if err := fs.Parse(args); err != nil {
		return
	}

	if fs.NArg() < 1 {
		PrintError("item_id 필요: frequent delete [--slot slot1] <item_id>")
		os.Exit(1)
	}
	itemID := fs.Arg(0)

	client := NewClient(cfg)
	data, err := client.DeleteFrequentItem(cfg.SectionID, *slot, itemID)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runFrequentSort는 자주입력 항목 순서 변경
func runFrequentSort(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("frequent sort", flag.ExitOnError)
	slot := fs.String("slot", "slot1", "슬롯 (slot1/slot2/slot3)")
	if err := fs.Parse(args); err != nil {
		return
	}

	if fs.NArg() < 1 {
		PrintError("ID 목록 필요: frequent sort [--slot slot1] <id1,id2,id3>")
		os.Exit(1)
	}
	ids := strings.Split(fs.Arg(0), ",")

	client := NewClient(cfg)
	data, err := client.SortFrequentItems(cfg.SectionID, *slot, ids)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runFrequentUse는 자주입력 항목을 실제 거래로 변환
// 서버 엔드포인트 없음 — frequent 항목 조회 후 entries 생성하는 클라이언트 래퍼
func runFrequentUse(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("frequent use", flag.ExitOnError)
	slot := fs.String("slot", "slot1", "슬롯 (slot1/slot2/slot3)")
	money := fs.Int64("money", 0, "금액 덮어쓰기 (0=항목 기본값 사용)")
	memo := fs.String("memo", "", "메모")
	date := fs.String("date", "", "날짜 YYYYMMDD (기본: 오늘)")
	if err := fs.Parse(args); err != nil {
		return
	}

	if fs.NArg() < 1 {
		PrintError("item_id 필요: frequent use [--slot slot1] <item_id> [옵션]")
		os.Exit(1)
	}
	itemID := fs.Arg(0)

	client := NewClient(cfg)

	// 슬롯 전체 조회 후 해당 item_id 검색
	raw, err := client.GetFrequentItemsSlot(cfg.SectionID, *slot)
	if err != nil {
		PrintError("자주입력 조회 실패: %v", err)
		os.Exit(1)
	}

	fi := findFrequentItemByID(raw, itemID)
	if fi == nil {
		PrintError("item_id '%s'를 찾을 수 없습니다", itemID)
		os.Exit(1)
	}

	// 금액 결정
	useMoney := *money
	if useMoney == 0 {
		useMoney = fi.Money
	}

	// 날짜 결정
	useDate := *date
	if useDate == "" {
		useDate = time.Now().Format("20060102")
	}

	entry, err := client.CreateEntry(
		cfg.SectionID, useDate,
		fi.LAccount, fi.LAccountID, fi.RAccount, fi.RAccountID,
		fi.Item, *memo, float64(useMoney),
	)
	if err != nil {
		PrintError("거래 생성 실패: %v", err)
		os.Exit(1)
	}
	out, _ := marshalPretty(entry)
	fmt.Println(string(out))
}

// frequentItemInfo는 frequent use에서 필요한 항목 정보
type frequentItemInfo struct {
	Item       string
	Money      int64
	LAccount   string
	LAccountID string
	RAccount   string
	RAccountID string
}

// findFrequentItemByID는 raw API 응답에서 특정 ID의 자주입력 항목을 찾아 반환
func findFrequentItemByID(raw []byte, itemID string) *frequentItemInfo {
	var wrapper map[string]interface{}
	if err := parseJSONResponse(raw, &wrapper); err != nil {
		return nil
	}

	// results 레벨 언래핑
	var root map[string]interface{}
	if r, ok := wrapper["results"]; ok {
		if rm, ok := r.(map[string]interface{}); ok {
			root = rm
		}
	}
	if root == nil {
		root = wrapper
	}

	for _, v := range root {
		arr, ok := v.([]interface{})
		if !ok {
			continue
		}
		for _, elem := range arr {
			m, ok := elem.(map[string]interface{})
			if !ok {
				continue
			}
			id, _ := m["frequent_item_id"].(string)
			if id != itemID {
				continue
			}
			fi := &frequentItemInfo{}
			fi.Item, _ = m["item"].(string)
			fi.LAccount, _ = m["l_account"].(string)
			fi.LAccountID, _ = m["l_account_id"].(string)
			fi.RAccount, _ = m["r_account"].(string)
			fi.RAccountID, _ = m["r_account_id"].(string)
			if mv, ok := m["money"].(float64); ok {
				fi.Money = int64(mv)
			}
			return fi
		}
	}
	return nil
}

func showFrequentHelp() {
	fmt.Print(`자주입력 관리

사용법:
  whoo frequent [서브커맨드] [옵션]

서브커맨드:
  list   [--slot slot1]                          자주입력 목록 조회
  add    --slot slot1 --item <명> --l-account <타입> --l-id <id>
         --r-account <타입> --r-id <id> [--money <금액>]
                                                  자주입력 항목 추가
  edit   [--slot slot1] <item_id> [옵션]          자주입력 항목 수정
  delete [--slot slot1] <item_id>                자주입력 항목 삭제
  sort   [--slot slot1] <id1,id2,id3>            순서 변경
  use    [--slot slot1] <item_id> [--money N]
         [--memo <메모>] [--date YYYYMMDD]        거래로 변환 (entries 생성)

예시:
  whoo frequent list
  whoo frequent list --slot slot2
  whoo frequent add --item "커피" --l-account expenses --l-id x1 \
    --r-account assets --r-id x2 --money 5000
  whoo frequent use f1 --memo "점심"
`)
}
