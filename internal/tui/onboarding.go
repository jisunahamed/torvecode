package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jisunahamed/torvecode/internal/auth"
)

type onboardingStage int

const (
	stageChoose onboardingStage = iota
	stageAPIKey
	stageStartingBrowser
	stageWaitingBrowser
)

var errOnboardingCancelled = errors.New("sign in cancelled")

type deviceStartedMsg struct {
	grant auth.DeviceAuthorization
	url   string
}

type credentialReadyMsg struct{ credential auth.Credential }
type onboardingErrorMsg struct{ err error }

type onboardingModel struct {
	ctx        context.Context
	width      int
	height     int
	selected   int
	stage      onboardingStage
	apiKey     textinput.Model
	userCode   string
	verifyURL  string
	status     string
	credential auth.Credential
	err        error
}

func newOnboardingModel(ctx context.Context) onboardingModel {
	input := textinput.New()
	input.Placeholder = "sk-trv-..."
	input.Prompt = "API key  "
	input.EchoMode = textinput.EchoPassword
	input.EchoCharacter = '•'
	input.CharLimit = 512
	return onboardingModel{
		ctx:    ctx,
		stage:  stageChoose,
		apiKey: input,
		status: "Connect your Torve AI account to start coding.",
	}
}

func (m onboardingModel) Init() tea.Cmd { return textinput.Blink }

func startDevice(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		grant, err := auth.StartDevice(ctx, auth.DeviceName())
		if err != nil {
			return onboardingErrorMsg{err: err}
		}
		loginURL := grant.VerificationURIComplete
		if loginURL == "" {
			loginURL = auth.AddCodeToURL(grant.VerificationURI, grant.UserCode)
		}
		_ = auth.OpenBrowser(loginURL)
		return deviceStartedMsg{grant: grant, url: loginURL}
	}
}

func pollDevice(ctx context.Context, grant auth.DeviceAuthorization) tea.Cmd {
	return func() tea.Msg {
		credential, err := auth.PollDevice(ctx, grant)
		if err != nil {
			return onboardingErrorMsg{err: err}
		}
		if err := auth.Save(credential); err != nil {
			return onboardingErrorMsg{err: fmt.Errorf("secure credential storage failed: %w", err)}
		}
		return credentialReadyMsg{credential: credential}
	}
}

func validateAPIKey(ctx context.Context, key string) tea.Cmd {
	return func() tea.Msg {
		credential := auth.Credential{AccessToken: key, Method: "api-key"}
		requestCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
		defer cancel()
		if _, err := auth.FetchModels(requestCtx, credential); err != nil {
			return onboardingErrorMsg{err: fmt.Errorf("API key rejected: %w", err)}
		}
		if err := auth.Save(credential); err != nil {
			return onboardingErrorMsg{err: fmt.Errorf("secure credential storage failed: %w", err)}
		}
		return credentialReadyMsg{credential: credential}
	}
}

