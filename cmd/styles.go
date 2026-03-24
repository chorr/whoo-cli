// cmd/styles.go
// TUI 공통 스타일 정의

package cmd

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	headerStyle   = lipgloss.NewStyle().Bold(true)
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	loadingStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	confirmStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)
)
