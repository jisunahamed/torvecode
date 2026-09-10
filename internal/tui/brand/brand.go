package brand

import (
	"github.com/charmbracelet/lipgloss"
	"os"
	"strings"
)

// Ink follows the terminal appearance and respects uncoloured terminals.
func Ink(dark, light string) lipgloss.TerminalColor {
	if os.Getenv("NO_COLOR") != "" {
		return lipgloss.NoColor{}
	}
	return lipgloss.AdaptiveColor{Dark: dark, Light: light}
}

// Signature translates Torve's cube robot into portable terminal characters.
func Signature(width int) string {
	white := lipgloss.NewStyle().Foreground(Ink("#F7FAFC", "#172421")).Bold(true)
	lime := lipgloss.NewStyle().Foreground(Ink("#C7FF24", "#527000"))
	mint := lipgloss.NewStyle().Foreground(Ink("#40F887", "#087F48"))
	cyan := lipgloss.NewStyle().Foreground(Ink("#00DCC4", "#007B70"))
	title := white.Render("TORVECODE")
	if width < 48 {
		return title
	}
	robot := strings.Join([]string{
		lime.Render("     ▄██████▄"),
		mint.Render("  ▄██▀      ▀██▄"),
		cyan.Render("  ██ ") + white.Render(" ▄▄▄▄▄▄▄ ") + mint.Render("█"),
		cyan.Render("  ██ ") + white.Render(" █ ● ● █ ") + mint.Render("█"),
		cyan.Render("  ▀█ ") + white.Render(" ▀▀█ █▀▀ ") + mint.Render("█"),
		cyan.Render("    ▀▄ ") + white.Render(" █ █ "),
		cyan.Render("      ▀ ") + white.Render("▀▀▀"),
	}, "\n")
	copy := lipgloss.JoinVertical(lipgloss.Left, title, "", white.Bold(false).Render("Make something yours."), cyan.Render("Torve AI  /  coding workspace"))
	return lipgloss.JoinHorizontal(lipgloss.Center, robot, "    ", copy)
}
