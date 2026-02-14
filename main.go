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
