// cmd/user.go
// user CLI 커맨드 — 유저 정보 조회 (JSON 출력)

package cmd

import (
	"fmt"
	"os"

	"whoo-cli/config"
)

// RunUser는 user CLI 커맨드 실행
// API 응답을 파싱하지 않고 JSON 원본을 그대로 출력
func RunUser(cfg *config.Config, args []string) {
	if wantsHelp(args) {
		fmt.Println("사용법: whoo user")
		fmt.Println()
		fmt.Println("유저 정보를 JSON으로 출력합니다.")
		return
	}
	RequireAuth(cfg)

	client := NewClient(cfg)
	data, err := client.GetUser()
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}
