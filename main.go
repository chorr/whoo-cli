// main.go
// 후잉 CLI 진입점 — TUI/CLI 분기

package main

import (
	"fmt"
	"os"

	"whoo-cli/cmd"
	"whoo-cli/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[오류] 설정 로드 실패: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		if !cmd.HasInteractiveTTY() {
			fmt.Fprintln(os.Stderr, "[오류] TUI는 터미널(TTY)이 필요합니다")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "  whoo help          커맨드 목록")
			fmt.Fprintln(os.Stderr, "  whoo status        인증 상태")
			fmt.Fprintln(os.Stderr, "  whoo auth --help   인증 방법")
			os.Exit(1)
		}
		cmd.RunApp(cfg)
		return
	}

	args := os.Args[2:]
	switch os.Args[1] {
	case "auth", "login":
		cmd.RunAuth(cfg, args)
	case "user":
		cmd.RunUser(cfg, args)
	case "user_logs":
		cmd.RunUserLogs(cfg, args)
	case "user_point_logs":
		cmd.RunUserPointLogs(cfg, args)
	case "sections", "s":
		cmd.RunSections(cfg, args)
	case "accounts", "a":
		cmd.RunAccounts(cfg, args)
	case "entries", "e":
		cmd.RunEntries(cfg, args)
	case "frequent", "freq", "f":
		cmd.RunFrequent(cfg, args)
	case "monthly", "month", "m":
		cmd.RunMonthly(cfg, args)
	case "inout", "io":
		cmd.RunInOut(cfg, args)
	case "bs":
		cmd.RunBS(cfg, args)
	case "report", "r":
		cmd.RunReport(cfg, args)
	case "budget":
		cmd.RunBudget(cfg, args)
	case "budget-goal":
		cmd.RunBudgetGoal(cfg, args)
	case "goal":
		cmd.RunGoal(cfg, args)
	case "bill", "b":
		cmd.RunBill(cfg, args)
	case "checkcard", "cc":
		cmd.RunCheckcard(cfg, args)
	case "status":
		cmd.RunStatus(cfg, args)
	case "version", "--version", "-v":
		cmd.ShowVersion()
	case "help", "--help", "-h":
		cmd.ShowHelp()
	default:
		cmd.ShowHelp()
	}
}
