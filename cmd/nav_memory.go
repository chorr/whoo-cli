// cmd/nav_memory.go
// 화면 전환 시 선택 커서 복원용 가벼운 인덱스 저장

package cmd

import "github.com/charmbracelet/bubbles/list"

// navMemory는 메인/1뎁스 선택 위치를 기억한다.
// 화면 서브모델은 transitionTo 때 소멸되므로 appModel에 보관한다.
type navMemory struct {
	menuIndex       int
	sectionHubIndex int
	frequentSlot    int
	monthlySlot     int
	cardTab         int
	budgetType      int // 0=expenses, 1=income
	accountType     int // AccountTypes 목록 커서
	flowType        int // 흐름 분석 타입 커서
}

// clampIndex는 0..n-1 범위로 인덱스를 보정한다
func clampIndex(i, n int) int {
	if n <= 0 {
		return 0
	}
	if i < 0 {
		return 0
	}
	if i >= n {
		return n - 1
	}
	return i
}

// selectListIndex는 bubbles/list 커서를 안전히 복원한다
func selectListIndex(l *list.Model, idx int) {
	n := len(l.Items())
	if n == 0 {
		return
	}
	l.Select(clampIndex(idx, n))
}
