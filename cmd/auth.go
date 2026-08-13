// cmd/auth.go
// auth CLI — TTY가 있으면 TUI, 없으면 URL/PIN 헤드리스 인증

package cmd

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"whoo-cli/auth"
	"whoo-cli/config"
)

// RunAuth는 auth/login 커맨드 실행
func RunAuth(cfg *config.Config, args []string) {
	if wantsHelp(args) {
		showAuthHelp()
		return
	}

	fs := flag.NewFlagSet("auth", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	urlOnly := fs.Bool("url", false, "인증 URL만 출력")
	pinFlag := fs.String("pin", "", "OAuth PIN")
	if err := fs.Parse(args); err != nil {
		PrintError("%v", err)
		showAuthHelp()
		os.Exit(1)
	}

	pin := strings.TrimSpace(*pinFlag)
	if *urlOnly {
		if pin != "" {
			PrintError("--url 과 --pin 은 함께 사용할 수 없습니다")
			showAuthHelp()
			os.Exit(1)
		}
		runAuthPrintURL(cfg)
		return
	}
	if pin == "" {
		pin = strings.TrimSpace(os.Getenv("WHOOING_PIN"))
	}

	if pin != "" {
		runAuthExchangePIN(cfg, pin)
		return
	}
	if !HasInteractiveTTY() {
		runAuthPrintURL(cfg)
		return
	}
	RunApp(cfg)
}

func runAuthPrintURL(cfg *config.Config) {
	oauth := auth.NewOAuth(cfg)
	tokenResp, err := oauth.RequestToken()
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	if err := oauth.SavePendingRequest(tokenResp.Token, tokenResp.Signiture); err != nil {
		PrintError("요청 토큰 저장 실패: %v", err)
		os.Exit(1)
	}

	authURL := oauth.GetAuthorizationURL(tokenResp.Token)
	fmt.Println(authURL)
	fmt.Fprintln(os.Stderr, "브라우저에서 위 URL을 연 뒤 PIN을 확인하세요.")
	fmt.Fprintln(os.Stderr, "그 다음: whoo auth --pin <PIN>")
}

func runAuthExchangePIN(cfg *config.Config, pin string) {
	if cfg.PendingRequestToken == "" {
		PrintError("요청 토큰이 없습니다. 먼저 whoo auth --url 을 실행하세요")
		os.Exit(1)
	}

	oauth := auth.NewOAuth(cfg)
	access, err := oauth.ExchangeToken(cfg.PendingRequestToken, cfg.PendingSigniture, pin)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	if err := oauth.CompleteAuth(access.Token, access.TokenSecret); err != nil {
		PrintError("토큰 저장 실패: %v", err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, "인증이 완료되었습니다.")
}

func showAuthHelp() {
	fmt.Println("사용법: whoo auth [플래그]")
	fmt.Println()
	fmt.Println("후잉 OAuth PIN 인증.")
	fmt.Println()
	fmt.Println("플래그:")
	fmt.Println("  --url          인증 URL만 stdout에 출력 (TTY 불필요)")
	fmt.Println("  --pin <PIN>    PIN으로 액세스 토큰 교환")
	fmt.Println("  -h, --help     도움말")
	fmt.Println()
	fmt.Println("환경변수:")
	fmt.Println("  WHOOING_PIN    --pin 과 동일")
	fmt.Println()
	fmt.Println("TTY가 있으면 기본으로 TUI 인증을 실행합니다.")
	fmt.Println("TTY가 없으면 --url 과 같이 인증 URL을 출력합니다.")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whoo auth --url")
	fmt.Println("  whoo auth --pin 123456")
	fmt.Println("  WHOOING_PIN=123456 whoo auth")
}
