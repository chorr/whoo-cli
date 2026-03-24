// cmd/status.go
// status CLI 커맨드 — 인증/설정 상태 요약

package cmd

import (
	"fmt"

	"whooing-cli/config"
)

// RunStatus는 현재 인증 및 설정 상태를 표시
// 보안성 높은 값(token, secret)은 존재 여부만 표시
func RunStatus(cfg *config.Config) {
	fmt.Printf("후잉 CLI v%s\n\n", Version)

	// 인증 상태
	if cfg.IsAuthenticated() {
		fmt.Println("  인증:   완료")
	} else {
		fmt.Println("  인증:   필요 (whooing auth)")
	}

	// 섹션
	if cfg.SectionID != "" {
		fmt.Printf("  섹션:   %s\n", cfg.SectionID)
	} else {
		fmt.Println("  섹션:   미선택 (whooing sections set <id>)")
	}
}
