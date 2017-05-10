package session

import (
	"time"

	"github.com/satori/go.uuid"
)

type CommandSession struct {
	SessionID string
	Started   time.Time
}

//CurrentSession stores the last pointer created by NewCommandSession
var CurrentSession *CommandSession

func NewCommandSession() *CommandSession {
	s := &CommandSession{}
	s.SessionID = uuid.NewV4().String()

	//persist the new session as the current session
	CurrentSession = s
	return s
}

func (s CommandSession) Start() *CommandSession {
	s.Started = time.Now()
	return &s
}

func (s CommandSession) Duration() time.Duration {
	return time.Since(s.Started)
}
