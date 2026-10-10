package layout

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jisunahamed/torvecode/internal/tui/theme"
)

type SplitPaneLayout interface {
	tea.Model
	Sizeable
	Bindings
	SetLeftPanel(panel Container) tea.Cmd
	SetRightPanel(panel Container) tea.Cmd
	SetBottomPanel(panel Container) tea.Cmd

	ClearLeftPanel() tea.Cmd
	ClearRightPanel() tea.Cmd
	ClearBottomPanel() tea.Cmd
}

type splitPaneLayout struct {
	width  int
	height int
	// ratio and verticalRatio are legacy options kept for API compatibility;
	// current layout uses fixed percentages: sidebar 24% (clamped 24–28) and
	// bottom panel height = H/6 clamped 3–8. These fields are ignored.
	ratio         float64
	verticalRatio float64

	rightPanel  Container
	leftPanel   Container
	bottomPanel Container
	// rightWidth tracks the last computed right panel width (0 when hidden at W<80)
	rightWidth int
}

type SplitPaneOption func(*splitPaneLayout)

func (s *splitPaneLayout) Init() tea.Cmd {
	var cmds []tea.Cmd

	if s.leftPanel != nil {
		cmds = append(cmds, s.leftPanel.Init())
	}

	if s.rightPanel != nil {
		cmds = append(cmds, s.rightPanel.Init())
	}

	if s.bottomPanel != nil {
		cmds = append(cmds, s.bottomPanel.Init())
	}

	return tea.Batch(cmds...)
}

