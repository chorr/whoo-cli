// cmd/sortlist.go
// 순서 변경(정렬) 모드 공용 컴포넌트
// 화면에 보이는 목록을 사용자가 재배열한 뒤 전체 ID 순서를 서버에 저장한다

package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// sortEntry는 정렬 대상 한 줄 (ID + 표시 라벨)
type sortEntry struct {
	id    string
	label string
}

// sortState는 정렬 모드의 로컬 상태
type sortState struct {
	entries []sortEntry
	cursor  int
	moved   bool // 하나라도 이동했는지 (저장 안내용)
}

// newSortState는 정렬 상태를 초기화
func newSortState(entries []sortEntry, cursor int) sortState {
	if cursor < 0 || cursor >= len(entries) {
		cursor = 0
	}
	return sortState{entries: entries, cursor: cursor}
}

// ids는 현재 순서의 ID 슬라이스 반환
func (s *sortState) ids() []string {
	out := make([]string, len(s.entries))
	for i, e := range s.entries {
		out[i] = e.id
	}
	return out
}

// sortKeyResult는 정렬 모드 키 처리 결과
type sortKeyResult int

const (
	sortKeyHandled sortKeyResult = iota // 내부에서 처리됨 (계속 정렬 모드)
	sortKeySave                         // enter: 현재 순서 저장 요청
	sortKeyCancel                       // esc: 취소
)

// handleKey는 정렬 모드의 키 입력 처리
// ↑/↓/j/k = 커서 이동, K/J(Shift) = 항목 이동, Enter = 저장, Esc = 취소
func (s *sortState) handleKey(msg tea.KeyMsg) sortKeyResult {
	switch msg.String() {
	case "esc":
		return sortKeyCancel
	case "enter":
		return sortKeySave
	case "up", "k":
		if s.cursor > 0 {
			s.cursor--
		}
	case "down", "j":
		if s.cursor < len(s.entries)-1 {
			s.cursor++
		}
	case "K", "shift+up":
		if s.cursor > 0 {
			s.entries[s.cursor-1], s.entries[s.cursor] = s.entries[s.cursor], s.entries[s.cursor-1]
			s.cursor--
			s.moved = true
		}
	case "J", "shift+down":
		if s.cursor < len(s.entries)-1 {
			s.entries[s.cursor+1], s.entries[s.cursor] = s.entries[s.cursor], s.entries[s.cursor+1]
			s.cursor++
			s.moved = true
		}
	}
	return sortKeyHandled
}

// render는 정렬 목록과 도움말을 렌더링
func (s *sortState) render(b *strings.Builder) {
	for i, e := range s.entries {
		line := fmt.Sprintf("  %s", e.label)
		if i == s.cursor {
			b.WriteString(selectedStyle.Render("> "+line[2:]) + "\n")
		} else {
			b.WriteString(line + "\n")
		}
	}
	b.WriteString("\n" + helpStyle.Render("[↑/↓/j/k] 이동  [Shift+↑/↓ 또는 K/J] 순서 변경  [Enter] 저장  [Esc] 취소"))
}
