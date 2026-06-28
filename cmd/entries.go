// cmd/entries.go
// entries CLI 커맨드 — 거래내역 관리

package cmd

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"whoo-cli/api"
	"whoo-cli/config"
)

// RunEntries는 entries CLI 커맨드 실행
func RunEntries(cfg *config.Config, args []string) {
	RequireAuth(cfg)
	RequireSection(cfg)

	if len(args) == 0 {
		runEntriesList(cfg, args)
		return
	}

	switch args[0] {
	case "add":
		runEntriesAdd(cfg, args[1:])
	case "batch":
		runEntriesBatch(cfg, args[1:])
	case "update":
		runEntriesUpdate(cfg, args[1:])
	case "delete", "del", "rm":
		runEntriesDelete(cfg, args[1:])
	case "search":
		runEntriesSearch(cfg, args[1:])
	case "latest":
		runEntriesLatest(cfg, args[1:])
	case "suggest", "latest_items":
		runEntriesLatestItems(cfg)
	case "flow":
		runEntriesFlow(cfg, args[1:])
	case "changes":
		runEntriesChanges(cfg, args[1:])
	case "outside":
		runEntriesOutside(cfg, args[1:])
	case "help", "--help", "-h":
		showEntriesHelp()
	default:
		if strings.HasPrefix(args[0], "-") {
			runEntriesList(cfg, args)
		} else {
			runEntriesGet(cfg, args[0])
		}
	}
}

