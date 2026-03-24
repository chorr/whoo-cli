// cmd/sections.go
// sections CLI 커맨드 — 섹션 관리 (서브커맨드 구조)

package cmd

import (
	"fmt"
	"os"

	"whooing-cli/config"
)

// RunSections는 sections CLI 커맨드 실행
// 서브커맨드에 따라 분기 처리
func RunSections(cfg *config.Config, args []string) {
	RequireAuth(cfg)

	// 서브커맨드 없으면 전체 목록
	if len(args) == 0 {
		runSectionsList(cfg)
		return
	}

	switch args[0] {
	case "default":
		runSectionsDefault(cfg)
	case "set":
		if len(args) < 2 {
			PrintError("섹션 ID가 필요합니다")
			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "  사용법: whooing sections set <section_id>")
			os.Exit(1)
		}
		runSectionsSet(cfg, args[1])
	case "help", "--help", "-h":
		showSectionsHelp()
	default:
		// section_id로 간주
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

// showSectionsHelp는 sections 서브커맨드 도움말 출력
func showSectionsHelp() {
	fmt.Println("사용법: whooing sections [command]")
	fmt.Println()
	fmt.Println("커맨드:")
	fmt.Println("  (없음)           전체 섹션 목록")
	fmt.Println("  default          기본 섹션 조회")
	fmt.Println("  <section_id>     특정 섹션 조회")
	fmt.Println("  set <section_id> 현재 섹션 설정")
	fmt.Println("  help             도움말 표시")
	fmt.Println()
	fmt.Println("예시:")
	fmt.Println("  whooing sections             전체 섹션 목록")
	fmt.Println("  whooing sections default      기본 섹션 조회")
	fmt.Println("  whooing sections s123         특정 섹션 조회")
	fmt.Println("  whooing sections set s123     섹션을 s123으로 설정")
}
