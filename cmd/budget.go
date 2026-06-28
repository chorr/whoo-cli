// cmd/budget.go
// budget / budget-goal / goal CLI 커맨드

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

// ─── budget ──────────────────────────────────────────────────

// RunBudget는 budget CLI 커맨드 실행
func RunBudget(cfg *config.Config, args []string) {
	RequireAuth(cfg)
	RequireSection(cfg)

	if len(args) == 0 {
		showBudgetHelp()
		return
	}

	switch args[0] {
	case "get":
		runBudgetGet(cfg, args[1:])
	case "set":
		runBudgetSet(cfg, args[1:])
	case "reset":
		runBudgetReset(cfg, args[1:])
	case "help", "--help", "-h":
		showBudgetHelp()
	default:
		fmt.Fprintf(os.Stderr, "[오류] 알 수 없는 서브커맨드: %s\n", args[0])
		showBudgetHelp()
		os.Exit(1)
	}
}

func runBudgetGet(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("budget get", flag.ExitOnError)
	from := fs.Int("from", 0, "시작 월 YYYYMM (기본: 이번달)")
	to := fs.Int("to", 0, "종료 월 YYYYMM (기본: 이번달)")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "[오류] 계정 타입 필요: expenses | income")
		os.Exit(1)
	}
	account := fs.Arg(0)
	if account != "expenses" && account != "income" {
		fmt.Fprintln(os.Stderr, "[오류] 계정 타입은 expenses 또는 income")
		os.Exit(1)
	}

	now := time.Now()
	ym := int(now.Year())*100 + int(now.Month())
	if *from == 0 {
		*from = ym
	}
	if *to == 0 {
		*to = ym
	}

	client := NewClient(cfg)
	resp, err := client.GetBudget(cfg.SectionID, account, *from, *to)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[오류] %v\n", err)
		os.Exit(1)
	}
	printTypedJSON(resp)
}

