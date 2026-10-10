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
// Palette: amber/graphite SaaS look.
func Signature(width int) string {
	amber := lipgloss.NewStyle().Foreground(Ink("#F59E0B", "#D97706")).Bold(true)
	amberDim := lipgloss.NewStyle().Foreground(Ink("#FBBF24", "#F59E0B"))
	stone := lipgloss.NewStyle().Foreground(Ink("#E7E5E4", "#1C1917"))
	stoneMuted := lipgloss.NewStyle().Foreground(Ink("#A8A29E", "#78716C"))
	title := amber.Render("TORVECODE")
	if width < 48 {
		return title
	}
	robot := strings.Join([]string{
		amber.Render("     ▄██████▄"),
		amberDim.Render("  ▄██▀      ▀██▄"),
		stone.Render("  ██ ") + amber.Render(" ▄▄▄▄▄▄▄ ") + stoneMuted.Render("█"),
		stone.Render("  ██ ") + amber.Render(" █ ● ● █ ") + stoneMuted.Render("█"),
		stone.Render("  ▀█ ") + amber.Render(" ▀▀█ █▀▀ ") + stoneMuted.Render("█"),
		stone.Render("    ▀▄ ") + amber.Render(" █ █ "),
		stone.Render("      ▀ ") + amber.Render("▀▀▀"),
	}, "\n")
	copy := lipgloss.JoinVertical(lipgloss.Left, title, "", stone.Bold(false).Render("Make something yours."), stoneMuted.Render("Torve AI  /  coding workspace"))
	return lipgloss.JoinHorizontal(lipgloss.Center, robot, "    ", copy)
}
