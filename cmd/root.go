// cmd/root.go
// 공통 헬퍼 함수

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"whoo-cli/api"
	"whoo-cli/config"
)

// 버전 정보
const Version = "1.2.0"

// GlobalOptions는 모든 CLI 커맨드에서 공통으로 처리하는 실행 옵션이다.
type GlobalOptions struct {
	ConfiguredSection string
	SelectedSection   string
	SectionOverridden bool
	AllowOverride     bool
}

var outputSectionID string

// ParseGlobalOptions는 커맨드 위치와 무관하게 전역 플래그를 제거하고 설정에 반영한다.
// 우선순위: --section > WHOO_SECTION > config.json.
func ParseGlobalOptions(cfg *config.Config, args []string) ([]string, GlobalOptions, error) {
	options := GlobalOptions{ConfiguredSection: cfg.SectionID}
	selectedSection := strings.TrimSpace(os.Getenv("WHOO_SECTION"))
	if selectedSection != "" {
		options.SectionOverridden = selectedSection != cfg.SectionID
	}

	remaining := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--section":
			if i+1 >= len(args) || strings.TrimSpace(args[i+1]) == "" {
				return nil, options, fmt.Errorf("--section 값이 필요합니다")
			}
			i++
			selectedSection = strings.TrimSpace(args[i])
			options.SectionOverridden = selectedSection != cfg.SectionID
		case strings.HasPrefix(arg, "--section="):
			selectedSection = strings.TrimSpace(strings.TrimPrefix(arg, "--section="))
			if selectedSection == "" {
				return nil, options, fmt.Errorf("--section 값이 필요합니다")
			}
			options.SectionOverridden = selectedSection != cfg.SectionID
		case arg == "--verbose":
			cfg.Verbose = true
		case arg == "--allow-section-override":
			options.AllowOverride = true
		default:
			remaining = append(remaining, arg)
		}
	}

	if selectedSection != "" {
		cfg.SectionID = selectedSection
	}
	options.SelectedSection = cfg.SectionID
	SetOutputSection(cfg.SectionID)
	return remaining, options, nil
}

// WarnSectionOverride는 다른 섹션에 쓰는 커맨드일 때 stderr로 명확히 알린다.
func WarnSectionOverride(options GlobalOptions, args []string) {
	if !options.SectionOverridden || options.AllowOverride || !isSectionWriteCommand(args) {
		return
	}
	fmt.Fprintf(
		os.Stderr,
		"[경고] 현재 섹션 %s 대신 %s에 기록합니다. 경고를 숨기려면 --allow-section-override를 사용하세요.\n",
		displaySection(options.ConfiguredSection),
		displaySection(options.SelectedSection),
	)
}

func displaySection(sectionID string) string {
	if sectionID == "" {
		return "(미설정)"
	}
	return sectionID
}

func isSectionWriteCommand(args []string) bool {
	if len(args) < 2 {
		return false
	}
	command, subcommand := args[0], args[1]
	switch command {
	case "accounts", "a":
		return subcommand == "add" || subcommand == "edit" || subcommand == "delete" ||
			subcommand == "del" || subcommand == "rm" || subcommand == "sort"
	case "entries", "e":
		return subcommand == "add" || subcommand == "batch" || subcommand == "update" ||
			subcommand == "delete" || subcommand == "del" || subcommand == "rm" ||
			subcommand == "outside"
	case "frequent", "freq", "f", "monthly", "month", "m":
		return subcommand == "add" || subcommand == "edit" || subcommand == "delete" ||
			subcommand == "del" || subcommand == "rm" || subcommand == "sort" ||
			subcommand == "use" || subcommand == "pay"
	case "budget":
		return subcommand == "set" || subcommand == "reset"
	case "budget-goal", "goal":
		return subcommand == "set"
	default:
		return false
	}
}

// SetOutputSection은 이후 JSON 출력에 포함할 실제 섹션을 설정한다.
func SetOutputSection(sectionID string) {
	outputSectionID = sectionID
}

// RequireAuth는 인증 상태를 확인하고 미인증 시 안내 후 종료
func RequireAuth(cfg *config.Config) {
	if !cfg.IsAuthenticated() {
		fmt.Fprintln(os.Stderr, "[오류] 인증이 필요합니다")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  whoo auth  인증 진행")
		os.Exit(1)
	}
}

