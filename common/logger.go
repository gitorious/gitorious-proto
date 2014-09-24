package common

import (
	"fmt"
	"log"
)

type Logger interface {
	Printf(string, ...interface{})
}

type SessionLogger struct {
	Target    *log.Logger
	SessionId string
}

func (l *SessionLogger) Printf(format string, args ...interface{}) {
	l.Target.Printf("[%v] %v", l.SessionId, fmt.Sprintf(format, args...))
}
