// cmd/root.go
// 공통 헬퍼 함수

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"whooing-cli/api"
	"whooing-cli/config"
)

// 버전 정보
const Version = "0.3.0"

// RequireAuth는 인증 상태를 확인하고 미인증 시 안내 후 종료
func RequireAuth(cfg *config.Config) {
	if !cfg.IsAuthenticated() {
		fmt.Fprintln(os.Stderr, "[오류] 인증이 필요합니다")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  whooing auth  인증 진행")
		os.Exit(1)
	}
}

// RequireSection은 섹션 선택 상태를 확인하고 미선택 시 안내 후 종료
func RequireSection(cfg *config.Config) {
	if cfg.SectionID == "" {
		fmt.Fprintln(os.Stderr, "[오류] 섹션이 선택되지 않았습니다")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  whooing section  섹션 목록 확인 및 선택")
		os.Exit(1)
	}
}

// NewClient는 인증된 API 클라이언트를 생성
func NewClient(cfg *config.Config) *api.WhooingClient {
	return api.NewWhooingClient(cfg)
}

// FormatMoney는 금액을 쉼표 구분 형식으로 변환
// 예: 1234567.00 → "1,234,567"
func FormatMoney(amount float64) string {
	// 정수 부분만 표시 (소수점 이하 버림)
	intAmount := int64(amount)
	if intAmount < 0 {
		return "-" + formatPositiveMoney(-intAmount)
	}
	return formatPositiveMoney(intAmount)
}

// formatPositiveMoney는 양수 금액을 쉼표 구분 형식으로 변환
func formatPositiveMoney(amount int64) string {
	s := fmt.Sprintf("%d", amount)
	n := len(s)
	if n <= 3 {
		return s
	}

	var result strings.Builder
	remainder := n % 3
	if remainder > 0 {
		result.WriteString(s[:remainder])
	}
	for i := remainder; i < n; i += 3 {
		if result.Len() > 0 {
			result.WriteByte(',')
		}
		result.WriteString(s[i : i+3])
	}
	return result.String()
}

// FormatDate는 YYYYMMDD 형식을 YYYY-MM-DD로 변환
func FormatDate(date string) string {
	if len(date) == 8 {
		return date[:4] + "-" + date[4:6] + "-" + date[6:]
	}
	return date
}

// FormatAccount는 계정 타입 코드를 한글로 변환
func FormatAccount(account string) string {
	for _, at := range AccountTypes {
		if at.Code == account {
			return at.Name
		}
	}
	return account
}

// AccountTypes는 계정 타입 목록 (순서 보장)
var AccountTypes = []struct {
	Code string
	Name string
}{
	{"assets", "자산"},
	{"liabilities", "부채"},
	{"capital", "자본"},
	{"expenses", "비용"},
	{"income", "수익"},
}

// PrintError는 오류 메시지를 stderr에 출력
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[오류] "+format+"\n", args...)
}

// printJSON은 raw JSON 바이트를 pretty-print 출력
func printJSON(data []byte) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		fmt.Println(string(data))
		return
	}
	fmt.Println(buf.String())
}

// ShowHelp는 도움말을 출력
func ShowHelp() {
	fmt.Printf("후잉 CLI v%s\n", Version)
	fmt.Println()
	fmt.Println("사용법: whooing [command]")
	fmt.Println()
	fmt.Println("커맨드:")
	fmt.Println("  (없음)       인터랙티브 TUI 실행")
	fmt.Println("  auth         OAuth 인증")
	fmt.Println("  status       인증/설정 상태 확인")
	fmt.Println("  user         유저 정보 조회")
	fmt.Println("  user_logs    유저 로그 조회")
	fmt.Println("  sections     섹션 관리 (sections help 참조)")
	fmt.Println("  accounts     항목 관리 (accounts help 참조)")
	fmt.Println("  entries      거래내역 조회 (entries help 참조)")
	fmt.Println("  help         도움말 표시")
	fmt.Println()
	fmt.Println("단축:")
	fmt.Println("  s = sections, a = accounts, e = entries")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whooing                  TUI 실행")
	fmt.Println("  whooing entries           이번 달 거래내역")
	fmt.Println("  whooing accounts assets   자산 항목 목록")
	fmt.Println("  whooing sections          섹션 목록")
}
