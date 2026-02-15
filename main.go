package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateSetup state = iota
	stateRunning
)

type sessionType int

const (
	sessionFocus sessionType = iota
	sessionBreak
)

type tickMsg time.Time

type model struct {
	state       state
	sessionType sessionType

	// Setup
	focusInput   textinput.Model
	breakInput   textinput.Model
	currentInput int // 0 = focus, 1 = break

	// Timer
	focusDuration     time.Duration
	breakDuration     time.Duration
	timeLeft          time.Duration
	running           bool
	completedSessions int

	width  int
	height int
}

func initialModel() model {
	focusInput := textinput.New()
	focusInput.Placeholder = "25"
	focusInput.Focus()
	focusInput.CharLimit = 3
	focusInput.Width = 20

	breakInput := textinput.New()
	breakInput.Placeholder = "5"
	breakInput.CharLimit = 3
	breakInput.Width = 20

	return model{
		state:             stateSetup,
		sessionType:       sessionFocus,
		focusInput:        focusInput,
		breakInput:        breakInput,
		currentInput:      0,
		running:           false,
		completedSessions: 0,
	}

}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
		if m.state == stateSetup {
			return m.updateSetup(msg)
		} else {
			return m.updateRunning(msg)
		}
	case tickMsg:
		if m.running && m.timeLeft > 0 {
			m.timeLeft -= time.Second

			//cambio de sesion
			if m.timeLeft <= 0 {
				if m.sessionType == sessionFocus {
					m.completedSessions++
					m.sessionType = sessionBreak
					m.timeLeft = m.breakDuration
				} else {
					m.sessionType = sessionFocus
					m.timeLeft = m.focusDuration
				}
			}

			return m, tick()
		}

		return m, nil
	}

	return m, nil
}

func (m model) updateSetup(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "enter":
		if m.currentInput == 0 {
			m.currentInput = 1 //siguiente input
			m.focusInput.Blur()
			m.breakInput.Focus()
			return m, textinput.Blink
		} else { //iniciar timer
			focusMin := parseInput(m.focusInput.Value(), 25)
			breakMin := parseInput(m.breakInput.Value(), 5)

			m.focusDuration = time.Duration(focusMin) * time.Minute
			m.breakDuration = time.Duration(breakMin) * time.Minute
			m.timeLeft = m.focusDuration
			m.state = stateRunning
			m.running = true

			return m, tick()
		}

	case "tab", "shift+tab": //cambiar input
		if m.currentInput == 0 {
			m.currentInput = 1
			m.focusInput.Blur()
			m.breakInput.Focus()
		} else {
			m.currentInput = 0
			m.breakInput.Blur()
			m.focusInput.Focus()
		}
		return m, textinput.Blink
	}

	if m.currentInput == 0 {
		m.focusInput, cmd = m.focusInput.Update(msg)
	} else {
		m.breakInput, cmd = m.breakInput.Update(msg)
	}

	return m, cmd
}

func (m model) updateRunning(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "p", " ": //pausar/reanudar
		m.running = !m.running
		
	case "r": //reset
		if m.sessionType == sessionFocus {
			m.timeLeft = m.focusDuration
		} else {
			m.timeLeft = m.breakDuration
		}
		m.running = false
	case "s": //skip session
		if m.sessionType == sessionFocus {
			m.completedSessions++
			m.sessionType = sessionBreak
			m.timeLeft = m.breakDuration
		} else {
			m.sessionType = sessionFocus
			m.timeLeft = m.focusDuration
		}
		m.running = true
	case "n": //nueva config
		m.state = stateSetup
		m.running = false
		m.completedSessions = 0
		m.currentInput = 0
		m.focusInput.Focus()
		return m, textinput.Blink
	}

	return m, nil
}

func (m model) View() string {
	if m.state == stateSetup {
		return m.viewSetup()
	}
	return m.viewRunning()
}

func (m model) viewSetup() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF6B6B")).
		MarginBottom(2).
		Align(lipgloss.Center)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#95E1D3")).
		Bold(true)

	inputStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#4ECDC4")).
		Padding(0, 1).
		MarginBottom(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#999")).
		MarginTop(2).
		Italic(true)

	title := titleStyle.Render(" POMODORO TIMER")

	focusLabel := labelStyle.Render("Sesión de enfoque (minutos):")
	focusInputView := inputStyle.Render(m.focusInput.View())

	breakLabel := labelStyle.Render("Sesión de descanso (minutos):")
	breakInputView := inputStyle.Render(m.breakInput.View())

	help := helpStyle.Render("Tab: cambiar campo • Enter: siguiente/iniciar • Ctrl+C: salir")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		focusLabel,
		focusInputView,
		breakLabel,
		breakInputView,
		help,
	)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		content,
	)
}

func (m model) viewRunning() string {
	containerStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(2, 4).
		Align(lipgloss.Center)

	sessionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFE66D")).
		MarginBottom(1)

	timerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#4ECDC4")).
		MarginBottom(2)

	buttonStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#4ECDC4")).
		Foreground(lipgloss.Color("#000")).
		Padding(0, 2).
		MarginRight(1).
		Bold(true)

	statsStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#95E1D3")).
		MarginTop(2)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#999")).
		MarginTop(2).
		Italic(true)

	// Contenido
	sessionText := " ENFOQUE"
	emoji := ""
	if m.sessionType == sessionBreak {
		sessionText = " DESCANSO"
		emoji = ""
	}

	session := sessionStyle.Render(sessionText)

	minutes := int(m.timeLeft.Minutes())
	seconds := int(m.timeLeft.Seconds()) % 60
	timerText := fmt.Sprintf("%s %02d:%02d", emoji, minutes, seconds)
	timer := timerStyle.Render(timerText)

	// Botones
	var playPauseBtn string
	if m.running {
		playPauseBtn = buttonStyle.Render(" Pausar (P)")
	} else {
		playPauseBtn = buttonStyle.Render(" Iniciar (P)")
	}

	resetBtn := buttonStyle.Render(" Reset (R)")
	skipBtn := buttonStyle.Render(" Skip (S)")
	newBtn := buttonStyle.Render(" Config (N)")

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Top,
		playPauseBtn,
		resetBtn,
		skipBtn,
		newBtn,
	)

	stats := statsStyle.Render(fmt.Sprintf("Sesiones completadas: %d", m.completedSessions))
	help := helpStyle.Render("Q: salir • Espacio: pausar/reanudar")

	content := lipgloss.JoinVertical(
		lipgloss.Center,
		session,
		timer,
		buttons,
		stats,
		help,
	)

	box := containerStyle.Render(content)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		box,
	)
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func parseInput(input string, defaultVal int) int {
	if input == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(input)
	if err != nil || val <= 0 {
		return defaultVal
	}
	return val
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
