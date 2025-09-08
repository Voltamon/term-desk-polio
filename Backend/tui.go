package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PTY Terminal TUI - A terminal interface with sidebar and command line

// User represents a user in session
type User struct {
	name      string
	status    string
	connected time.Time
}

// Model represents the application state
type Model struct {
	// Navigation and state
	cursor   int
	quitting bool

	// Users in session
	users []User

	// Terminal content
	terminalContent []string
	commandInput    string
	inputFocused    bool

	// UI dimensions
	width  int
	height int
}

// Initial state of the application
func initialModel() Model {
	return Model{
		cursor:       0,
		quitting:     false,
		inputFocused: true, // Start with command input focused
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

// Init returns initial commands
func (m Model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

// Update handles messages and updates the model
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

// Handle key presses
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		if !m.inputFocused && m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "j":
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

// Execute command in terminal
func (m Model) executeCommand() (tea.Model, tea.Cmd) {
	cmd := strings.TrimSpace(m.commandInput)
	output := ""

	switch cmd {
	case "help":
		output = "Available commands: help, clear, users, time, exit"
	case "clear":
		m.terminalContent = []string{}
		m.commandInput = ""
		return m, nil
	case "users":
		output = fmt.Sprintf("Active users: %d", len(m.users))
	case "time":
		output = time.Now().Format("Current time: 15:04:05")
	case "exit":
		m.quitting = true
		return m, tea.Quit
	default:
		output = fmt.Sprintf("Unknown command: %s", cmd)
	}

	// Add command and output to terminal
	m.terminalContent = append(m.terminalContent, fmt.Sprintf("> %s", cmd))
	if output != "" {
		m.terminalContent = append(m.terminalContent, output)
	}
	m.terminalContent = append(m.terminalContent, "")

	// Clear command input
	m.commandInput = ""

	return m, nil
}

// View renders the current screen
func (m Model) View() string {
	if m.quitting {
		return "\n👋 Thanks for using pty-terminal!\n\n"
	}

	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	return m.renderTerminalInterface()
}

// Render the complete terminal interface
func (m Model) renderTerminalInterface() string {
	// Calculate dimensions based on percentages
	sidebarWidth := int(float64(m.width) * 0.20)  // 20% of screen width
	mainAreaWidth := int(float64(m.width) * 0.80) // 80% of screen width

	// Command line takes exactly 1 line + borders (3 lines total)
	commandHeight := 3
	// Terminal takes the remaining height
	terminalHeight := m.height - commandHeight - 2 // Account for top/bottom borders

	// Ensure minimum dimensions
	if sidebarWidth < 15 {
		sidebarWidth = 15
	}
	if mainAreaWidth < 30 {
		mainAreaWidth = 30
	}
	if terminalHeight < 5 {
		terminalHeight = 5
	}

	// Create styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#666666"))

	// Sidebar takes full height
	sidebarStyle := borderStyle.Copy().
		Width(sidebarWidth-2). // Account for border
		Height(m.height-2).    // Full height minus top/bottom borders
		Padding(1, 1)

	// Terminal takes remaining height above command line
	mainTerminalStyle := borderStyle.Copy().
		Width(mainAreaWidth-2). // Account for border
		Height(terminalHeight). // Account for border
		Padding(1, 2)

	// Command line is exactly 1 line high
	commandLineStyle := borderStyle.Copy().
		Width(mainAreaWidth-2). // Account for border
		Height(1).              // Exactly 1 line content + 2 for borders = 3 total
		Padding(0, 0, 0)
	// Render sidebar
	sidebar := m.renderSidebar()

	// Render main terminal area
	mainTerminal := m.renderMainTerminal(terminalHeight - 4) // Account for padding and borders

	// Render command line
	commandLine := m.renderCommandLine()

	// Layout the right side (terminal + command line vertically)
	rightSide := lipgloss.JoinVertical(
		lipgloss.Left,
		mainTerminalStyle.Render(mainTerminal),
		commandLineStyle.Render(commandLine),
	)

	// Layout the complete interface (sidebar + right side horizontally)
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebarStyle.Render(sidebar),
		rightSide,
	)
}

// Render sidebar with users
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

		status := "•" // Active dot
		if user.status == "idle" {
			status = "◦" // Idle circle
		}

		content.WriteString(fmt.Sprintf("%s %s %s\n", cursor, status, style.Render(user.name)))
	}

	return content.String()
}

// Render main terminal area
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

	// Calculate how many lines we can show
	startLine := 0
	if len(m.terminalContent) > height-3 {
		startLine = len(m.terminalContent) - (height - 3)
	}

	// Show terminal content
	for i := startLine; i < len(m.terminalContent); i++ {
		content.WriteString(contentStyle.Render(m.terminalContent[i]))
		content.WriteString("\n")
	}

	return content.String()
}

// Render command line input
func (m Model) renderCommandLine() string {
	promptStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	cursorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#FFFFFF"))

	prompt := "cmd line "
	input := m.commandInput

	if m.inputFocused {
		// Add blinking cursor
		input += cursorStyle.Render(" ")
	}

	return promptStyle.Render(prompt) + inputStyle.Render(input)
}

func main() {
	// Create the program
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(), // Use alternate screen buffer
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running pty-terminal: %v\n", err)
		os.Exit(1)
	}
}
