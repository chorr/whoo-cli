// cmd/help.go
// help/--help/-h 예약어 처리. 이 경로에서는 인증·API·TUI를 실행하지 않는다.

package cmd

import (
	"os"

	"github.com/mattn/go-isatty"
)

// isHelpArg는 단일 인자가 도움말 예약어인지 판별
func isHelpArg(s string) bool {
	return s == "help" || s == "--help" || s == "-h"
}

// wantsHelp는 인자 어디에든 help/--help/-h가 있으면 true
func wantsHelp(args []string) bool {
	for _, a := range args {
		if isHelpArg(a) {
			return true
		}
	}
	return false
}

// HasInteractiveTTY는 TUI를 열 수 있는 터미널이 있는지 확인
// bubbletea는 /dev/tty를 열므로, 에이전트 셸처럼 /dev/tty가 없으면 false
func HasInteractiveTTY() bool {
	f, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err == nil {
		_ = f.Close()
		return true
	}
	fd := os.Stdin.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}
