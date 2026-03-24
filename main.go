// main.go
// 후잉 CLI 진입점 — TUI/CLI 분기

package main

import (
	"fmt"
	"os"

	"whooing-cli/cmd"
	"whooing-cli/config"
)

func main() {
	// 설정 로드
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[오류] 설정 로드 실패: %v\n", err)
		os.Exit(1)
	}

	// 인자 없이 실행 시 통합 앱 실행
	if len(os.Args) < 2 {
		cmd.RunApp(cfg)
		return
	}

	// 서브커맨드 디스패치
	switch os.Args[1] {
	case "auth", "login":
		cmd.RunApp(cfg)
	case "user":
		cmd.RunUser(cfg)
	case "user_logs":
		cmd.RunUserLogs(cfg)
	case "sections", "s":
		cmd.RunSections(cfg, os.Args[2:])
	case "accounts", "a":
		cmd.RunAccounts(cfg, os.Args[2:])
	case "entries", "e":
		cmd.RunEntries(cfg, os.Args[2:])
	case "status":
		cmd.RunStatus(cfg)
	case "help", "--help", "-h":
		cmd.ShowHelp()
	default:
		cmd.ShowHelp()
	}
}
