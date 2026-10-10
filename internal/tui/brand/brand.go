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
// Palette: lime #A3E635 on dark #0B0D0F, matches reference spec.
func Signature(width int) string {
	lime := lipgloss.NewStyle().Foreground(Ink("#A3E635", "#65A30D")).Bold(true)
	limeDim := lipgloss.NewStyle().Foreground(Ink("#86C530", "#65A30D"))
	txt := lipgloss.NewStyle().Foreground(Ink("#E8EAED", "#0F172A"))
	muted := lipgloss.NewStyle().Foreground(Ink("#77818C", "#64748B"))
	title := lime.Render("TORVECODE")
	if width < 48 {
		return title
	}
	robot := strings.Join([]string{
		lime.Render("     ▄██████▄"),
		limeDim.Render("  ▄██▀      ▀██▄"),
		txt.Render("  ██ ") + lime.Render(" ▄▄▄▄▄▄▄ ") + muted.Render("█"),
		txt.Render("  ██ ") + lime.Render(" █ ● ● █ ") + muted.Render("█"),
		txt.Render("  ▀█ ") + lime.Render(" ▀▀█ █▀▀ ") + muted.Render("█"),
		txt.Render("    ▀▄ ") + lime.Render(" █ █ "),
		txt.Render("      ▀ ") + lime.Render("▀▀▀"),
	}, "\n")
	copy := lipgloss.JoinVertical(lipgloss.Left, title, "", txt.Bold(false).Render("Make something yours."), muted.Render("Torve AI  /  coding workspace"))
	return lipgloss.JoinHorizontal(lipgloss.Center, robot, "    ", copy)
}