// RequireSection은 섹션 선택 상태를 확인하고 미선택 시 안내 후 종료
func RequireSection(cfg *config.Config) {
	if cfg.SectionID == "" {
		fmt.Fprintln(os.Stderr, "[오류] 섹션이 선택되지 않았습니다")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  whoo sections       섹션 목록")
		fmt.Fprintln(os.Stderr, "  whoo sections set   섹션 선택")
		fmt.Fprintln(os.Stderr, "  whoo                TUI에서 섹션 선택")
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

// parseJSONResponse는 raw API 응답 바이트를 임의 구조체로 언마샬
// sections edit 등에서 기존 값 읽기 용도로 사용
func parseJSONResponse(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}

// printJSON은 raw JSON 바이트를 pretty-print 출력
func printJSON(data []byte) {
	var value interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		fmt.Println(string(data))
		return
	}
	printJSONValue(value)
}

// printJSONValue는 구조화 값에 실행 섹션을 추가해 JSON으로 출력한다.
func printJSONValue(value interface{}) {
	data, err := json.Marshal(value)
	if err != nil {
		PrintError("JSON 직렬화 실패: %v", err)
		return
	}

	var normalized interface{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&normalized); err != nil {
		PrintError("JSON 변환 실패: %v", err)
		return
	}
	addOutputSection(normalized)

	out, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		PrintError("JSON 출력 실패: %v", err)
		return
	}
	fmt.Println(string(out))
}

func addOutputSection(value interface{}) {
	if outputSectionID == "" {
		return
	}
	switch root := value.(type) {
	case map[string]interface{}:
		if _, exists := root["section_id"]; !exists {
			root["section_id"] = outputSectionID
		}
	case []interface{}:
		for _, item := range root {
			if row, ok := item.(map[string]interface{}); ok {
				if _, exists := row["section_id"]; !exists {
					row["section_id"] = outputSectionID
				}
			}
		}
	}
}

// ShowHelp는 도움말을 출력
func ShowHelp() {
	fmt.Printf("후잉 CLI v%s\n", Version)
	fmt.Println()
	fmt.Println("사용법: whoo [전역 플래그] [command]")
	fmt.Println()
	fmt.Println("전역 플래그:")
	fmt.Println("  --section <id>             실행할 섹션 (WHOO_SECTION, config.json보다 우선)")
	fmt.Println("  --allow-section-override   다른 섹션 쓰기 경고 숨김")
	fmt.Println("  --verbose                  API 오류 원문 포함")
	fmt.Println()
	fmt.Println("커맨드:")
	fmt.Println("  (없음)         인터랙티브 TUI 실행 (TTY 필요)")
	fmt.Println("  auth           OAuth 인증 (auth --help 참조)")
	fmt.Println("  status         인증/설정 상태 확인")
	fmt.Println("  user           유저 정보 조회/수정 (user help 참조)")
	fmt.Println("  user_logs      유저 로그 조회")
	fmt.Println("  user_point_logs  유저 포인트 로그 조회")
	fmt.Println("  sections       섹션 관리 (sections help 참조)")
	fmt.Println("  accounts       항목 관리 (accounts help 참조)")
	fmt.Println("  entries        거래내역 (entries help 참조)")
	fmt.Println("  frequent       자주입력 관리 (frequent help 참조)")
	fmt.Println("  monthly        월별입력 관리 (monthly help 참조)")
	fmt.Println("  inout          자금증감 조회 (inout --help 참조)")
	fmt.Println("  bs             자산/부채 잔액 (bs --help 참조)")
	fmt.Println("  balance        bs 별칭; 단일 항목 잔액 조회")
	fmt.Println("  report         통합 보고서 (report help 참조)")
	fmt.Println("  budget         예산 관리 (budget help 참조)")
	fmt.Println("  budget-goal    예산 목표")
	fmt.Println("  goal           자본 목표")
	fmt.Println("  bill           신용카드 청구 (bill help 참조)")
	fmt.Println("  checkcard      체크카드 연동 (checkcard help 참조)")
	fmt.Println("  version        버전 표시")
	fmt.Println("  help           도움말 표시")
	fmt.Println()
	fmt.Println("단축:")
	fmt.Println("  s = sections, a = accounts, e = entries, r = report")
	fmt.Println("  f = frequent, m = monthly, b = bill, cc = checkcard, io = inout")
	fmt.Println()
	fmt.Println("help, --help, -h 는 모든 커맨드에서 예약어입니다. 인증·API·TUI를 실행하지 않습니다.")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whoo                   TUI 실행")
	fmt.Println("  whoo auth --url        인증 URL 출력 (헤드리스)")
	fmt.Println("  whoo entries           이번 달 거래내역")
	fmt.Println("  whoo accounts assets   자산 항목 메타 (잔액 없음)")
	fmt.Println("  whoo bs                자산/부채 잔액")
	fmt.Println("  whoo balance --account-id x2")
	fmt.Println("  whoo report --from 202601 --to 202612")
	fmt.Println("  whoo inout --from 20260801 --to 20260813")
	fmt.Println("  whoo sections          섹션 목록")
	fmt.Println("  whoo budget get expenses")
}

// ShowVersion은 버전만 출력
func ShowVersion() {
	fmt.Printf("whoo %s\n", Version)
}
