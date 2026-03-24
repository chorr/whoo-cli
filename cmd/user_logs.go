// cmd/user_logs.go
// user_logs CLI 커맨드 — 유저 로그 리스트 조회 (JSON 출력)

package cmd

import (
	"os"

	"whooing-cli/config"
)

// RunUserLogs는 user_logs CLI 커맨드 실행
// API 응답을 파싱하지 않고 JSON 원본을 그대로 출력
func RunUserLogs(cfg *config.Config) {
	RequireAuth(cfg)

	client := NewClient(cfg)
	data, err := client.GetUserLogs()
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}
