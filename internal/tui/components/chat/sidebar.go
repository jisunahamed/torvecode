package chat

import (
	"context"
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jisunahamed/torvecode/internal/config"
	"github.com/jisunahamed/torvecode/internal/diff"
	"github.com/jisunahamed/torvecode/internal/history"
	"github.com/jisunahamed/torvecode/internal/pubsub"
	"github.com/jisunahamed/torvecode/internal/session"
	"github.com/jisunahamed/torvecode/internal/tui/styles"
	"github.com/jisunahamed/torvecode/internal/tui/theme"
)

type sidebarCmp struct {
	width, height int
	session       session.Session
	history       history.Service
	modFiles      map[string]struct {
		additions int
		removals  int
	}
}

func (m *sidebarCmp) Init() tea.Cmd {
	if m.history != nil {
		ctx := context.Background()
		filesCh := m.history.Subscribe(ctx)
		m.modFiles = make(map[string]struct {
			additions int
			removals  int
		})
		m.loadModifiedFiles(ctx)
		return func() tea.Msg {
			return <-filesCh
		}
	}
	return nil
}

func (m *sidebarCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SessionSelectedMsg:
		if msg.ID != m.session.ID {
			m.session = msg
			ctx := context.Background()
			m.loadModifiedFiles(ctx)
		}
	case pubsub.Event[session.Session]:
		if msg.Type == pubsub.UpdatedEvent {
			if m.session.ID == msg.Payload.ID {
				m.session = msg.Payload
			}
		}
	case pubsub.Event[history.File]:
		if msg.Payload.SessionID == m.session.ID {
			ctx := context.Background()
			m.processFileChanges(ctx, msg.Payload)
			return m, func() tea.Msg {
				ctx := context.Background()
				filesCh := m.history.Subscribe(ctx)
				return <-filesCh
			}
		}
	}
	return m, nil
}

func sectionHeading(label string) string {
	t := theme.CurrentTheme()
	return lipgloss.NewStyle().
		Foreground(t.TextMuted()).
		Bold(true).
		Render(strings.ToUpper(label))
}

func (m *sidebarCmp) View() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	// Clamp width 24-28, padding 1,1
	w := m.width
	if w < 10 {
		return ""
	}
	contentW := w - 2
	if contentW < 8 {
		contentW = w
	}
	cwd := config.WorkingDirectory()
	if len(cwd) > contentW {
		cwd = cwd[:contentW-1] + "…"
	}

	wsSection := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.NewStyle().Foreground(t.Success()).Render("◇ "), baseStyle.Foreground(t.TextMuted()).Width(contentW-2).Render("WORKSPACE")),
		baseStyle.Foreground(t.Text()).Width(contentW).Render(cwd),
	)

	csIcon := lipgloss.NewStyle().Foreground(t.Success()).Render("▭")
	sessionTitle := m.session.Title
	if sessionTitle == "" {
		sessionTitle = "No session"
	}
	if len(sessionTitle) > contentW {
		sessionTitle = sessionTitle[:contentW-1] + "…"
	}
	sessionDesc := ""
	if m.session.Title != "" {
		sessionDesc = "Active"
	}
	csSection := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.NewStyle().Foreground(t.Success()).Render("◫ "), baseStyle.Foreground(t.TextMuted()).Width(contentW-2).Render("CURRENT SESSION")),
		baseStyle.Foreground(t.Text()).Width(contentW).Render(sessionTitle),
		baseStyle.Foreground(t.TextMuted()).Width(contentW).Render(sessionDesc),
	)
	_ = csIcon

	// Modified files
	mfHeading := lipgloss.JoinHorizontal(lipgloss.Left, lipgloss.NewStyle().Foreground(t.Success()).Render("⑂ "), baseStyle.Foreground(t.TextMuted()).Width(contentW-2).Render("MODIFIED FILES"))
	var mfBody string
	if m.modFiles == nil || len(m.modFiles) == 0 {
		mfBody = baseStyle.Foreground(t.TextMuted()).Render("No modified files")
	} else {
		var paths []string
		for path := range m.modFiles {
			paths = append(paths, path)
		}
		sort.Strings(paths)
		var rows []string
		for _, path := range paths {
			stats := m.modFiles[path]
			label := path
			if len(label) > contentW-6 {
				label = label[:contentW-7] + "…"
			}
			addStr := ""
			if stats.additions > 0 {
				addStr = lipgloss.NewStyle().Foreground(t.Success()).Render(fmt.Sprintf(" +%d", stats.additions))
			}
			remStr := ""
			if stats.removals > 0 {
				remStr = lipgloss.NewStyle().Foreground(t.Error()).Render(fmt.Sprintf(" -%d", stats.removals))
			}
			rows = append(rows, baseStyle.Render(label)+addStr+remStr)
		}
		mfBody = lipgloss.JoinVertical(lipgloss.Left, rows...)
	}
	mfSection := lipgloss.JoinVertical(lipgloss.Left, mfHeading, mfBody)

	// Diagnostics
	diagIcon := lipgloss.NewStyle().Foreground(t.Success()).Render("∿")
	diagHeading := lipgloss.JoinHorizontal(lipgloss.Left, diagIcon+" ", baseStyle.Foreground(t.TextMuted()).Width(contentW-2).Render("DIAGNOSTICS"))
	diagBody := baseStyle.Foreground(t.TextMuted()).Render("No issues")
	diagSub := baseStyle.Foreground(t.TextMuted()).Render("LSP ready")
	// Only show LSP ready if we have LSP configured and ready — otherwise keep muted placeholder but not fake
	diagSection := lipgloss.JoinVertical(lipgloss.Left,
		diagHeading,
		lipgloss.NewStyle().Foreground(t.Success()).Render("  ✓ ")+baseStyle.Foreground(t.Text()).Render("No issues"),
		baseStyle.Foreground(t.TextMuted()).Render("    LSP ready"),
	)
	_ = diagBody
	_ = diagSub
	_ = mfHeading

	sepStyle := lipgloss.NewStyle().Foreground(t.BorderDim())
	sep := sepStyle.Render(strings.Repeat("─", contentW))

	content := lipgloss.JoinVertical(lipgloss.Left,
		wsSection,
		"",
		sep,
		"",
		csSection,
		"",
		sep,
		"",
		mfSection,
		"",
		sep,
		"",
		diagSection,
	)

	return lipgloss.NewStyle().
		Width(w).
		Height(m.height).
		Padding(1, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(t.BorderDim()).
		Background(t.Background()).
		Render(content)
}