func (s *splitPaneLayout) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return s, s.SetSize(msg.Width, msg.Height)
	}

	if s.rightPanel != nil {
		u, cmd := s.rightPanel.Update(msg)
		s.rightPanel = u.(Container)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if s.leftPanel != nil {
		u, cmd := s.leftPanel.Update(msg)
		s.leftPanel = u.(Container)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if s.bottomPanel != nil {
		u, cmd := s.bottomPanel.Update(msg)
		s.bottomPanel = u.(Container)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return s, tea.Batch(cmds...)
}

func (s *splitPaneLayout) View() string {
	var topSection string

	// When rightWidth==0 (narrow W<80) we hide the sidebar entirely and must not
	// JoinHorizontal an empty right panel — it would still emit a zero-width cell.
	showRight := s.rightPanel != nil && s.rightWidth > 0
	if s.leftPanel != nil && showRight {
		leftView := s.leftPanel.View()
		rightView := s.rightPanel.View()
		topSection = lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
	} else if s.leftPanel != nil {
		topSection = s.leftPanel.View()
	} else if showRight {
		topSection = s.rightPanel.View()
	} else {
		topSection = ""
	}

	var finalView string

	if s.bottomPanel != nil && topSection != "" {
		bottomView := s.bottomPanel.View()
		finalView = lipgloss.JoinVertical(lipgloss.Left, topSection, bottomView)
	} else if s.bottomPanel != nil {
		finalView = s.bottomPanel.View()
	} else {
		finalView = topSection
	}

	if finalView != "" {
		t := theme.CurrentTheme()

		style := lipgloss.NewStyle().
			Width(s.width).
			Height(s.height).
			Background(t.Background())

		return style.Render(finalView)
	}

	return finalView
}

func (s *splitPaneLayout) SetSize(width, height int) tea.Cmd {
	if width <= 0 || height <= 0 {
		return nil
	}
	s.width = width
	s.height = height

	var topHeight, bottomHeight int
	if s.bottomPanel != nil {
		// Input composer: min 3, max 8, otherwise height/6
		bottomHeight = height / 6
		if bottomHeight < 3 {
			bottomHeight = 3
		}
		if bottomHeight > 8 {
			bottomHeight = 8
		}
		topHeight = height - bottomHeight
	} else {
		topHeight = height
		bottomHeight = 0
	}

	var leftWidth, rightWidth int
	if s.leftPanel != nil && s.rightPanel != nil {
		// Responsive: hide sidebar when narrow, clamp 24-28 cols when visible
		if width < 80 {
			leftWidth = width
			rightWidth = 0
		} else {
			rightWidth = width * 24 / 100
			if rightWidth < 24 {
				rightWidth = 24
			}
			if rightWidth > 28 {
				rightWidth = 28
			}
			leftWidth = width - rightWidth
		}
	} else if s.leftPanel != nil {
		leftWidth = width
		rightWidth = 0
	} else if s.rightPanel != nil {
		leftWidth = 0
		rightWidth = width
	}
	// Persist rightWidth for View() responsive hide check
	s.rightWidth = rightWidth

	var cmds []tea.Cmd
	if s.leftPanel != nil {
		cmd := s.leftPanel.SetSize(leftWidth, topHeight)
		cmds = append(cmds, cmd)
	}

	if s.rightPanel != nil {
		// Don't call SetSize(0,...) when hidden — sidebarCmp.SetSize clamps 24..28
		if rightWidth > 0 {
			cmd := s.rightPanel.SetSize(rightWidth, topHeight)
			cmds = append(cmds, cmd)
		}
	}

	if s.bottomPanel != nil {
		cmd := s.bottomPanel.SetSize(width, bottomHeight)
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (s *splitPaneLayout) GetSize() (int, int) {
	return s.width, s.height
}

func (s *splitPaneLayout) SetLeftPanel(panel Container) tea.Cmd {
	s.leftPanel = panel
	if s.width > 0 && s.height > 0 {
		return s.SetSize(s.width, s.height)
	}
	return nil
}

func (s *splitPaneLayout) SetRightPanel(panel Container) tea.Cmd {
	s.rightPanel = panel
	if s.width > 0 && s.height > 0 {
		return s.SetSize(s.width, s.height)
	}
	return nil
}

func (s *splitPaneLayout) SetBottomPanel(panel Container) tea.Cmd {
	s.bottomPanel = panel
	if s.width > 0 && s.height > 0 {
		return s.SetSize(s.width, s.height)
	}
	return nil
}

func (s *splitPaneLayout) ClearLeftPanel() tea.Cmd {
	s.leftPanel = nil
	if s.width > 0 && s.height > 0 {
		return s.SetSize(s.width, s.height)
	}
	return nil
}

func (s *splitPaneLayout) ClearRightPanel() tea.Cmd {
	s.rightPanel = nil
	if s.width > 0 && s.height > 0 {
		return s.SetSize(s.width, s.height)
	}
	return nil
}

func (s *splitPaneLayout) ClearBottomPanel() tea.Cmd {
	s.bottomPanel = nil
	if s.width > 0 && s.height > 0 {
		return s.SetSize(s.width, s.height)
	}
	return nil
}

func (s *splitPaneLayout) BindingKeys() []key.Binding {
	keys := []key.Binding{}
	if s.leftPanel != nil {
		if b, ok := s.leftPanel.(Bindings); ok {
			keys = append(keys, b.BindingKeys()...)
		}
	}
	if s.rightPanel != nil {
		if b, ok := s.rightPanel.(Bindings); ok {
			keys = append(keys, b.BindingKeys()...)
		}
	}
	if s.bottomPanel != nil {
		if b, ok := s.bottomPanel.(Bindings); ok {
			keys = append(keys, b.BindingKeys()...)
		}
	}
	return keys
}

func NewSplitPane(options ...SplitPaneOption) SplitPaneLayout {

	layout := &splitPaneLayout{
		ratio:         0.7,
		verticalRatio: 0.9, // Default 90% for top section, 10% for bottom
	}
	for _, option := range options {
		option(layout)
	}
	return layout
}

func WithLeftPanel(panel Container) SplitPaneOption {
	return func(s *splitPaneLayout) {
		s.leftPanel = panel
	}
}

func WithRightPanel(panel Container) SplitPaneOption {
	return func(s *splitPaneLayout) {
		s.rightPanel = panel
	}
}

func WithRatio(ratio float64) SplitPaneOption {
	return func(s *splitPaneLayout) {
		s.ratio = ratio
	}
}

func WithBottomPanel(panel Container) SplitPaneOption {
	return func(s *splitPaneLayout) {
		s.bottomPanel = panel
	}
}

func WithVerticalRatio(ratio float64) SplitPaneOption {
	return func(s *splitPaneLayout) {
		s.verticalRatio = ratio
	}
}