// runEntriesList는 거래내역 조회 (flag 파싱)
func runEntriesList(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("entries", flag.ContinueOnError)
	from := fs.String("from", "", "시작 날짜 (YYYYMMDD)")
	to := fs.String("to", "", "종료 날짜 (YYYYMMDD)")
	limit := fs.Int("limit", 0, "조회 수 제한")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

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

// runEntriesAdd는 단건 거래 추가 (반복/할부 명령어 지원)
func runEntriesAdd(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("entries add", flag.ExitOnError)
	lAccount := fs.String("l-account", "", "왼쪽 계정 (필수)")
	lID := fs.String("l-id", "", "왼쪽 항목 ID (필수)")
	rAccount := fs.String("r-account", "", "오른쪽 계정 (필수)")
	rID := fs.String("r-id", "", "오른쪽 항목 ID (필수)")
	money := fs.Int64("money", 0, "금액 (필수)")
	item := fs.String("item", "", "아이템 (명령어 포함 가능)")
	memo := fs.String("memo", "", "메모")
	date := fs.String("date", "", "날짜 YYYYMMDD (기본: 오늘)")
	repeat := fs.Int("repeat", 0, "반복 횟수 (item에 **n 추가)")
	split := fs.Int("split", 0, "할부 개월 (item에 //n 추가)")
	fee := fs.Float64("fee", 0, "할부 수수료율 %%")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *lAccount == "" || *lID == "" || *rAccount == "" || *rID == "" || *money == 0 {
		PrintError("--l-account, --l-id, --r-account, --r-id, --money 는 필수입니다")
		os.Exit(1)
	}
	if *date == "" {
		*date = time.Now().Format("20060102")
	}

	// item 문자열 구성
	itemStr := *item
	if *split > 0 || *repeat > 0 {
		if itemStr == "" {
			PrintError("반복/할부 사용 시 --item 이 필요합니다")
			os.Exit(1)
		}
		cmd := &ItemCommand{
			Base:   itemStr,
			Split:  *split,
			Repeat: *repeat,
			Fee:    *fee,
		}
		if _, err := ParseItemCommand(cmd.String()); err != nil {
			PrintError("아이템 명령어 오류: %v", err)
			os.Exit(1)
		}
		itemStr = cmd.String()
	}

	client := NewClient(cfg)
	data, err := client.CreateEntry(
		cfg.SectionID, *date,
		*lAccount, *lID, *rAccount, *rID,
		itemStr, *memo, float64(*money),
	)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	out, _ := marshalPretty(data)
	fmt.Println(string(out))
}

// runEntriesBatch는 JSON 파일에서 일괄 입력
func runEntriesBatch(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("entries batch", flag.ExitOnError)
	file := fs.String("file", "", "거래 목록 JSON 파일 (필수)")
	sectionID := fs.String("section", cfg.SectionID, "섹션 ID")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *file == "" {
		PrintError("--file 은 필수입니다")
		os.Exit(1)
	}

	content, err := os.ReadFile(*file)
	if err != nil {
		PrintError("파일 읽기 실패: %v", err)
		os.Exit(1)
	}

	var rows []api.EntryInput
	if err := parseJSONResponse(content, &rows); err != nil {
		PrintError("JSON 파싱 실패: %v", err)
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.CreateEntriesBatch(*sectionID, rows)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runEntriesUpdate는 단건 거래 수정
func runEntriesUpdate(cfg *config.Config, args []string) {
	if len(args) == 0 {
		PrintError("entry_id가 필요합니다")
		os.Exit(1)
	}
	entryID := args[0]

	fs := flag.NewFlagSet("entries update", flag.ExitOnError)
	lAccount := fs.String("l-account", "", "왼쪽 계정")
	lID := fs.String("l-id", "", "왼쪽 항목 ID")
	rAccount := fs.String("r-account", "", "오른쪽 계정")
	rID := fs.String("r-id", "", "오른쪽 항목 ID")
	money := fs.String("money", "", "금액")
	item := fs.String("item", "", "아이템")
	memo := fs.String("memo", "", "메모")
	date := fs.String("date", "", "날짜 YYYYMMDD")
	if err := fs.Parse(args[1:]); err != nil {
		os.Exit(1)
	}

	fields := make(map[string]string)
	if *lAccount != "" {
		fields["l_account"] = *lAccount
		fields["l_account_id"] = *lID
	}
	if *rAccount != "" {
		fields["r_account"] = *rAccount
		fields["r_account_id"] = *rID
	}
	if *money != "" {
		fields["money"] = *money
	}
	if *item != "" {
		fields["item"] = *item
	}
	if *memo != "" {
		fields["memo"] = *memo
	}
	if *date != "" {
		fields["entry_date"] = *date
	}

	entryIDInt, err := strconv.Atoi(entryID)
	if err != nil {
		PrintError("유효하지 않은 entry_id: %s", entryID)
		os.Exit(1)
	}

	client := NewClient(cfg)
	entry, err := client.UpdateEntry(cfg.SectionID, entryIDInt, fields)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	out, _ := marshalPretty(entry)
	fmt.Println(string(out))
}

// runEntriesDelete는 단건 또는 복수 삭제
func runEntriesDelete(cfg *config.Config, args []string) {
	if len(args) == 0 {
		PrintError("entry_id가 필요합니다 (콤마로 복수 지정 가능)")
		os.Exit(1)
	}
	idStrs := splitCommaOrArgs(args)
	var ids []int64
	for _, s := range idStrs {
		id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil {
			PrintError("유효하지 않은 entry_id: %s", s)
			os.Exit(1)
		}
		ids = append(ids, id)
	}

	client := NewClient(cfg)
	data, err := client.DeleteEntries(cfg.SectionID, ids)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runEntriesSearch는 고급 필터 검색
func runEntriesSearch(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("entries search", flag.ExitOnError)
	from := fs.String("from", "", "시작 날짜 YYYYMMDD")
	to := fs.String("to", "", "종료 날짜 YYYYMMDD")
	limit := fs.Int("limit", 20, "최대 조회 수 (기본 20)")
	account := fs.String("account", "", "계정 필터")
	accountID := fs.String("account-id", "", "항목 ID 필터")
	lAccount := fs.String("l-account", "", "왼쪽 계정 필터")
	lID := fs.String("l-id", "", "왼쪽 항목 ID 필터")
	rAccount := fs.String("r-account", "", "오른쪽 계정 필터")
	rID := fs.String("r-id", "", "오른쪽 항목 ID 필터")
	item := fs.String("item", "", "아이템 필터 (* 와일드카드)")
	memo := fs.String("memo", "", "메모 필터 (공백=AND, ! prefix=제외)")
	moneyFrom := fs.Int64("money-from", 0, "최소 금액")
	moneyTo := fs.Int64("money-to", 0, "최대 금액")
	sortCol := fs.String("sort", "", "정렬 기준 (entry_date|item|money)")
	sortOrder := fs.String("order", "desc", "정렬 방향 (desc|asc)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	now := time.Now()
	if *from == "" {
		*from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("20060102")
	}
	if *to == "" {
		*to = now.Format("20060102")
	}
	fromInt, _ := strconv.Atoi(*from)
	toInt, _ := strconv.Atoi(*to)

	client := NewClient(cfg)
	data, err := client.SearchEntries(api.EntrySearch{
		SectionID:  cfg.SectionID,
		StartDate:  fromInt,
		EndDate:    toInt,
		Limit:      *limit,
		Account:    *account,
		AccountID:  *accountID,
		LAccount:   *lAccount,
		LAccountID: *lID,
		RAccount:   *rAccount,
		RAccountID: *rID,
		Item:       *item,
		Memo:       *memo,
		MoneyFrom:  *moneyFrom,
		MoneyTo:    *moneyTo,
		SortColumn: *sortCol,
		SortOrder:  *sortOrder,
	})
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runEntriesLatest는 최근 거래 조회
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

// runEntriesLatestItems는 최근 아이템 목록 (Suggest)
func runEntriesLatestItems(cfg *config.Config) {
	client := NewClient(cfg)
	data, err := client.GetLatestItems(cfg.SectionID)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runEntriesFlow는 계정/항목 흐름 분석
func runEntriesFlow(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("entries flow", flag.ExitOnError)
	from := fs.String("from", "", "시작 날짜 YYYYMMDD (필수)")
	to := fs.String("to", "", "종료 날짜 YYYYMMDD (필수)")
	account := fs.String("account", "", "계정 (flow_of_account 용)")
	accountID := fs.String("account-id", "", "항목 ID (flow_of_account_id 용)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *from == "" || *to == "" {
		PrintError("--from 과 --to 는 필수입니다")
		os.Exit(1)
	}
	fromInt, _ := strconv.Atoi(*from)
	toInt, _ := strconv.Atoi(*to)

	q := api.FlowQuery{
		SectionID: cfg.SectionID,
		StartDate: fromInt,
		EndDate:   toInt,
		Account:   *account,
		AccountID: *accountID,
	}

	client := NewClient(cfg)
	var data []byte
	var err error
	if *accountID != "" {
		data, err = client.FlowOfAccountID(q)
	} else {
		data, err = client.FlowOfAccount(q)
	}
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runEntriesChanges는 일일 변동 분석
func runEntriesChanges(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("entries changes", flag.ExitOnError)
	from := fs.String("from", "", "시작 날짜 YYYYMMDD (필수)")
	to := fs.String("to", "", "종료 날짜 YYYYMMDD (필수)")
	accountID := fs.String("account-id", "", "항목 ID (changes_of_account_id)")
	client_ := fs.String("client", "", "거래처 (changes_of_client)")
	item := fs.String("item", "", "아이템 (changes_of_item)")
	rowsType := fs.String("rows-type", "day", "집계 단위 day|month|year")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *from == "" || *to == "" {
		PrintError("--from 과 --to 는 필수입니다")
		os.Exit(1)
	}
	fromInt, _ := strconv.Atoi(*from)
	toInt, _ := strconv.Atoi(*to)

	q := api.FlowQuery{
		SectionID: cfg.SectionID,
		StartDate: fromInt,
		EndDate:   toInt,
		AccountID: *accountID,
		Item:      *item,
		Memo:      *client_, // changes_of_client는 client를 item처럼 전달
		RowsType:  *rowsType,
	}

	client := NewClient(cfg)
	var data []byte
	var err error
	switch {
	case *client_ != "":
		q.Item = *client_
		q.Memo = ""
		data, err = client.ChangesOfClient(q)
	case *item != "":
		data, err = client.ChangesOfItem(q)
	default:
		data, err = client.ChangesOfAccountID(q)
	}
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runEntriesOutside는 외부 데이터 파싱 (SMS 등)
func runEntriesOutside(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("entries outside", flag.ExitOnError)
	file := fs.String("file", "", "외부 데이터 텍스트 파일 (필수)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *file == "" {
		PrintError("--file 은 필수입니다")
		os.Exit(1)
	}

	content, err := os.ReadFile(*file)
	if err != nil {
		PrintError("파일 읽기 실패: %v", err)
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.ParseOutside(cfg.SectionID, string(content))
	if err != nil {
		// 400이면 outside_report 자동 호출 권장
		PrintError("%v", err)
		fmt.Fprintln(os.Stderr, "  지원하지 않는 형식입니다. outside_report API로 보고하면 향후 지원될 수 있습니다.")
		os.Exit(1)
	}
	printJSON(data)
}

// showEntriesHelp는 entries 서브커맨드 도움말 출력
func showEntriesHelp() {
	fmt.Println("사용법: whoo entries [command] [flags]")
	fmt.Println()
	fmt.Println("커맨드:")
	fmt.Println("  (없음)       거래내역 조회 (기본: 이번 달)")
	fmt.Println("  add          단건 거래 추가")
	fmt.Println("  batch        JSON 파일로 일괄 입력")
	fmt.Println("  update       단건 거래 수정")
	fmt.Println("  delete       단건/복수 삭제")
	fmt.Println("  search       고급 필터 검색")
	fmt.Println("  latest       최근 거래내역")
	fmt.Println("  suggest      최근 아이템 목록 (Suggest)")
	fmt.Println("  flow         계정/항목 흐름 분석")
	fmt.Println("  changes      일일 변동 분석")
	fmt.Println("  outside      외부 데이터(SMS 등) 파싱 입력")
	fmt.Println("  <entry_id>   특정 거래 조회")
	fmt.Println("  help         도움말")
	fmt.Println()
	fmt.Println("entries add 플래그:")
	fmt.Println("  --l-account   왼쪽 계정 (필수)")
	fmt.Println("  --l-id        왼쪽 항목 ID (필수)")
	fmt.Println("  --r-account   오른쪽 계정 (필수)")
	fmt.Println("  --r-id        오른쪽 항목 ID (필수)")
	fmt.Println("  --money       금액 (필수)")
	fmt.Println("  --item        아이템 명")
	fmt.Println("  --memo        메모")
	fmt.Println("  --date        날짜 YYYYMMDD (기본: 오늘)")
	fmt.Println("  --split N     할부 N개월 (item에 //N 추가)")
	fmt.Println("  --fee F       할부 수수료율 %")
	fmt.Println("  --repeat N    반복 N회 (item에 **N 추가)")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whoo entries add --l-account expenses --l-id x12 --r-account assets --r-id x5 --money 8000 --item 커피")
	fmt.Println("  whoo entries add --l-account liabilities --l-id x10 --r-account assets --r-id x5 --money 1200000 --item 노트북 --split 3 --fee 2.5")
	fmt.Println("  whoo entries delete 1352827,1352828")
	fmt.Println("  whoo entries search --item '커피*' --money-from 3000 --money-to 10000")
	fmt.Println("  whoo entries flow --from 20260101 --to 20260131 --account expenses")
	fmt.Println("  whoo entries changes --from 20260101 --to 20260131 --account-id x12")
}