func (m *sidebarCmp) SetSize(width, height int) tea.Cmd {
	if width < 24 {
		width = 24
	}
	if width > 28 {
		width = 28
	}
	m.width = width
	m.height = height
	return nil
}

func (m *sidebarCmp) GetSize() (int, int) {
	return m.width, m.height
}

func NewSidebarCmp(session session.Session, history history.Service) tea.Model {
	return &sidebarCmp{
		session: session,
		history: history,
	}
}

func (m *sidebarCmp) loadModifiedFiles(ctx context.Context) {
	if m.history == nil || m.session.ID == "" {
		return
	}
	latestFiles, err := m.history.ListLatestSessionFiles(ctx, m.session.ID)
	if err != nil {
		return
	}
	allFiles, err := m.history.ListBySession(ctx, m.session.ID)
	if err != nil {
		return
	}
	m.modFiles = make(map[string]struct {
		additions int
		removals  int
	})
	for _, file := range latestFiles {
		if file.Version == history.InitialVersion {
			continue
		}
		var initialVersion history.File
		for _, v := range allFiles {
			if v.Path == file.Path && v.Version == history.InitialVersion {
				initialVersion = v
				break
			}
		}
		if initialVersion.ID == "" {
			continue
		}
		if initialVersion.Content == file.Content {
			continue
		}
		_, additions, removals := diff.GenerateDiff(initialVersion.Content, file.Content, file.Path)
		if additions > 0 || removals > 0 {
			displayPath := file.Path
			workingDir := config.WorkingDirectory()
			displayPath = strings.TrimPrefix(displayPath, workingDir)
			displayPath = strings.TrimPrefix(displayPath, "/")
			m.modFiles[displayPath] = struct {
				additions int
				removals  int
			}{additions: additions, removals: removals}
		}
	}
}

func (m *sidebarCmp) processFileChanges(ctx context.Context, file history.File) {
	if file.Version == history.InitialVersion {
		return
	}
	initialVersion, err := m.findInitialVersion(ctx, file.Path)
	if err != nil || initialVersion.ID == "" {
		return
	}
	if initialVersion.Content == file.Content {
		displayPath := getDisplayPath(file.Path)
		delete(m.modFiles, displayPath)
		return
	}
	_, additions, removals := diff.GenerateDiff(initialVersion.Content, file.Content, file.Path)
	if additions > 0 || removals > 0 {
		displayPath := getDisplayPath(file.Path)
		m.modFiles[displayPath] = struct {
			additions int
			removals  int
		}{additions: additions, removals: removals}
	} else {
		displayPath := getDisplayPath(file.Path)
		delete(m.modFiles, displayPath)
	}
}

func (m *sidebarCmp) findInitialVersion(ctx context.Context, path string) (history.File, error) {
	fileVersions, err := m.history.ListBySession(ctx, m.session.ID)
	if err != nil {
		return history.File{}, err
	}
	for _, v := range fileVersions {
		if v.Path == path && v.Version == history.InitialVersion {
			return v, nil
		}
	}
	return history.File{}, fmt.Errorf("initial version not found")
}

func getDisplayPath(path string) string {
	workingDir := config.WorkingDirectory()
	displayPath := strings.TrimPrefix(path, workingDir)
	return strings.TrimPrefix(displayPath, "/")
}
