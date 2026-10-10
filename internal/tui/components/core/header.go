package core

import (
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jisunahamed/torvecode/internal/app"
	"github.com/jisunahamed/torvecode/internal/config"
	"github.com/jisunahamed/torvecode/internal/llm/models"
	"github.com/jisunahamed/torvecode/internal/tui/theme"
	"github.com/jisunahamed/torvecode/internal/version"
)

type HeaderCmp struct {
	width int
	app   *app.App
}

func NewHeaderCmp(app *app.App) HeaderCmp {
	return HeaderCmp{app: app}
}

func (h *HeaderCmp) SetSize(width, _ int) tea.Cmd {
	h.width = width
	return nil
}

func (h HeaderCmp) Height() int { return 2 }

func truncateMiddle(s string, maxLen int) string {
	if maxLen <= 0 || lipgloss.Width(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		// Return prefix truncated by display width
		runes := []rune(s)
		w := 0
		for i, r := range runes {
			rw := lipgloss.Width(string(r))
			if w+rw > maxLen {
				return string(runes[:i])
			}
			w += rw
		}
		return s
	}
	// Width-aware truncation: split budget between prefix and suffix around "…"
	ellipsisW := lipgloss.Width("…")
	budget := maxLen - ellipsisW
	if budget <= 0 {
		return "…"
	}
	half := budget / 2
	other := budget - half

	runes := []rune(s)
	// Accumulate prefix by width
	prefixW := 0
	prefixEnd := 0
	for i, r := range runes {
		rw := lipgloss.Width(string(r))
		if prefixW+rw > half {
			break
		}
		prefixW += rw
		prefixEnd = i + 1
	}
	// Accumulate suffix by width from the end
	suffixW := 0
	suffixStart := len(runes)
	for i := len(runes) - 1; i >= 0; i-- {
		rw := lipgloss.Width(string(runes[i]))
		if suffixW+rw > other {
			break
		}
		suffixW += rw
		suffixStart = i
	}
	// If prefix+suffix covers the string, no truncation needed
	if prefixEnd >= suffixStart {
		return s
	}
	return string(runes[:prefixEnd]) + "…" + string(runes[suffixStart:])
}

func (h HeaderCmp) View() string {
	t := theme.CurrentTheme()
	if t == nil {
		return ""
	}
	w := h.width
	if w < 10 {
		return ""
	}

	// Left: ⬡ TorveCode version
	hexIcon := lipgloss.NewStyle().Foreground(t.Success()).Render("⬡")
	nameStyle := lipgloss.NewStyle().Foreground(t.Text()).Bold(true)
	verStyle := lipgloss.NewStyle().Foreground(t.TextMuted())
	left := hexIcon + " " + nameStyle.Render("TorveCode") + "  " + verStyle.Render(version.Version)

	// Center: workspace | model
	cwd := config.WorkingDirectory()
	// Show basename-heavy path; truncate middle for long paths
	displayCwd := cwd
	if lipgloss.Width(displayCwd) > 36 {
		base := filepath.Base(cwd)
		dir := filepath.Dir(cwd)
		if len(base)+4 < 36 {
			displayCwd = truncateMiddle(dir, 36-len(base)-1) + "/" + base
		} else {
			displayCwd = truncateMiddle(cwd, 36)
		}
	}
	folderStyle := lipgloss.NewStyle().Foreground(t.TextMuted())
	folderIcon := folderStyle.Render("◇")
	center := folderIcon + " " + folderStyle.Render(displayCwd)

	modelName := ""
	if cfg := config.Get(); cfg != nil {
		if ag, ok := cfg.Agents[config.AgentCoder]; ok {
			if m, ok := models.SupportedModels[ag.Model]; ok && m.Name != "" {
				modelName = m.Name
			} else {
				modelName = string(ag.Model)
			}
		}
	}
	if modelName != "" {
		cubeStyle := lipgloss.NewStyle().Foreground(t.TextMuted())
		center += "  " + cubeStyle.Render("│") + "  " + lipgloss.NewStyle().Foreground(t.TextMuted()).Render("⬢") + " " + lipgloss.NewStyle().Foreground(t.TextMuted()).Render(modelName)
	}

	// Right: connection + shortcut hint
	connectedStyle := lipgloss.NewStyle().Foreground(t.Success())
	dot := connectedStyle.Render("●")
	rightConnected := dot + " " + connectedStyle.Render("Connected")
	hintStyle := lipgloss.NewStyle().Foreground(t.TextMuted())
	rightHint := hintStyle.Render("Ctrl+B  Toggle Sidebar")
	sepStyle := lipgloss.NewStyle().Foreground(t.BorderNormal())
	sep := sepStyle.Render(" │ ")

	// Measure and truncate center if needed
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(rightConnected) + lipgloss.Width(sep) + lipgloss.Width(rightHint)
	centerW := lipgloss.Width(center)
	availableForCenter := w - leftW - rightW - 6 // padding
	if availableForCenter < 10 {
		// Hide hint when narrow
		rightHint = ""
		rightW = lipgloss.Width(rightConnected)
		availableForCenter = w - leftW - rightW - 6
	}
	if availableForCenter < 8 {
		center = ""
		centerW = 0
	} else if centerW > availableForCenter {
		center = truncateMiddle(center, availableForCenter)
		centerW = lipgloss.Width(center)
	}

	// Build row with spacing
	gap := w - leftW - centerW - rightW - 4
	if gap < 2 {
		gap = 2
	}
	// Distribute gap: left-center and center-right
	gap1 := gap / 2
	gap2 := gap - gap1

	row := left + strings.Repeat(" ", gap1) + center + strings.Repeat(" ", gap2) + rightConnected
	if rightHint != "" {
		row += sep + rightHint
	}
	row = lipgloss.NewStyle().
		Width(w).
		Padding(0, 1).
		Background(t.Background()).
		Render(row)

	// Thin bottom separator — same inner width as row (Padding 0,1) for alignment
	sep2 := lipgloss.NewStyle().
		Width(w).
		Padding(0, 1).
		Foreground(t.BorderDim()).
		Render(strings.Repeat("─", w-2))

	return lipgloss.JoinVertical(lipgloss.Left, row, sep2)
}

// Height returns the rendered height (used by tui.go to reserve space).
func HeaderHeight() int { return 2 }
