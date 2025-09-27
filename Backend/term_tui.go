package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
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

	terminalContent  []string
	terminalViewport viewport.Model
	commandInput     string
	inputFocused     bool

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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Update viewport size when window size changes
		if m.width > 0 && m.height > 0 {
			mainAreaWidth := int(float64(m.width) * 0.80)
			commandHeight := 3
			terminalHeight := m.height - commandHeight - 2

			if mainAreaWidth < 30 {
				mainAreaWidth = 30
			}
			if terminalHeight < 5 {
				terminalHeight = 5
			}

			// Create style to calculate frame size
			mainTerminalStyle := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#666666")).
				Width(mainAreaWidth-2).
				Height(terminalHeight).
				Padding(1, 2)

			vPad, hPad := mainTerminalStyle.GetFrameSize()
			m.terminalViewport.Width = mainAreaWidth - 2 - hPad
			m.terminalViewport.Height = terminalHeight - vPad

			// Update viewport content
			content := strings.Join(m.terminalContent, "\n")
			m.terminalViewport.SetContent(content)
		}

		return m, nil
	}

	// Update viewport
	m.terminalViewport, cmd = m.terminalViewport.Update(msg)
	return m, cmd
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.quitting = true
		return m, tea.Quit

	case "up":
		if !m.inputFocused {
			if m.cursor > 0 {
				m.cursor--
			} else {
				// Scroll viewport up when in terminal area
				m.terminalViewport.LineUp(1)
			}
		}
		return m, nil

	case "down":
		if !m.inputFocused {
			if m.cursor < len(m.users)-1 {
				m.cursor++
			} else {
				// Scroll viewport down when in terminal area
				m.terminalViewport.LineDown(1)
			}
		}
		return m, nil

	case "pgup":
		if !m.inputFocused {
			m.terminalViewport.HalfPageUp()
		}
		return m, nil

	case "pgdn":
		if !m.inputFocused {
			m.terminalViewport.HalfPageDown()
		}
		return m, nil

	case "home":
		if !m.inputFocused {
			m.terminalViewport.GotoTop()
		}
		return m, nil

	case "end":
		if !m.inputFocused {
			m.terminalViewport.GotoBottom()
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
		// Update viewport content after clearing
		content := strings.Join(m.terminalContent, "\n")
		m.terminalViewport.SetContent(content)
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

	// Update viewport content after adding new content
	content := strings.Join(m.terminalContent, "\n")
	m.terminalViewport.SetContent(content)

	// Auto-scroll to bottom to show latest content
	m.terminalViewport.GotoBottom()

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
	mainTerminal := m.renderMainTerminal()
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

func (m Model) renderMainTerminal() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Align(lipgloss.Center).
		MarginBottom(2)

	var content strings.Builder
	content.WriteString(titleStyle.Render("pty-terminal"))
	content.WriteString("\n\n")

	// Render the viewport which contains all the scrollable terminal content
	content.WriteString(m.terminalViewport.View())

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
