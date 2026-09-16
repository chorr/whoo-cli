// cmd/bs.go
// bs CLI 커맨드 — 자산/부채 잔액(Balance Sheet) 조회

package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"whoo-cli/api"
	"whoo-cli/config"
)

type balanceOutput struct {
	SectionID string  `json:"section_id"`
	AccountID string  `json:"account_id"`
	Title     string  `json:"title"`
	Type      string  `json:"type"`
	Money     float64 `json:"money"`
	AsOf      string  `json:"as_of"`
}

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
	account := fs.String("account", "", "계정 필터 (assets | liabilities)")
	accountID := fs.String("account-id", "", "단일 항목 ID")
	flat := fs.Bool("flat", false, "항목별 compact 배열")
	if err := fs.Parse(args); err != nil {
		PrintError("%v", err)
		showBSHelp()
		os.Exit(1)
	}
	if *account != "" && *account != "assets" && *account != "liabilities" {
		PrintError("--account는 assets 또는 liabilities여야 합니다")
		os.Exit(1)
	}

	if *end == "" {
		*end = time.Now().Format("20060102")
	}

	client := NewClient(cfg)
	if *flat || *account != "" || *accountID != "" {
		bs, err := client.GetBS(cfg.SectionID, *end)
		if err != nil {
			PrintError("%v", err)
			os.Exit(1)
		}
		accounts, _ := client.GetAccountsMap(cfg.SectionID)
		rows := buildBalanceRows(cfg.SectionID, *end, bs, accounts, *account)
		if *accountID != "" {
			for _, row := range rows {
				if row.AccountID == *accountID {
					printJSONValue(row)
					return
				}
			}
			if accounts != nil {
				for _, accountType := range []string{"assets", "liabilities"} {
					if *account != "" && *account != accountType {
						continue
					}
					if detail, ok := accounts.GetAccountsByType(accountType)[*accountID]; ok {
						printJSONValue(balanceOutput{
							SectionID: cfg.SectionID,
							AccountID: *accountID,
							Title:     detail.Title,
							Type:      accountType,
							Money:     0,
							AsOf:      *end,
						})
						return
					}
				}
			}
			scope := "자산/부채"
			if *account != "" {
				scope = *account
			}
			PrintError("%s에서 account_id %s의 잔액을 찾을 수 없습니다", scope, *accountID)
			os.Exit(1)
		}
		printJSONValue(rows)
		return
	}

	data, err := client.GetBSRaw(cfg.SectionID, *end)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

func buildBalanceRows(
	sectionID, endDate string,
	bs *api.BSResponse,
	accounts *api.AccountsMap,
	accountFilter string,
) []balanceOutput {
	rows := make([]balanceOutput, 0, len(bs.Assets.Accounts)+len(bs.Liabilities.Accounts))
	groups := []struct {
		accountType string
		accounts    []api.BSAccount
	}{
		{"assets", bs.Assets.Accounts},
		{"liabilities", bs.Liabilities.Accounts},
	}
	for _, group := range groups {
		if accountFilter != "" && accountFilter != group.accountType {
			continue
		}
		for _, account := range group.accounts {
			title := account.AccountID
			if accounts != nil {
				title = accounts.GetTitle(group.accountType, account.AccountID)
			}
			rows = append(rows, balanceOutput{
				SectionID: sectionID,
				AccountID: account.AccountID,
				Title:     title,
				Type:      group.accountType,
				Money:     account.Money,
				AsOf:      endDate,
			})
		}
	}
	return rows
}

func showBSHelp() {
	fmt.Println("사용법: whoo bs|balance [플래그]")
	fmt.Println()
	fmt.Println("현재 섹션의 자산/부채 잔액을 JSON으로 출력합니다.")
	fmt.Println("공식 report API를 자산/부채로 나눠 호출하므로 콤마 경로 403과 무관합니다.")
	fmt.Println("항목 메타만 보려면 whoo accounts, 기간 증감은 whoo inout 을 사용하세요.")
	fmt.Println("기간별 보고서가 필요하면 whoo report 를 사용하세요.")
	fmt.Println()
	fmt.Println("플래그:")
	fmt.Println("  --end          기준일 YYYYMMDD (기본: 오늘)")
	fmt.Println("  --account      계정 필터 (assets | liabilities)")
	fmt.Println("  --account-id   단일 항목 ID")
	fmt.Println("  --flat         항목별 compact 배열")
	fmt.Println("  --section      섹션 ID (전역 플래그, WHOO_SECTION보다 우선)")
	fmt.Println("  -h, --help     도움말")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whoo bs")
	fmt.Println("  whoo bs --end 20260813")
	fmt.Println("  whoo balance --account-id x2")
	fmt.Println("  whoo bs --account assets --flat")
	fmt.Println("  WHOO_SECTION=s36258 whoo bs")
}
