// cmd/accounts.go
// accounts CLI 커맨드 — 항목 관리 (서브커맨드 구조)

package cmd

import (
	"flag"
	"fmt"
	"os"

	"whoo-cli/api"
	"whoo-cli/config"
)

// RunAccounts는 accounts CLI 커맨드 실행
func RunAccounts(cfg *config.Config, args []string) {
	RequireAuth(cfg)
	RequireSection(cfg)

	if len(args) == 0 {
		runAccountsList(cfg)
		return
	}

	switch args[0] {
	case "add":
		runAccountsAdd(cfg, args[1:])
	case "edit":
		runAccountsEdit(cfg, args[1:])
	case "delete", "del", "rm":
		runAccountsDelete(cfg, args[1:])
	case "exists":
		runAccountsExists(cfg, args[1:])
	case "sort":
		runAccountsSort(cfg, args[1:])
	case "help", "--help", "-h":
		showAccountsHelp()
	default:
		account := args[0]
		if len(args) == 1 {
			runAccountsByType(cfg, account)
		} else {
			runAccountByID(cfg, account, args[1])
		}
	}
}

// runAccountsList는 전체 항목 목록 조회
func runAccountsList(cfg *config.Config) {
	client := NewClient(cfg)
	data, err := client.GetAccountsList(cfg.SectionID)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runAccountsByType은 계정별 항목 목록 조회
func runAccountsByType(cfg *config.Config, account string) {
	client := NewClient(cfg)
	data, err := client.GetAccountsByType(cfg.SectionID, account)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runAccountByID는 특정 항목 상세 조회
func runAccountByID(cfg *config.Config, account, accountID string) {
	client := NewClient(cfg)
	data, err := client.GetAccountByID(cfg.SectionID, account, accountID)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runAccountsAdd는 항목 신규 생성
func runAccountsAdd(cfg *config.Config, args []string) {
	if len(args) == 0 {
		PrintError("계정 타입이 필요합니다 (assets|liabilities|capital|expenses|income)")
		os.Exit(1)
	}
	account := args[0]

	fs := flag.NewFlagSet("accounts add", flag.ExitOnError)
	sectionID := fs.String("section", cfg.SectionID, "섹션 ID")
	title := fs.String("title", "", "항목 이름 (필수)")
	itemType := fs.String("type", "account", "항목 종류 (account|group)")
	openDate := fs.Int("open-date", 20010101, "사용 시작일 (YYYYMMDD)")
	closeDate := fs.Int("close-date", 29991231, "사용 종료일 (YYYYMMDD)")
	memo := fs.String("memo", "", "항목 설명")
	category := fs.String("category", "normal", "항목 종류 (normal|client|creditcard|checkcard|steady|floating)")
	optUseDate := fs.String("opt-use-date", "", "[신용카드] 사용기간 시작일 (pp1~p31)")
	optPayDate := fs.Int("opt-pay-date", 0, "[신용카드] 대금결제일 (1~31)")
	optPayAccID := fs.String("opt-pay-account", "", "[신용카드] 대금결제 자산항목 ID")
	if err := fs.Parse(args[1:]); err != nil {
		os.Exit(1)
	}

	if *title == "" {
		PrintError("--title 은 필수입니다")
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.CreateAccount(account, api.AccountCreateParams{
		SectionID:   *sectionID,
		Title:       *title,
		Type:        *itemType,
		OpenDate:    *openDate,
		CloseDate:   *closeDate,
		Memo:        *memo,
		Category:    *category,
		OptUseDate:  *optUseDate,
		OptPayDate:  *optPayDate,
		OptPayAccID: *optPayAccID,
	})
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runAccountsEdit는 항목 정보 수정
func runAccountsEdit(cfg *config.Config, args []string) {
	if len(args) < 2 {
		PrintError("계정 타입과 account_id가 필요합니다")
		fmt.Fprintln(os.Stderr, "  사용법: whoo accounts edit <type> <account_id> [옵션...]")
		os.Exit(1)
	}
	account := args[0]
	accountID := args[1]

	fs := flag.NewFlagSet("accounts edit", flag.ExitOnError)
	sectionID := fs.String("section", cfg.SectionID, "섹션 ID")
	title := fs.String("title", "", "항목 이름")
	openDate := fs.Int("open-date", 0, "사용 시작일 (YYYYMMDD)")
	closeDate := fs.Int("close-date", 0, "사용 종료일 (YYYYMMDD)")
	memo := fs.String("memo", "", "항목 설명")
	category := fs.String("category", "", "항목 종류")
	optUseDate := fs.String("opt-use-date", "", "[신용카드] 사용기간 시작일")
	optPayDate := fs.Int("opt-pay-date", 0, "[신용카드] 대금결제일")
	optPayAccID := fs.String("opt-pay-account", "", "[신용카드] 대금결제 자산항목 ID")
	if err := fs.Parse(args[2:]); err != nil {
		os.Exit(1)
	}

	// 현재 항목 정보 조회해서 기본값 설정
	client := NewClient(cfg)
	existing, err := client.GetAccountByID(*sectionID, account, accountID)
	if err != nil {
		PrintError("항목 조회 실패: %v", err)
		os.Exit(1)
	}

	var acc struct {
		Results struct {
			Title     string `json:"title"`
			OpenDate  int    `json:"open_date"`
			CloseDate int    `json:"close_date"`
			Memo      string `json:"memo"`
			Category  string `json:"category"`
			OptUseDate  string `json:"opt_use_date"`
			OptPayDate  int    `json:"opt_pay_date"`
			OptPayAccID string `json:"opt_pay_account_id"`
		} `json:"results"`
	}
	if err := parseJSONResponse(existing, &acc); err == nil {
		if *title == "" {
			*title = acc.Results.Title
		}
		if *openDate == 0 {
			*openDate = acc.Results.OpenDate
		}
		if *closeDate == 0 {
			*closeDate = acc.Results.CloseDate
		}
		if *memo == "" {
			*memo = acc.Results.Memo
		}
		if *category == "" {
			*category = acc.Results.Category
		}
		if *optUseDate == "" {
			*optUseDate = acc.Results.OptUseDate
		}
		if *optPayDate == 0 {
			*optPayDate = acc.Results.OptPayDate
		}
		if *optPayAccID == "" {
			*optPayAccID = acc.Results.OptPayAccID
		}
	}

	data, err := client.UpdateAccount(account, accountID, api.AccountUpdateParams{
		SectionID:   *sectionID,
		Title:       *title,
		OpenDate:    *openDate,
		CloseDate:   *closeDate,
		Memo:        *memo,
		Category:    *category,
		OptUseDate:  *optUseDate,
		OptPayDate:  *optPayDate,
		OptPayAccID: *optPayAccID,
	})
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runAccountsDelete는 항목 삭제 (exists 자동 선행)
func runAccountsDelete(cfg *config.Config, args []string) {
	if len(args) < 2 {
		PrintError("계정 타입과 account_id가 필요합니다")
		fmt.Fprintln(os.Stderr, "  사용법: whoo accounts delete <type> <account_id> [--section <id>] [--force]")
		os.Exit(1)
	}
	account := args[0]
	accountID := args[1]

	fs := flag.NewFlagSet("accounts delete", flag.ExitOnError)
	sectionID := fs.String("section", cfg.SectionID, "섹션 ID")
	force := fs.Bool("force", false, "거래 있어도 강제 삭제")
	if err := fs.Parse(args[2:]); err != nil {
		os.Exit(1)
	}

	client := NewClient(cfg)

	// exists 선행 검사
	result, err := client.AccountExists(account, accountID, *sectionID)
	if err != nil {
		PrintError("항목 검사 실패: %v", err)
		os.Exit(1)
	}

	if result.Count > 0 && !*force {
		PrintError("해당 항목에 거래 %d건이 있습니다", result.Count)
		fmt.Fprintf(os.Stderr, "  기간: %d ~ %d\n", result.MinDate, result.MaxDate)
		fmt.Fprintf(os.Stderr, "  잔액: %s\n", FormatMoney(float64(result.Balance)))
		fmt.Fprintf(os.Stderr, "  강제 삭제하려면 --force 플래그를 추가하세요\n")
		os.Exit(1)
	}

	if result.LastOne == "y" {
		PrintError("마지막 항목이라 삭제할 수 없습니다 (last_one=y)")
		os.Exit(1)
	}

	if result.Count > 0 {
		fmt.Fprintf(os.Stderr, "[경고] 거래 %d건이 있는 항목을 강제 삭제합니다. 관련 거래의 항목이 x0으로 변환됩니다.\n", result.Count)
	}

	data, err := client.DeleteAccount(account, accountID, *sectionID)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runAccountsExists는 항목 거래 존재 여부 확인
func runAccountsExists(cfg *config.Config, args []string) {
	if len(args) < 2 {
		PrintError("계정 타입과 account_id가 필요합니다")
		fmt.Fprintln(os.Stderr, "  사용법: whoo accounts exists <type> <account_id> [--section <id>]")
		os.Exit(1)
	}
	account := args[0]
	accountID := args[1]

	fs := flag.NewFlagSet("accounts exists", flag.ExitOnError)
	sectionID := fs.String("section", cfg.SectionID, "섹션 ID")
	if err := fs.Parse(args[2:]); err != nil {
		os.Exit(1)
	}

	client := NewClient(cfg)
	result, err := client.AccountExists(account, accountID, *sectionID)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}

	// 구조화된 JSON 출력
	out, _ := marshalPretty(result)
	fmt.Println(string(out))
}

// runAccountsSort는 항목 순서 변경
func runAccountsSort(cfg *config.Config, args []string) {
	if len(args) < 2 {
		PrintError("계정 타입과 account_id 목록이 필요합니다")
		fmt.Fprintln(os.Stderr, "  사용법: whoo accounts sort <type> <x1,x2,x3> [--section <id>]")
		os.Exit(1)
	}
	account := args[0]

	fs := flag.NewFlagSet("accounts sort", flag.ExitOnError)
	sectionID := fs.String("section", cfg.SectionID, "섹션 ID")
	if err := fs.Parse(args[1:]); err != nil {
		os.Exit(1)
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		PrintError("account_id 목록이 필요합니다")
		os.Exit(1)
	}
	ids := splitCommaOrArgs(remaining)

	client := NewClient(cfg)
	data, err := client.SortAccounts(account, *sectionID, ids)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// showAccountsHelp는 accounts 서브커맨드 도움말 출력
func showAccountsHelp() {
	fmt.Println("사용법: whoo accounts [command]")
	fmt.Println()
	fmt.Println("커맨드:")
	fmt.Println("  (없음)                              전체 항목 목록")
	fmt.Println("  add <type> --title <이름> [옵션...]  항목 생성")
	fmt.Println("  edit <type> <account_id> [옵션...]  항목 수정")
	fmt.Println("  delete <type> <account_id> [옵션]  항목 삭제")
	fmt.Println("  exists <type> <account_id>          거래 존재 확인")
	fmt.Println("  sort <type> <x1,x2,x3>              순서 변경")
	fmt.Println("  <type>                              계정별 항목 목록")
	fmt.Println("  <type> <account_id>                 특정 항목 조회")
	fmt.Println("  help                                도움말")
	fmt.Println()
	fmt.Println("계정 타입: assets|liabilities|capital|expenses|income")
	fmt.Println()
	fmt.Println("accounts add 옵션:")
	fmt.Println("  --title         항목 이름 (필수)")
	fmt.Println("  --section       섹션 ID (기본: 현재 섹션)")
	fmt.Println("  --type          account|group (기본: account)")
	fmt.Println("  --open-date     사용 시작일 YYYYMMDD (기본: 20010101)")
	fmt.Println("  --close-date    사용 종료일 YYYYMMDD (기본: 29991231)")
	fmt.Println("  --memo          설명")
	fmt.Println("  --category      normal|client|creditcard|checkcard|steady|floating")
	fmt.Println("  --opt-use-date  [신용카드] 사용기간 시작일 (pp1~p31)")
	fmt.Println("  --opt-pay-date  [신용카드] 대금결제일 (1~31)")
	fmt.Println("  --opt-pay-account [신용카드] 대금결제 자산항목 ID")
	fmt.Println()
	fmt.Println("accounts delete 옵션:")
	fmt.Println("  --force   거래 있어도 강제 삭제 (거래 항목이 x0으로 변환됨)")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whoo accounts add assets --title \"토스\" --category normal")
	fmt.Println("  whoo accounts add liabilities --title \"신한카드\" --category creditcard --opt-use-date pp1 --opt-pay-date 25 --opt-pay-account x2")
	fmt.Println("  whoo accounts edit assets x2 --title \"새 이름\"")
	fmt.Println("  whoo accounts delete assets x2")
	fmt.Println("  whoo accounts exists assets x2")
	fmt.Println("  whoo accounts sort assets x2,x4,x3,x5")
}
