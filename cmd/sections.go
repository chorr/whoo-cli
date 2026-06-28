// cmd/sections.go
// sections CLI 커맨드 — 섹션 관리 (서브커맨드 구조)

package cmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"whoo-cli/api"
	"whoo-cli/config"
)

// RunSections는 sections CLI 커맨드 실행
func RunSections(cfg *config.Config, args []string) {
	RequireAuth(cfg)

	if len(args) == 0 {
		runSectionsList(cfg)
		return
	}

	switch args[0] {
	case "add":
		runSectionsAdd(cfg, args[1:])
	case "edit":
		runSectionsEdit(cfg, args[1:])
	case "delete", "del", "rm":
		runSectionsDelete(cfg, args[1:])
	case "sort":
		runSectionsSort(cfg, args[1:])
	case "default":
		runSectionsDefault(cfg)
	case "set":
		if len(args) < 2 {
			PrintError("섹션 ID가 필요합니다")
			fmt.Fprintln(os.Stderr, "  사용법: whoo sections set <section_id>")
			os.Exit(1)
		}
		runSectionsSet(cfg, args[1])
	case "help", "--help", "-h":
		showSectionsHelp()
	default:
		runSectionsGet(cfg, args[0])
	}
}

// runSectionsList는 전체 섹션 목록 조회
func runSectionsList(cfg *config.Config) {
	client := NewClient(cfg)
	data, err := client.GetSectionsAll()
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runSectionsDefault는 기본 섹션 조회
func runSectionsDefault(cfg *config.Config) {
	client := NewClient(cfg)
	data, err := client.GetSectionDefault()
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runSectionsGet은 특정 섹션 조회
func runSectionsGet(cfg *config.Config, sectionID string) {
	client := NewClient(cfg)
	data, err := client.GetSection(sectionID)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runSectionsSet은 현재 섹션을 설정
func runSectionsSet(cfg *config.Config, sectionID string) {
	cfg.SectionID = sectionID
	if err := cfg.Save(); err != nil {
		PrintError("섹션 설정 저장 실패: %v", err)
		os.Exit(1)
	}
	fmt.Printf("섹션이 %s로 설정되었습니다\n", sectionID)
}

// runSectionsAdd는 섹션 신규 생성
func runSectionsAdd(cfg *config.Config, args []string) {
	fs := flag.NewFlagSet("sections add", flag.ExitOnError)
	title := fs.String("title", "", "섹션 제목 (필수)")
	currency := fs.String("currency", "KRW", "통화 단위")
	memo := fs.String("memo", "", "섹션 설명")
	skinID := fs.Int("skin-id", 0, "스킨 번호")
	decimalPlaces := fs.Int("decimal-places", 0, "소수점 자릿수 (0~3)")
	dateFormat := fs.String("date-format", "", "날짜 표시 방식 (예: YMD)")
	startYear := fs.Int("start-year", 0, "항목 시작연도")
	templateID := fs.Int("template-id", 0, "초기 항목 템플릿 ID")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *title == "" {
		PrintError("--title 은 필수입니다")
		os.Exit(1)
	}

	client := NewClient(cfg)
	data, err := client.CreateSection(api.SectionCreateParams{
		Title:         *title,
		Currency:      *currency,
		Memo:          *memo,
		SkinID:        *skinID,
		DecimalPlaces: *decimalPlaces,
		DateFormat:    *dateFormat,
		StartYear:     *startYear,
		TemplateID:    *templateID,
	})
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runSectionsEdit는 섹션 정보 수정
func runSectionsEdit(cfg *config.Config, args []string) {
	if len(args) == 0 {
		PrintError("section_id가 필요합니다")
		fmt.Fprintln(os.Stderr, "  사용법: whoo sections edit <section_id> [--title ...] [--ui-key key=value]")
		os.Exit(1)
	}
	sectionID := args[0]

	fs := flag.NewFlagSet("sections edit", flag.ExitOnError)
	title := fs.String("title", "", "섹션 제목")
	currency := fs.String("currency", "", "통화 단위")
	memo := fs.String("memo", "", "섹션 설명")
	skinID := fs.Int("skin-id", -1, "스킨 번호")
	decimalPlaces := fs.Int("decimal-places", -1, "소수점 자릿수")
	dateFormat := fs.String("date-format", "", "날짜 표시 방식")
	var uiKeys multiFlag
	fs.Var(&uiKeys, "ui-key", "UI 설정 (key=value 형식, 여러 번 사용 가능)")
	if err := fs.Parse(args[1:]); err != nil {
		os.Exit(1)
	}

	// 현재 섹션 정보 조회해서 기본값 설정
	client := NewClient(cfg)
	existing, err := client.GetSection(sectionID)
	if err != nil {
		PrintError("섹션 조회 실패: %v", err)
		os.Exit(1)
	}

	// 기존 값 파싱
	var sec struct {
		Results struct {
			Title         string `json:"title"`
			Currency      string `json:"currency"`
			Memo          string `json:"memo"`
			SkinID        int    `json:"skin_id"`
			DecimalPlaces int    `json:"decimal_places"`
			DateFormat    string `json:"date_format"`
		} `json:"results"`
	}
	if err := parseJSONResponse(existing, &sec); err == nil {
		if *title == "" {
			*title = sec.Results.Title
		}
		if *currency == "" {
			*currency = sec.Results.Currency
		}
		if *memo == "" {
			*memo = sec.Results.Memo
		}
		if *skinID < 0 {
			*skinID = sec.Results.SkinID
		}
		if *decimalPlaces < 0 {
			*decimalPlaces = sec.Results.DecimalPlaces
		}
		if *dateFormat == "" {
			*dateFormat = sec.Results.DateFormat
		}
	}

	// UI 설정 파싱: "key=value" 형식
	uiSettings := make(map[string]string)
	for _, kv := range uiKeys {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			PrintError("--ui-key 형식 오류: %q (key=value 형식 필요)", kv)
			os.Exit(1)
		}
		uiSettings[parts[0]] = parts[1]
	}

	data, err := client.UpdateSection(sectionID, api.SectionUpdateParams{
		Title:         *title,
		Currency:      *currency,
		Memo:          *memo,
		SkinID:        *skinID,
		DecimalPlaces: *decimalPlaces,
		DateFormat:    *dateFormat,
		UISettings:    uiSettings,
	})
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runSectionsDelete는 섹션 삭제
func runSectionsDelete(cfg *config.Config, args []string) {
	if len(args) == 0 {
		PrintError("section_id가 필요합니다 (콤마로 복수 지정 가능)")
		os.Exit(1)
	}
	// 첫 번째 인자를 콤마로 분리 (또는 여러 인자 허용)
	ids := splitCommaOrArgs(args)

	client := NewClient(cfg)
	data, err := client.DeleteSections(ids)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// runSectionsSort는 섹션 순서 변경
func runSectionsSort(cfg *config.Config, args []string) {
	if len(args) == 0 {
		PrintError("section_id 목록이 필요합니다 (콤마로 구분)")
		fmt.Fprintln(os.Stderr, "  사용법: whoo sections sort <s1,s2,s3>")
		os.Exit(1)
	}
	ids := splitCommaOrArgs(args)

	client := NewClient(cfg)
	data, err := client.SortSections(ids)
	if err != nil {
		PrintError("%v", err)
		os.Exit(1)
	}
	printJSON(data)
}

// showSectionsHelp는 sections 서브커맨드 도움말 출력
func showSectionsHelp() {
	fmt.Println("사용법: whoo sections [command]")
	fmt.Println()
	fmt.Println("커맨드:")
	fmt.Println("  (없음)                           전체 섹션 목록")
	fmt.Println("  add --title <제목> [옵션...]      섹션 생성")
	fmt.Println("  edit <section_id> [옵션...]      섹션 수정")
	fmt.Println("  delete <section_id[,id,...]>     섹션 삭제")
	fmt.Println("  sort <s1,s2,s3>                  섹션 순서 변경")
	fmt.Println("  default                          기본 섹션 조회")
	fmt.Println("  set <section_id>                 현재 섹션 설정")
	fmt.Println("  <section_id>                     특정 섹션 조회")
	fmt.Println("  help                             도움말")
	fmt.Println()
	fmt.Println("sections add 옵션:")
	fmt.Println("  --title         섹션 제목 (필수)")
	fmt.Println("  --currency      통화 (기본: KRW)")
	fmt.Println("  --memo          설명")
	fmt.Println("  --skin-id       스킨 번호")
	fmt.Println("  --decimal-places  소수점 자릿수")
	fmt.Println("  --date-format   날짜 형식 (예: YMD)")
	fmt.Println("  --start-year    항목 시작연도")
	fmt.Println("  --template-id   초기 항목 템플릿")
	fmt.Println()
	fmt.Println("sections edit 옵션:")
	fmt.Println("  --title / --currency / --memo / --skin-id / --decimal-places / --date-format")
	fmt.Println("  --ui-key key=value  UI 설정 (여러 번 사용 가능)")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whoo sections add --title \"주머니\" --currency KRW")
	fmt.Println("  whoo sections edit s123 --title \"새 이름\" --ui-key budgetLong=y")
	fmt.Println("  whoo sections delete s129")
	fmt.Println("  whoo sections delete s129,s118")
	fmt.Println("  whoo sections sort s99,s72,s78")
}

// ─── 헬퍼 ─────────────────────────────────────────────────────

// multiFlag는 flag.Value 인터페이스를 구현하는 문자열 슬라이스 (--flag 반복 허용)
type multiFlag []string

func (f *multiFlag) String() string { return strings.Join(*f, ", ") }
func (f *multiFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

// splitCommaOrArgs는 "s1,s2" 또는 ["s1", "s2"] 형태 모두 처리
func splitCommaOrArgs(args []string) []string {
	if len(args) == 1 {
		return strings.Split(args[0], ",")
	}
	return args
}
