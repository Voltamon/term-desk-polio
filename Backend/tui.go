package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type User struct {
	name      string
	status    string
	connected time.Time
}

type Model struct {
	cursor   int
	quitting bool

	users []User

	terminalContent []string
	commandInput    string
	inputFocused    bool

	width  int
	height int
}

func initialModel() Model {
	return Model{
		cursor:       0,
		quitting:     false,
		inputFocused: true,
		users: []User{
			{"john_doe", "active", time.Now().Add(-time.Hour * 2)},
			{"jane_smith", "idle", time.Now().Add(-time.Minute * 30)},
			{"admin", "active", time.Now().Add(-time.Minute * 5)},
		},
		terminalContent: []string{
			"Welcome to pty-terminal",
			"Type 'help' for available commands",
			"Current session started at " + time.Now().Format("15:04:05"),
			"",
		},
		commandInput: "",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.quitting = true
		return m, tea.Quit

	case "up":
		if !m.inputFocused && m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down":
		if !m.inputFocused && m.cursor < len(m.users)-1 {
			m.cursor++
		}
		return m, nil

	case "tab":
		m.inputFocused = !m.inputFocused
		return m, nil

	case "enter":
		if m.inputFocused && strings.TrimSpace(m.commandInput) != "" {
			return m.executeCommand()
		}
		return m, nil

	case "backspace":
		if m.inputFocused && len(m.commandInput) > 0 {
			m.commandInput = m.commandInput[:len(m.commandInput)-1]
		}
		return m, nil

	default:
		if m.inputFocused {
			m.commandInput += msg.String()
		}
		return m, nil
	}
}

func (m Model) executeCommand() (tea.Model, tea.Cmd) {
	cmd := strings.TrimSpace(m.commandInput)
	output := ""

	switch cmd {
	case "clear":
		m.terminalContent = []string{}
		m.commandInput = ""
		return m, nil

	case "exit":
		m.quitting = true
		return m, tea.Quit

	default:
		execOutput, err := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", cmd).CombinedOutput()
		if err != nil {
			output = fmt.Sprintf("(error) %v", err)
		} else {
			output = string(execOutput)
		}
	}

	m.terminalContent = append(m.terminalContent, fmt.Sprintf("> %s", cmd))
	if output != "" {
		m.terminalContent = append(m.terminalContent, output)
	}
	m.terminalContent = append(m.terminalContent, "")
	m.commandInput = ""

	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return "\n👋 Thanks for using pty-terminal!\n\n"
	}

	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	return m.renderTerminalInterface()
}

func (m Model) renderTerminalInterface() string {
	sidebarWidth := int(float64(m.width) * 0.20)
	mainAreaWidth := int(float64(m.width) * 0.80)

	commandHeight := 3
	terminalHeight := m.height - commandHeight - 2

	if sidebarWidth < 15 {
		sidebarWidth = 15
	}
	if mainAreaWidth < 30 {
		mainAreaWidth = 30
	}
	if terminalHeight < 5 {
		terminalHeight = 5
	}

	sidebarStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Width(sidebarWidth-2).
		Height(m.height-2).
		MarginRight(1).
		Padding(1, 1)

	mainTerminalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Width(mainAreaWidth-2).
		Height(terminalHeight).
		Padding(1, 2)

	commandLineStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666")).
		Width(mainAreaWidth-2).
		Height(1).
		Padding(0, 0, 0)

	sidebar := m.renderSidebar()
	mainTerminal := m.renderMainTerminal(terminalHeight - 4)
	commandLine := m.renderCommandLine()

	rightSide := lipgloss.JoinVertical(
		lipgloss.Left,
		mainTerminalStyle.Render(mainTerminal),
		commandLineStyle.Render(commandLine),
	)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebarStyle.Render(sidebar),
		rightSide,
	)
}

func (m Model) renderSidebar() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Align(lipgloss.Center).
		MarginBottom(1)

	userStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC"))

	selectedUserStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	var content strings.Builder
	content.WriteString(titleStyle.Render("Users"))
	content.WriteString("\n")
	content.WriteString(titleStyle.Render("In"))
	content.WriteString("\n")
	content.WriteString(titleStyle.Render("Session"))
	content.WriteString("\n\n")

	for i, user := range m.users {
		cursor := " "
		style := userStyle
		if !m.inputFocused && i == m.cursor {
			cursor = ">"
			style = selectedUserStyle
		}

		status := "•"
		if user.status == "idle" {
			status = "◦"
		}

		content.WriteString(fmt.Sprintf("%s %s %s\n", cursor, status, style.Render(user.name)))
	}

	return content.String()
}

func (m Model) renderMainTerminal(height int) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Align(lipgloss.Center).
		MarginBottom(2)

	contentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC"))

	var content strings.Builder
	content.WriteString(titleStyle.Render("pty-terminal"))
	content.WriteString("\n\n")

	startLine := 0
	if len(m.terminalContent) > height-3 {
		startLine = len(m.terminalContent) - (height - 3)
	}

	for i := startLine; i < len(m.terminalContent); i++ {
		content.WriteString(contentStyle.Render(m.terminalContent[i]))
		content.WriteString("\n")
	}

	return content.String()
}

func (m Model) renderCommandLine() string {
	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#FFFFFF"))

	prompt := "> "
	input := m.commandInput

	if m.inputFocused {
		input += cursorStyle.Render(" ")
	}

	return promptStyle.Render(prompt) + inputStyle.Render(input)
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running pty-terminal: %v\n", err)
		os.Exit(1)
	}
}
