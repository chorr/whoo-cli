// cmd/entries.go
// entries CLI 커맨드 — 거래내역 관리

package cmd

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"whoo-cli/api"
	"whoo-cli/config"
)

type flatEntry struct {
	SectionID  string             `json:"section_id"`
	EntryID    int                `json:"entry_id"`
	EntryDate  api.FlexibleString `json:"entry_date"`
	LAccount   string             `json:"l_account"`
	LAccountID string             `json:"l_account_id"`
	RAccount   string             `json:"r_account"`
	RAccountID string             `json:"r_account_id"`
	Money      float64            `json:"money"`
	Item       string             `json:"item"`
	Memo       string             `json:"memo"`
}

// RunEntries는 entries CLI 커맨드 실행
func RunEntries(cfg *config.Config, args []string) {
	if wantsHelp(args) {
		showEntriesHelpFor(args)
		return
	}
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
	case "outside_report", "outside-report":
		runEntriesOutsideReport(cfg, args[1:])
	case "agg":
		runEntriesAgg(cfg, args[1:])
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
	flat := fs.Bool("flat", false, "compact 배열로 출력")
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
	if *flat {
		rows, err := parseFlatEntries(data, cfg.SectionID)
		if err != nil {
			PrintError("%v", err)
			os.Exit(1)
		}
		printJSONValue(rows)
		return
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
	printJSONValue(data)
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

// runEntriesUpdate는 단건/복수 거래 수정
// entry_id를 콤마로 이으면 복수 수정 (최대 100건, PUT entries/:entry_ids/:section_id.json)
func runEntriesUpdate(cfg *config.Config, args []string) {
	if len(args) == 0 {
		PrintError("entry_id가 필요합니다 (콤마로 복수 지정 가능)")
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

	client := NewClient(cfg)

	// 복수 ID (콤마 구분) → 일괄 수정
	if strings.Contains(entryID, ",") {
		var ids []int64
		for _, s := range strings.Split(entryID, ",") {
			id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
			if err != nil {
				PrintError("유효하지 않은 entry_id: %s", s)
				os.Exit(1)
			}
			ids = append(ids, id)
		}
		changes := api.EntryChanges{
			LAccount:   *lAccount,
			LAccountID: *lID,
			RAccount:   *rAccount,
			RAccountID: *rID,
			Item:       *item,
			Memo:       *memo,
		}
		if *money != "" {
			mv, err := strconv.ParseInt(*money, 10, 64)
			if err != nil {
				PrintError("유효하지 않은 금액: %s", *money)
				os.Exit(1)
			}
			changes.Money = mv
		}
		if *date != "" {
			dv, err := strconv.Atoi(*date)
			if err != nil {
				PrintError("유효하지 않은 날짜: %s", *date)
				os.Exit(1)
			}
			changes.EntryDate = dv
		}
		data, err := client.UpdateEntriesBatch(cfg.SectionID, ids, changes)
		if err != nil {
			PrintError("%v", err)
			os.Exit(1)
		}
		printJSON(data)
		return
	}

	entryIDInt, err := strconv.Atoi(entryID)
	if err != nil {
		PrintError("유효하지 않은 entry_id: %s", entryID)
		os.Exit(1)
	}

	entry, err := client.UpdateEntry(cfg.SectionID, entryIDInt, fields)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSONValue(entry)
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
	max := fs.String("max", "", "entry_date 커서 (예: 20260203.0034) — 페이지네이션")
	sortCol := fs.String("sort", "", "정렬 기준 (entry_date|item|money|total|l_account_id|r_account_id)")
	sortOrder := fs.String("order", "desc", "정렬 방향 (desc|asc)")
	flat := fs.Bool("flat", false, "rows만 compact 배열로 출력")
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
	search := api.EntrySearch{
		SectionID:  cfg.SectionID,
		StartDate:  fromInt,
		EndDate:    toInt,
		Max:        *max,
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
	}
	if err := resolveEntryAccountFilters(client, &search); err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	data, err := client.SearchEntries(search)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	if *flat {
		rows, err := parseFlatEntries(data, cfg.SectionID)
		if err != nil {
			PrintError("%v", err)
			os.Exit(1)
		}
		printJSONValue(rows)
		return
	}
	printJSON(data)
}

// runEntriesLatest는 최근 거래 조회
func runEntriesLatest(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("entries latest", flag.ContinueOnError)
	limit := fs.Int("limit", 0, "조회 수 제한")
	flat := fs.Bool("flat", false, "compact 배열로 출력")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
	client := NewClient(cfg)
	data, err := client.GetLatestEntries(cfg.SectionID, *limit)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	if *flat {
		rows, err := parseFlatEntries(data, cfg.SectionID)
		if err != nil {
			PrintError("%v", err)
			os.Exit(1)
		}
		printJSONValue(rows)
		return
	}
	printJSON(data)
}

func resolveEntryAccountFilters(client *api.WhooingClient, search *api.EntrySearch) error {
	needsAccounts := (search.AccountID != "" && search.Account == "") ||
		(search.LAccountID != "" && search.LAccount == "") ||
		(search.RAccountID != "" && search.RAccount == "")
	if !needsAccounts {
		return nil
	}

	accounts, err := client.GetAccountsMap(search.SectionID)
	if err != nil {
		return fmt.Errorf("항목 타입 확인 실패: %w", err)
	}
	resolve := func(accountID, currentType string) (string, error) {
		if accountID == "" || currentType != "" {
			return currentType, nil
		}
		accountType, ok := findAccountType(accounts, accountID)
		if !ok {
			return "", fmt.Errorf("account_id %s를 현재 섹션에서 찾을 수 없습니다", accountID)
		}
		return accountType, nil
	}

	if search.Account, err = resolve(search.AccountID, search.Account); err != nil {
		return err
	}
	if search.LAccount, err = resolve(search.LAccountID, search.LAccount); err != nil {
		return err
	}
	if search.RAccount, err = resolve(search.RAccountID, search.RAccount); err != nil {
		return err
	}
	return nil
}

func findAccountType(accounts *api.AccountsMap, accountID string) (string, bool) {
	for _, accountType := range AccountTypes {
		if _, ok := accounts.GetAccountsByType(accountType.Code)[accountID]; ok {
			return accountType.Code, true
		}
	}
	return "", false
}

func parseFlatEntries(data []byte, sectionID string) ([]flatEntry, error) {
	var envelope struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Results json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("거래 응답 파싱 실패: %w", err)
	}
	if envelope.Code != 200 && envelope.Code != 204 {
		return nil, &api.APIError{Code: envelope.Code, Message: envelope.Message}
	}

	var entries []api.Entry
	results := strings.TrimSpace(string(envelope.Results))
	switch {
	case results == "" || results == "null":
		return []flatEntry{}, nil
	case strings.HasPrefix(results, "["):
		if err := json.Unmarshal(envelope.Results, &entries); err != nil {
			return nil, fmt.Errorf("거래 배열 파싱 실패: %w", err)
		}
	default:
		var rows struct {
			Rows []api.Entry `json:"rows"`
		}
		if err := json.Unmarshal(envelope.Results, &rows); err != nil {
			return nil, fmt.Errorf("거래 rows 파싱 실패: %w", err)
		}
		entries = rows.Rows
	}

	flat := make([]flatEntry, 0, len(entries))
	for _, entry := range entries {
		flat = append(flat, flatEntry{
			SectionID:  sectionID,
			EntryID:    entry.EntryID,
			EntryDate:  entry.EntryDate,
			LAccount:   entry.LAccount,
			LAccountID: entry.LAccountID,
			RAccount:   entry.RAccount,
			RAccountID: entry.RAccountID,
			Money:      entry.Money,
			Item:       entry.Item,
			Memo:       entry.Memo,
		})
	}
	return flat, nil
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
	reportSource := fs.String("report-source", "", "인식 실패 시 이 소스명으로 outside_report 자동 보고")
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
		PrintError("%v", err)
		if *reportSource != "" {
			// 인식 실패 소스를 서버에 보고 (향후 지원 요청)
			if _, rerr := client.ReportOutside(*reportSource); rerr != nil {
				fmt.Fprintf(os.Stderr, "  outside_report 보고 실패: %v\n", rerr)
			} else {
				fmt.Fprintf(os.Stderr, "  소스 '%s'를 outside_report로 보고했습니다.\n", *reportSource)
			}
		} else {
			fmt.Fprintln(os.Stderr, "  지원하지 않는 형식입니다. whoo entries outside_report --source <소스명> 으로 보고하면 향후 지원될 수 있습니다.")
		}
		os.Exit(1)
	}
	printJSON(data)
}

// runEntriesOutsideReport는 인식되지 않은 외부 데이터 소스를 보고
// POST /api/entries/outside_report.json
func runEntriesOutsideReport(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("entries outside_report", flag.ExitOnError)
	source := fs.String("source", "", "데이터 소스명 (필수, 예: 은행/카드사 이름)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}
	if *source == "" {
		PrintError("--source 는 필수입니다")
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.ReportOutside(*source)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runEntriesAgg는 계정/항목별 금액 집계 조회
// --account          → GET entries/account_ids_of_account.json (계정의 항목별 금액)
// --account-id --by clients → GET entries/clients_of_account_id.json (항목의 거래처별 금액)
// --account-id --by items   → GET entries/items_of_account_id.json (항목의 아이템별 금액)
func runEntriesAgg(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("entries agg", flag.ExitOnError)
	from := fs.String("from", "", "시작 날짜 YYYYMMDD (필수)")
	to := fs.String("to", "", "종료 날짜 YYYYMMDD (필수)")
	account := fs.String("account", "", "계정 (항목별 집계)")
	accountID := fs.String("account-id", "", "항목 ID (거래처/아이템별 집계)")
	by := fs.String("by", "clients", "집계 기준 clients|items (--account-id 사용 시)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *from == "" || *to == "" {
		PrintError("--from 과 --to 는 필수입니다")
		os.Exit(1)
	}
	if *account == "" && *accountID == "" {
		PrintError("--account 또는 --account-id 가 필요합니다")
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
	switch {
	case *accountID != "" && *by == "items":
		data, err = client.ItemsOfAccountID(q)
	case *accountID != "":
		data, err = client.ClientsOfAccountID(q)
	default:
		data, err = client.AccountIDsOfAccount(q)
	}
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// showEntriesHelpFor는 서브커맨드별 도움말. help 예약어는 여기까지이고 API를 치지 않는다.
func showEntriesHelpFor(args []string) {
	if len(args) > 0 && !isHelpArg(args[0]) {
		switch args[0] {
		case "add":
			showEntriesAddHelp()
			return
		case "search":
			showEntriesSearchHelp()
			return
		case "flow":
			showEntriesFlowHelp()
			return
		}
	}
	showEntriesHelp()
}

func showEntriesAddHelp() {
	fmt.Println("사용법: whoo entries add [플래그]")
	fmt.Println()
	fmt.Println("단건 거래를 입력합니다. --split/--repeat는 후잉 item 명령어로 전송됩니다.")
	fmt.Println()
	fmt.Println("플래그:")
	fmt.Println("  --l-account, --l-id   왼쪽 계정과 항목 ID (필수)")
	fmt.Println("  --r-account, --r-id   오른쪽 계정과 항목 ID (필수)")
	fmt.Println("  --money               총액 (필수)")
	fmt.Println("  --item, --memo        아이템과 메모")
	fmt.Println("  --date                날짜 YYYYMMDD (기본: 오늘)")
	fmt.Println("  --split N             총액을 N개월로 분할 (item에 //N)")
	fmt.Println("  --fee F               할부 수수료율 %")
	fmt.Println("  --repeat N            같은 금액을 N개월 반복 (item에 **N)")
	fmt.Println("  --section             섹션 ID (전역 플래그)")
	fmt.Println()
	fmt.Println("할부 금액 분배:")
	fmt.Println("  후잉 서버가 카드 항목의 '할부입력시 처리방식' 단위(1원 또는 100원)에")
	fmt.Println("  맞춰 이후 회차를 절삭하고 나머지를 첫 회차에 더합니다.")
	fmt.Println("  예: 2,290,000원 / 12개월, 100원 단위 → 첫 회 191,200원 + 이후 190,800원×11")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whoo entries add --l-account expenses --l-id x12 --r-account liabilities --r-id x10 --money 2290000 --item 아이폰 --split 12")
}

func showEntriesSearchHelp() {
	fmt.Println("사용법: whoo entries search [플래그]")
	fmt.Println()
	fmt.Println("기간과 항목 조건으로 거래를 검색합니다. ID만 지정하면 계정 타입을 자동 확인합니다.")
	fmt.Println()
	fmt.Println("플래그:")
	fmt.Println("  --from, --to       날짜 범위 YYYYMMDD (기본: 이번 달)")
	fmt.Println("  --limit            최대 조회 수 (기본 20, 최대 100)")
	fmt.Println("  --account          좌우 공통 계정 타입")
	fmt.Println("  --account-id       좌우 어느 쪽이든 일치하는 항목 ID")
	fmt.Println("  --l-account        왼쪽 계정 타입")
	fmt.Println("  --l-id             왼쪽 항목 ID")
	fmt.Println("  --r-account        오른쪽 계정 타입")
	fmt.Println("  --r-id             오른쪽 항목 ID")
	fmt.Println("  --item, --memo     아이템/메모 필터")
	fmt.Println("  --money-from       최소 금액")
	fmt.Println("  --money-to         최대 금액")
	fmt.Println("  --max              entry_date 페이지 커서")
	fmt.Println("  --sort             entry_date|item|money|total|l_account_id|r_account_id")
	fmt.Println("  --order            desc|asc")
	fmt.Println("  --flat             rows만 compact 배열로 출력")
	fmt.Println("  --section          섹션 ID (전역 플래그)")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whoo entries search --account-id x2 --from 20260301 --to 20260916 --flat")
	fmt.Println("  whoo entries search --l-id x2 --item '이체*'")
	fmt.Println("  whoo entries search --r-account assets --r-id x2")
}

func showEntriesFlowHelp() {
	fmt.Println("사용법: whoo entries flow [플래그]")
	fmt.Println()
	fmt.Println("계정 또는 항목의 흐름 분석을 JSON으로 출력합니다.")
	fmt.Println()
	fmt.Println("플래그:")
	fmt.Println("  --from         시작 날짜 YYYYMMDD (필수)")
	fmt.Println("  --to           종료 날짜 YYYYMMDD (필수)")
	fmt.Println("  --account      계정 (flow_of_account)")
	fmt.Println("  --account-id   항목 ID (flow_of_account_id)")
	fmt.Println("  -h, --help     도움말")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whoo entries flow --from 20260101 --to 20260131 --account expenses")
	fmt.Println("  whoo entries flow --from 20260101 --to 20260131 --account-id x12")
}

// showEntriesHelp는 entries 서브커맨드 도움말 출력
func showEntriesHelp() {
	fmt.Println("사용법: whoo entries [command] [flags]")
	fmt.Println()
	fmt.Println("커맨드:")
	fmt.Println("  (없음)       거래내역 조회 (기본: 이번 달)")
	fmt.Println("  add          단건 거래 추가")
	fmt.Println("  batch        JSON 파일로 일괄 입력")
	fmt.Println("  update       단건/복수 거래 수정 (ID 콤마 구분, 최대 100건)")
	fmt.Println("  delete       단건/복수 삭제")
	fmt.Println("  search       고급 필터 검색 (--max 커서 페이지네이션)")
	fmt.Println("  latest       최근 거래내역")
	fmt.Println("  suggest      최근 아이템 목록 (Suggest)")
	fmt.Println("  flow         계정/항목 흐름 분석")
	fmt.Println("  changes      일일 변동 분석")
	fmt.Println("  agg          계정/항목/거래처별 금액 집계")
	fmt.Println("  outside      외부 데이터(SMS 등) 파싱 입력")
	fmt.Println("  outside_report  인식 실패 소스 보고 (--source)")
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
	fmt.Println("  분할 금액은 항목 설정 단위로 절삭되며 나머지는 첫 회차에 합산됩니다.")
	fmt.Println("  상세: whoo entries add --help")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whoo entries add --l-account expenses --l-id x12 --r-account assets --r-id x5 --money 8000 --item 커피")
	fmt.Println("  whoo entries add --l-account liabilities --l-id x10 --r-account assets --r-id x5 --money 1200000 --item 노트북 --split 3 --fee 2.5")
	fmt.Println("  whoo entries update 1352827,1352828 --memo 정산완료")
	fmt.Println("  whoo entries delete 1352827,1352828")
	fmt.Println("  whoo entries search --item '커피*' --money-from 3000 --money-to 10000")
	fmt.Println("  whoo entries search --account-id x2 --from 20260301 --to 20260916 --flat")
	fmt.Println("  whoo entries search --limit 100 --max 20260203.0034 --sort money --order asc")
	fmt.Println("  whoo entries flow --from 20260101 --to 20260131 --account expenses")
	fmt.Println("  whoo entries changes --from 20260101 --to 20260131 --account-id x12")
	fmt.Println("  whoo entries agg --from 20260101 --to 20260131 --account expenses")
	fmt.Println("  whoo entries agg --from 20260101 --to 20260131 --account-id x12 --by items")
}