func runBudgetSet(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("budget set", flag.ExitOnError)
	ym := fs.Int("ym", 0, "대상 월 YYYYMM (필수)")
	accountCSV := fs.String("account", "", "항목=금액 (쉼표 구분, 예: x50=150000,x51=80000)")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "[오류] 계정 타입 필요: expenses | income")
		os.Exit(1)
	}
	account := fs.Arg(0)
	if *ym == 0 {
		fmt.Fprintln(os.Stderr, "[오류] --ym 필수")
		os.Exit(1)
	}
	if *accountCSV == "" {
		fmt.Fprintln(os.Stderr, "[오류] --account 필수")
		os.Exit(1)
	}

	budgets, err := api.ParseBudgetAccountCSV(*accountCSV)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[오류] %v\n", err)
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.UpdateBudget(cfg.SectionID, account, *ym, budgets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[오류] %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func runBudgetReset(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("budget reset", flag.ExitOnError)
	from := fs.Int("from", 0, "시작 월 YYYYMM (필수)")
	to := fs.Int("to", 0, "종료 월 YYYYMM (필수)")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "[오류] 계정 타입 필요: expenses | income")
		os.Exit(1)
	}
	account := fs.Arg(0)
	if *from == 0 || *to == 0 {
		fmt.Fprintln(os.Stderr, "[오류] --from, --to 필수")
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.DeleteBudget(cfg.SectionID, account, *from, *to)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[오류] %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func showBudgetHelp() {
	fmt.Println(`사용법: whoo budget <서브커맨드> <expenses|income> [옵션]

서브커맨드:
  get <expenses|income> [--from YYYYMM] [--to YYYYMM]
  set <expenses|income> --ym YYYYMM --account id=금액,id2=금액2
  reset <expenses|income> --from YYYYMM --to YYYYMM`)
}

// ─── budget-goal ──────────────────────────────────────────────

// RunBudgetGoal는 budget-goal CLI 커맨드 실행
func RunBudgetGoal(cfg *config.Config, args []string) {
	RequireAuth(cfg)
	RequireSection(cfg)

	if len(args) == 0 {
		showBudgetGoalHelp()
		return
	}

	switch args[0] {
	case "get":
		runBudgetGoalGet(cfg)
	case "set":
		runBudgetGoalSet(cfg, args[1:])
	case "reset":
		runBudgetGoalReset(cfg)
	case "help", "--help", "-h":
		showBudgetGoalHelp()
	default:
		fmt.Fprintf(os.Stderr, "[오류] 알 수 없는 서브커맨드: %s\n", args[0])
		showBudgetGoalHelp()
		os.Exit(1)
	}
}

func runBudgetGoalGet(cfg *config.Config) {
	client := NewClient(cfg)
	resp, err := client.GetBudgetGoal(cfg.SectionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[오류] %v\n", err)
		os.Exit(1)
	}
	printTypedJSON(resp)
}

func runBudgetGoalSet(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("budget-goal set", flag.ExitOnError)
	baseYM := fs.Int("base-ym", 0, "시작 월 YYYYMM (필수)")
	goalYM := fs.Int("goal-ym", 0, "목표 월 YYYYMM (필수)")
	goalMoney := fs.Int64("goal-money", 0, "목표 자본 금액 (필수)")
	baseIncome := fs.Int64("base-income", 0, "월 기준 수입")
	baseExpenses := fs.Int64("base-expenses", 0, "월 기준 지출")
	splitType := fs.String("split-type", "equal", "배분 방식 (auto|equal|manual)")
	fs.Parse(args)

	if *baseYM == 0 || *goalYM == 0 || *goalMoney == 0 {
		fmt.Fprintln(os.Stderr, "[오류] --base-ym, --goal-ym, --goal-money 필수")
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.UpdateBudgetGoal(cfg.SectionID, api.BudgetGoalParams{
		BaseYM:       *baseYM,
		GoalYM:       *goalYM,
		GoalMoney:    *goalMoney,
		BaseIncome:   *baseIncome,
		BaseExpenses: *baseExpenses,
		SplitType:    *splitType,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "[오류] %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func runBudgetGoalReset(cfg *config.Config) {
	fmt.Print("[경고] 이 작업은 장기목표·자본목표·예산 데이터를 모두 초기화합니다. 되돌릴 수 없습니다.\n정말 삭제하려면 'DELETE'를 입력하세요: ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "DELETE" {
		fmt.Println("취소되었습니다.")
		return
	}

	client := NewClient(cfg)
	data, err := client.DeleteBudgetGoal(cfg.SectionID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[오류] %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func showBudgetGoalHelp() {
	fmt.Println(`사용법: whoo budget-goal <서브커맨드> [옵션]

서브커맨드:
  get
  set  --base-ym YYYYMM --goal-ym YYYYMM --goal-money N
       [--base-income N] [--base-expenses N] [--split-type auto|equal|manual]
  reset    # 되돌릴 수 없음, 이중 확인 필요`)
}

// ─── goal ─────────────────────────────────────────────────────

// RunGoal는 goal CLI 커맨드 실행
func RunGoal(cfg *config.Config, args []string) {
	RequireAuth(cfg)
	RequireSection(cfg)

	if len(args) == 0 {
		showGoalHelp()
		return
	}

	switch args[0] {
	case "get":
		runGoalGet(cfg, args[1:])
	case "set":
		runGoalSet(cfg, args[1:])
	case "help", "--help", "-h":
		showGoalHelp()
	default:
		fmt.Fprintf(os.Stderr, "[오류] 알 수 없는 서브커맨드: %s\n", args[0])
		showGoalHelp()
		os.Exit(1)
	}
}

func runGoalGet(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("goal get", flag.ExitOnError)
	from := fs.Int("from", 0, "시작 월 YYYYMM")
	to := fs.Int("to", 0, "종료 월 YYYYMM")
	fs.Parse(args)

	now := time.Now()
	ym := int(now.Year())*100 + int(now.Month())
	if *from == 0 {
		*from = ym
	}
	if *to == 0 {
		*to = *from + 11 // 기본: 1년치
	}

	client := NewClient(cfg)
	resp, err := client.GetGoal(cfg.SectionID, *from, *to)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[오류] %v\n", err)
		os.Exit(1)
	}
	printTypedJSON(resp)
}

// ymAmounts는 반복 플래그 --ym YYYYMM=amount 파싱용
type ymAmounts []string

func (y *ymAmounts) String() string  { return strings.Join(*y, ",") }
func (y *ymAmounts) Set(v string) error { *y = append(*y, v); return nil }

func runGoalSet(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("goal set", flag.ExitOnError)
	var yms ymAmounts
	fs.Var(&yms, "ym", "YYYYMM=금액 (반복 지정 가능)")
	fs.Parse(args)

	if len(yms) == 0 {
		fmt.Fprintln(os.Stderr, "[오류] --ym 필수 (예: --ym 202604=21000000)")
		os.Exit(1)
	}

	goals := make(map[int]int64, len(yms))
	for _, kv := range yms {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "[오류] 잘못된 형식: %q (YYYYMM=금액 필요)\n", kv)
			os.Exit(1)
		}
		ym, err := strconv.Atoi(parts[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "[오류] 월 파싱 오류: %q\n", parts[0])
			os.Exit(1)
		}
		amount, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[오류] 금액 파싱 오류: %q\n", parts[1])
			os.Exit(1)
		}
		goals[ym] = amount
	}

	client := NewClient(cfg)
	data, err := client.UpdateGoal(cfg.SectionID, goals)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[오류] %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func showGoalHelp() {
	fmt.Println(`사용법: whoo goal <서브커맨드> [옵션]

서브커맨드:
  get  [--from YYYYMM] [--to YYYYMM]
  set  --ym YYYYMM=금액 [--ym YYYYMM=금액 ...]`)
}

// ─── 내부 헬퍼 ───────────────────────────────────────────────

func printTypedJSON(v interface{}) {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[오류] JSON 직렬화 실패: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
