package main

import (
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	//"github.com/charmbracelet/lipgloss"
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
		if m.running {
			return m, tick()
		}
	case "r": //reset
		if m.sessionType == sessionFocus {
			m.timeLeft = m.focusDuration
		} else {
			m.timeLeft = m.breakDuration
		}
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
		return m, tick()
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
