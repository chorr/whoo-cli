// cmd/user_logs.go
// user_logs / user_point_logs CLI 커맨드 — 유저 로그 조회 (JSON 출력)

package cmd

import (
	"flag"
	"fmt"
	"os"

	"whoo-cli/config"
)

// RunUserLogs는 user_logs CLI 커맨드 실행
// API 응답을 파싱하지 않고 JSON 원본을 그대로 출력
func RunUserLogs(cfg *config.Config, args []string) {
	if wantsHelp(args) {
		fmt.Println("사용법: whoo user_logs [플래그]")
		fmt.Println()
		fmt.Println("유저 로그를 JSON으로 출력합니다.")
		fmt.Println()
		fmt.Println("플래그:")
		fmt.Println("  --max N     id가 N보다 작은 로그만 (페이지네이션 커서)")
		fmt.Println("  --limit N   표시할 로그 갯수")
		return
	}
	RequireAuth(cfg)

	fs := flag.NewFlagSet("user_logs", flag.ExitOnError)
	max := fs.Int64("max", 0, "id 커서 (id < max)")
	limit := fs.Int("limit", 0, "표시할 로그 갯수")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.GetUserLogs(*max, *limit)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// RunUserPointLogs는 user_point_logs CLI 커맨드 실행
// GET /api/user_point_logs.json
func RunUserPointLogs(cfg *config.Config, args []string) {
	if wantsHelp(args) {
		fmt.Println("사용법: whoo user_point_logs [플래그]")
		fmt.Println()
		fmt.Println("유저 포인트 로그를 JSON으로 출력합니다.")
		fmt.Println()
		fmt.Println("플래그:")
		fmt.Println("  --max N     point_id가 N보다 작은 로그만 (페이지네이션 커서)")
		fmt.Println("  --type T    포인트 종류 all|affiliate (기본: all)")
		fmt.Println("  --limit N   표시할 로그 갯수")
		return
	}
	RequireAuth(cfg)

	fs := flag.NewFlagSet("user_point_logs", flag.ExitOnError)
	max := fs.Int64("max", 0, "point_id 커서")
	typ := fs.String("type", "", "포인트 종류 all|affiliate")
	limit := fs.Int("limit", 0, "표시할 로그 갯수")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.GetUserPointLogs(*max, *typ, *limit)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}
