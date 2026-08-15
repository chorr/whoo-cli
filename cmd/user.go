// cmd/user.go
// user CLI 커맨드 — 유저 정보 조회/수정 (JSON 출력)

package cmd

import (
	"flag"
	"fmt"
	"os"

	"whoo-cli/config"
)

// RunUser는 user CLI 커맨드 실행
// API 응답을 파싱하지 않고 JSON 원본을 그대로 출력
func RunUser(cfg *config.Config, args []string) {
	if wantsHelp(args) {
		showUserHelp()
		return
	}
	RequireAuth(cfg)

	if len(args) > 0 && args[0] == "edit" {
		runUserEdit(cfg, args[1:])
		return
	}

	client := NewClient(cfg)
	data, err := client.GetUser()
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runUserEdit는 유저 정보 수정 (지정한 필드만 전송)
// PUT /api/user.json
func runUserEdit(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("user edit", flag.ExitOnError)
	username := fs.String("username", "", "별명")
	country := fs.String("country", "", "국가 (예: KR)")
	language := fs.String("language", "", "언어 (예: ko)")
	timezone := fs.String("timezone", "", "타임존 (예: Asia/Seoul)")
	currency := fs.String("currency", "", "기본 통화 (예: KRW)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	fields := make(map[string]string)
	if *username != "" {
		fields["username"] = *username
	}
	if *country != "" {
		fields["country"] = *country
	}
	if *language != "" {
		fields["language"] = *language
	}
	if *timezone != "" {
		fields["timezone"] = *timezone
	}
	if *currency != "" {
		fields["currency"] = *currency
	}
	if len(fields) == 0 {
		PrintError("수정할 필드가 없습니다 (--username, --country, --language, --timezone, --currency)")
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.UpdateUser(fields)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

func showUserHelp() {
	fmt.Println("사용법: whoo user [edit] [플래그]")
	fmt.Println()
	fmt.Println("유저 정보를 JSON으로 출력하거나 수정합니다.")
	fmt.Println()
	fmt.Println("커맨드:")
	fmt.Println("  (없음)   유저 정보 조회")
	fmt.Println("  edit     유저 정보 수정 (지정한 필드만 전송)")
	fmt.Println()
	fmt.Println("user edit 플래그:")
	fmt.Println("  --username   별명")
	fmt.Println("  --country    국가 (예: KR)")
	fmt.Println("  --language   언어 (예: ko)")
	fmt.Println("  --timezone   타임존 (예: Asia/Seoul)")
	fmt.Println("  --currency   기본 통화 (예: KRW)")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whoo user")
	fmt.Println("  whoo user edit --username 흥반장 --timezone Asia/Seoul")
}
