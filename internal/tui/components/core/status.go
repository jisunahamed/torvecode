package core

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jisunahamed/torvecode/internal/config"
	"github.com/jisunahamed/torvecode/internal/llm/models"
	"github.com/jisunahamed/torvecode/internal/lsp"
	"github.com/jisunahamed/torvecode/internal/pubsub"
	"github.com/jisunahamed/torvecode/internal/session"
	"github.com/jisunahamed/torvecode/internal/tui/components/chat"
	"github.com/jisunahamed/torvecode/internal/tui/theme"
	"github.com/jisunahamed/torvecode/internal/tui/util"
)

type StatusCmp interface {
	tea.Model
}

type statusCmp struct {
	info       util.InfoMsg
	width      int
	messageTTL time.Duration
	lspClients map[string]*lsp.Client
	session    session.Session
}

func (m statusCmp) clearMessageCmd(ttl time.Duration) tea.Cmd {
	return tea.Tick(ttl, func(time.Time) tea.Msg {
		return util.ClearStatusMsg{}
	})
}

func (m statusCmp) Init() tea.Cmd { return nil }

func (m statusCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case chat.SessionSelectedMsg:
		m.session = msg
	case chat.SessionClearedMsg:
		m.session = session.Session{}
	case pubsub.Event[session.Session]:
		if msg.Type == pubsub.UpdatedEvent && msg.Payload.ID == m.session.ID {
			m.session = msg.Payload
		}
	case util.InfoMsg:
		m.info = msg
		ttl := msg.TTL
		if ttl == 0 {
			ttl = m.messageTTL
		}
		return m, m.clearMessageCmd(ttl)
	case util.ClearStatusMsg:
		m.info = util.InfoMsg{}
	}
	return m, nil
}

func formatTokenCount(tokens int64) string {
	var s string
	switch {
	case tokens >= 1_000_000:
		s = fmt.Sprintf("%.1fM", float64(tokens)/1_000_000)
	case tokens >= 1_000:
		s = fmt.Sprintf("%.1fK", float64(tokens)/1_000)
	default:
		s = fmt.Sprintf("%d", tokens)
	}
	s = strings.ReplaceAll(s, ".0K", "K")
	s = strings.ReplaceAll(s, ".0M", "M")
	return s
}

func formatTokensAndCost(inputTokens, outputTokens, contextWindow int64, cost float64) string {
	total := inputTokens + outputTokens
	pct := 0
	if contextWindow > 0 {
		pct = int(float64(total) / float64(contextWindow) * 100)
	}
	s := fmt.Sprintf("IN %s  OUT %s  %s/%s (%d%%)", formatTokenCount(inputTokens), formatTokenCount(outputTokens), formatTokenCount(total), formatTokenCount(contextWindow), pct)
	if cost > 0 {
		s += fmt.Sprintf("  $%.4f", cost)
	}
	return s
}

func (m statusCmp) View() string {
	t := theme.CurrentTheme()
	if t == nil {
		return ""
	}
	w := m.width
	if w < 10 {
		return ""
	}

	muted := lipgloss.NewStyle().Foreground(t.TextMuted())
	dim := lipgloss.NewStyle().Foreground(t.BorderDim())
	green := lipgloss.NewStyle().Foreground(t.Success())

	// Left: IN 0 · OUT 0 · 0% context (0/128K) or (0/128K) + cost
	var left string
	if m.session.ID != "" {
		cfg := config.Get()
		var window int64
		if cfg != nil {
			if ag, ok := cfg.Agents[config.AgentCoder]; ok {
				if mod, ok := models.SupportedModels[ag.Model]; ok {
					window = mod.EffectiveContextWindow()
				}
			}
		}
		in := m.session.PromptTokens
		out := m.session.CompletionTokens
		total := in + out
		pct := 0
		if window > 0 {
			pct = int(float64(total) / float64(window) * 100)
		}
		left = fmt.Sprintf("IN %s · OUT %s", formatTokenCount(in), formatTokenCount(out))
		if window > 0 {
			left += fmt.Sprintf(" · %d%% context (%s/%s)", pct, formatTokenCount(total), formatTokenCount(window))
		} else {
			left += fmt.Sprintf(" · %s tokens", formatTokenCount(total))
		}
		if m.session.Cost > 0 {
			left += fmt.Sprintf("  $%.4f", m.session.Cost)
		}
		// Hide context on very narrow
		if w < 60 {
			left = fmt.Sprintf("IN %s · OUT %s", formatTokenCount(in), formatTokenCount(out))
		}
		if w < 45 {
			left = ""
		}
		left = muted.Render(left)
	} else {
		left = muted.Render("IN 0 · OUT 0 · 0% context (0/128K)")
		if w < 60 {
			left = muted.Render("IN 0 · OUT 0")
		}
		if w < 45 {
			left = ""
		}
	}

	// Transient info replaces left when present
	if m.info.Msg != "" {
		infoStyle := lipgloss.NewStyle().Foreground(t.Background())
		switch m.info.Type {
		case util.InfoTypeInfo:
			infoStyle = infoStyle.Background(t.Info())
		case util.InfoTypeWarn:
			infoStyle = infoStyle.Background(t.Warning())
		case util.InfoTypeError:
			infoStyle = infoStyle.Background(t.Error())
		}
		msg := m.info.Msg
		avail := w - 20
		if avail > 10 && len(msg) > avail {
			msg = msg[:avail-3] + "..."
		}
		left = infoStyle.Padding(0, 1).Render(msg)
	}

	// Right: Ctrl+? Help | ● LSP Ready
	help := muted.Render("Ctrl+? Help")
	lspReady := ""
	// Only show LSP Ready when we actually have LSP clients ready; otherwise keep help only
	hasReady := false
	for _, c := range m.lspClients {
		if c != nil && c.GetServerState() == lsp.StateReady {
			hasReady = true
			break
		}
	}
	dot := green.Render("●")
	_ = dot
	if hasReady {
		lspReady = dot + " " + green.Render("LSP Ready")
	} else if len(m.lspClients) > 0 {
		lspReady = lipgloss.NewStyle().Foreground(t.Warning()).Render("○ LSP")
	} else {
		// No LSP configured — only show help, not fake LSP Ready. Match spec only when truly ready.
		lspReady = ""
	}
	right := help
	if lspReady != "" {
		right = help + dim.Render("  │  ") + lspReady
	}
	if w < 40 {
		right = help
	}

	// Single line with top separator — same inner width as row (Padding 0,1) for alignment with header
	sep := lipgloss.NewStyle().Width(w).Padding(0, 1).Foreground(t.BorderDim()).Render(strings.Repeat("─", w-2))
	gap := w - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	row := lipgloss.NewStyle().Width(w).Padding(0, 1).Background(t.Background()).Render(left + strings.Repeat(" ", gap) + right)
	return lipgloss.JoinVertical(lipgloss.Left, sep, row)
}

func (m statusCmp) model() string {
	t := theme.CurrentTheme()
	if t == nil {
		return ""
	}
	cfg := config.Get()
	if cfg == nil {
		return ""
	}
	coder, ok := cfg.Agents[config.AgentCoder]
	if !ok {
		return ""
	}
	model := models.SupportedModels[coder.Model]
	_ = model
	return ""
}

func (m statusCmp) projectDiagnostics() string { return "" }

func (m statusCmp) availableFooterMsgWidth(_, _ string) int { return m.width }

func NewStatusCmp(lspClients map[string]*lsp.Client) StatusCmp {
	return &statusCmp{
		messageTTL: 10 * time.Second,
		lspClients: lspClients,
	}
}
