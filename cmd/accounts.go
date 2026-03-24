// cmd/accounts.go
// accounts CLI 커맨드 — 항목 관리 (서브커맨드 구조)

package cmd

import (
	"fmt"
	"os"

	"whooing-cli/config"
)

// RunAccounts는 accounts CLI 커맨드 실행
func RunAccounts(cfg *config.Config, args []string) {
	RequireAuth(cfg)
	RequireSection(cfg)

	// 서브커맨드 없으면 전체 항목 목록
	if len(args) == 0 {
		runAccountsList(cfg)
		return
	}

	switch args[0] {
	case "help", "--help", "-h":
		showAccountsHelp()
	default:
		account := args[0]
		if len(args) == 1 {
			// whooing accounts assets → 계정별 항목 목록
			runAccountsByType(cfg, account)
		} else {
			// whooing accounts assets x2 → 특정 항목 조회
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

// showAccountsHelp는 accounts 서브커맨드 도움말 출력
func showAccountsHelp() {
	fmt.Println("사용법: whooing accounts [command]")
	fmt.Println()
	fmt.Println("커맨드:")
	fmt.Println("  (없음)                      전체 항목 목록")
	fmt.Println("  <account>                   계정별 항목 목록")
	fmt.Println("  <account> <account_id>      특정 항목 조회")
	fmt.Println("  help                        도움말 표시")
	fmt.Println()
	fmt.Println("계정 종류:")
	fmt.Println("  assets        자산")
	fmt.Println("  liabilities   부채")
	fmt.Println("  capital       자본")
	fmt.Println("  income        수익")
	fmt.Println("  expenses      비용")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whooing accounts                 전체 항목 목록")
	fmt.Println("  whooing accounts assets          자산 항목 목록")
	fmt.Println("  whooing accounts assets x2       자산 항목 x2 상세")
}