func (m onboardingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.apiKey.Width = minInt(56, maxInt(24, msg.Width-18))
		return m, nil
	case credentialReadyMsg:
		m.credential = msg.credential
		m.status = "Connected. Opening your workspace…"
		return m, tea.Quit
	case deviceStartedMsg:
		m.stage = stageWaitingBrowser
		m.userCode = msg.grant.UserCode
		m.verifyURL = msg.url
		m.status = "Approve this device in your browser."
		return m, pollDevice(m.ctx, msg.grant)
	case onboardingErrorMsg:
		m.err = msg.err
		m.status = msg.err.Error()
		if m.stage == stageAPIKey {
			m.apiKey.SetValue("")
			m.apiKey.Focus()
			return m, textinput.Blink
		}
		m.stage = stageChoose
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.err = errOnboardingCancelled
			return m, tea.Quit
		}
		if msg.String() == "esc" && m.stage != stageChoose {
			m.stage = stageChoose
			m.err = nil
			m.status = "Choose how you want to connect."
			m.apiKey.Blur()
			return m, nil
		}

		switch m.stage {
		case stageChoose:
			switch msg.String() {
			case "up", "k", "shift+tab":
				m.selected = (m.selected + 1) % 2
			case "down", "j", "tab":
				m.selected = (m.selected + 1) % 2
			case "enter":
				m.err = nil
				if m.selected == 0 {
					m.stage = stageStartingBrowser
					m.status = "Creating a secure device code…"
					return m, startDevice(m.ctx)
				}
				m.stage = stageAPIKey
				m.status = "Paste a Torve AI API key. It will be stored securely."
				return m, m.apiKey.Focus()
			}
		case stageAPIKey:
			if msg.String() == "enter" {
				key := strings.TrimSpace(m.apiKey.Value())
				if key == "" {
					m.err = errors.New("API key cannot be empty")
					m.status = m.err.Error()
					return m, nil
				}
				m.apiKey.Blur()
				m.status = "Checking your API key…"
				return m, validateAPIKey(m.ctx, key)
			}
			var cmd tea.Cmd
			m.apiKey, cmd = m.apiKey.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

func (m onboardingModel) View() string {
	width := m.width
	height := m.height
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	cardWidth := minInt(68, maxInt(36, width-6))

	green := lipgloss.Color("#4ADE80")
	cyan := lipgloss.Color("#35D0BA")
	text := lipgloss.Color("#D7E2DA")
	muted := lipgloss.Color("#748579")
	panel := lipgloss.Color("#0C1811")
	if os.Getenv("NO_COLOR") != "" {
		green, cyan, text, muted, panel = "", "", "", "", ""
	}

	mark := lipgloss.NewStyle().Bold(true).Foreground(text).Render("TORVE") +
		lipgloss.NewStyle().Bold(true).Foreground(green).Render("CODE")
	tagline := lipgloss.NewStyle().Foreground(muted).Render("Your codebase. Your Torve models. One terminal.")

	cwd, _ := os.Getwd()
	folder := filepath.Base(cwd)
	if folder == "." || folder == string(filepath.Separator) || folder == "" {
		folder = cwd
	}
	location := lipgloss.NewStyle().Foreground(muted).Render("⌁ " + folder)

	var body string
	switch m.stage {
	case stageChoose:
		options := []string{"Login with website", "Use Torve AI API key"}
		rows := make([]string, 0, len(options))
		for i, option := range options {
			prefix := "  "
			style := lipgloss.NewStyle().Width(cardWidth-6).Padding(0, 1).Foreground(text)
			if i == m.selected {
				prefix = "› "
				style = style.Bold(true).Foreground(green).Background(panel)
			}
			rows = append(rows, prefix+style.Render(option))
		}
		body = lipgloss.JoinVertical(lipgloss.Left, rows...)
	case stageAPIKey:
		body = lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Foreground(text).Render("Connect with an API key"),
			"",
			m.apiKey.View(),
		)
	case stageStartingBrowser:
		body = lipgloss.NewStyle().Foreground(cyan).Render("◌ Creating a secure device code…")
	case stageWaitingBrowser:
		body = lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Foreground(text).Render("Confirm this code in your browser"),
			lipgloss.NewStyle().Bold(true).Foreground(green).MarginTop(1).Render(m.userCode),
			lipgloss.NewStyle().Foreground(muted).Render(m.verifyURL),
		)
	}

	statusColor := muted
	if m.err != nil {
		statusColor = lipgloss.Color("#F38BA8")
	}
	status := lipgloss.NewStyle().Width(cardWidth - 4).Foreground(statusColor).Render(m.status)
	card := lipgloss.NewStyle().
		Width(cardWidth).
		Border(lipgloss.NormalBorder()).
		BorderLeft(true).
		BorderTop(false).
		BorderRight(false).
		BorderBottom(false).
		BorderForeground(cyan).
		Padding(1, 2).
		Render(lipgloss.JoinVertical(lipgloss.Left, body, "", status))

	help := "↑/↓ choose   enter continue   esc back   ctrl+c quit"
	if m.stage == stageWaitingBrowser || m.stage == stageStartingBrowser {
		help = "Complete sign-in in your browser   esc cancel   ctrl+c quit"
	}
	content := lipgloss.JoinVertical(lipgloss.Center,
		mark,
		tagline,
		"",
		card,
		"",
		location,
		lipgloss.NewStyle().Foreground(muted).Render(help),
	)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

// RunOnboarding presents authentication inside the terminal before the coding
// workspace is initialized. It keeps first-run users inside the same UI as chat.
func RunOnboarding(ctx context.Context) (auth.Credential, error) {
	program := tea.NewProgram(newOnboardingModel(ctx), tea.WithAltScreen())
	result, err := program.Run()
	if err != nil {
		return auth.Credential{}, err
	}
	model, ok := result.(onboardingModel)
	if !ok {
		return auth.Credential{}, errors.New("authentication screen ended unexpectedly")
	}
	if model.err != nil {
		return auth.Credential{}, model.err
	}
	if model.credential.AccessToken == "" {
		return auth.Credential{}, errOnboardingCancelled
	}
	return model.credential, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
