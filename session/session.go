package session

import (
	"time"

	"github.com/satori/go.uuid"
)

type CommandSession struct {
	SessionID string
	Started   time.Time
}

func NewCommandSession() *CommandSession {
	s := &CommandSession{}
	s.SessionID = uuid.NewV4().String()
	return s
}

func (s CommandSession) Start() *CommandSession {
	s.Started = time.Now()
	return &s
}

func (s CommandSession) Duration() time.Duration {
	return time.Since(s.Started)
}
