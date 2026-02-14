package main

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
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



