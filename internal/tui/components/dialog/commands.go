package dialog

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	utilComponents "github.com/jisunahamed/torvecode/internal/tui/components/util"
	"github.com/jisunahamed/torvecode/internal/tui/layout"
	"github.com/jisunahamed/torvecode/internal/tui/styles"
	"github.com/jisunahamed/torvecode/internal/tui/theme"
	"github.com/jisunahamed/torvecode/internal/tui/util"
)

// Command represents a command that can be executed
type Command struct {
	ID          string
	Title       string
	Description string
	Handler     func(cmd Command) tea.Cmd
}

func (ci Command) Render(selected bool, width int) string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	itemStyle := baseStyle.
		Foreground(t.Text()).
		Background(t.Background())
	descStyle := itemStyle.Foreground(t.TextMuted())

	if selected {
		itemStyle = itemStyle.
			Background(t.Primary()).
			Foreground(t.Background()).
			Bold(true)
		descStyle = descStyle.
			Background(t.Primary()).
			Foreground(t.Background())
	}

	titleWidth := min(18, max(10, width/3))
	title := itemStyle.PaddingLeft(1).Width(titleWidth).Render(ci.Title)
	description := descStyle.PaddingRight(1).Width(max(1, width-titleWidth)).Render(ci.Description)
	return lipgloss.JoinHorizontal(lipgloss.Top, title, description)
}

// CommandSelectedMsg is sent when a command is selected
type CommandSelectedMsg struct {
	Command Command
}

// CloseCommandDialogMsg is sent when the command dialog is closed
type CloseCommandDialogMsg struct{}

// OpenCommandDialogMsg opens the command palette from an empty chat prompt.
type OpenCommandDialogMsg struct{}

// CommandDialog interface for the command selection dialog
type CommandDialog interface {
	tea.Model
	layout.Bindings
	SetCommands(commands []Command)
}

type commandDialogCmp struct {
	listView utilComponents.SimpleList[Command]
	commands []Command
	query    string
	width    int
	height   int
}

type commandKeyMap struct {
	Enter  key.Binding
	Escape key.Binding
}

var commandKeys = commandKeyMap{
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select command"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "close"),
	),
}

func (c *commandDialogCmp) Init() tea.Cmd {
	return c.listView.Init()
}

func (c *commandDialogCmp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, commandKeys.Enter):
			selectedItem, idx := c.listView.GetSelectedItem()
			if idx != -1 {
				return c, util.CmdHandler(CommandSelectedMsg{
					Command: selectedItem,
				})
			}
		case key.Matches(msg, commandKeys.Escape):
			c.query = ""
			return c, util.CmdHandler(CloseCommandDialogMsg{})
		case msg.Type == tea.KeyBackspace:
			if c.query != "" {
				c.query = strings.TrimSuffix(c.query, string([]rune(c.query)[len([]rune(c.query))-1]))
				c.filter()
			}
			return c, nil
		case len(msg.Runes) > 0:
			c.query += string(msg.Runes)
			c.filter()
			return c, nil
		}
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
	}

	u, cmd := c.listView.Update(msg)
	c.listView = u.(utilComponents.SimpleList[Command])
	cmds = append(cmds, cmd)

	return c, tea.Batch(cmds...)
}

func (c *commandDialogCmp) View() string {
	t := theme.CurrentTheme()
	baseStyle := styles.BaseStyle()

	maxWidth := 40

	commands := c.listView.GetItems()

	for _, cmd := range commands {
		if len(cmd.Title) > maxWidth-4 {
			maxWidth = len(cmd.Title) + 4
		}
		if cmd.Description != "" {
			if len(cmd.Description) > maxWidth-4 {
				maxWidth = len(cmd.Description) + 4
			}
		}
	}

	c.listView.SetMaxWidth(maxWidth)

	title := baseStyle.
		Foreground(t.Primary()).
		Bold(true).
		Width(maxWidth).
		Padding(0, 1).
		Render("/" + c.query)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		baseStyle.Width(maxWidth).Render(""),
		baseStyle.Width(maxWidth).Render(c.listView.View()),
		baseStyle.Width(maxWidth).Render(""),
	)

	return baseStyle.Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderBackground(t.Background()).
		BorderForeground(t.TextMuted()).
		Width(lipgloss.Width(content) + 4).
		Render(content)
}

func (c *commandDialogCmp) BindingKeys() []key.Binding {
	return layout.KeyMapToSlice(commandKeys)
}

func (c *commandDialogCmp) SetCommands(commands []Command) {
	c.commands = commands
	c.query = ""
	c.filter()
}

func (c *commandDialogCmp) filter() {
	query := strings.ToLower(c.query)
	filtered := make([]Command, 0, len(c.commands))
	for _, command := range c.commands {
		if query == "" || strings.Contains(strings.ToLower(command.Title+" "+command.Description), query) {
			filtered = append(filtered, command)
		}
	}
	c.listView.SetItems(filtered)
}

// NewCommandDialogCmp creates a new command selection dialog
func NewCommandDialogCmp() CommandDialog {
	listView := utilComponents.NewSimpleList[Command](
		[]Command{},
		10,
		"No commands available",
		false,
	)
	return &commandDialogCmp{
		listView: listView,
	}
}
