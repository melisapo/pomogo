package main

import (
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


func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
