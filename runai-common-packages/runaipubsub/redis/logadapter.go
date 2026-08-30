package redis

import (
	"fmt"
	"strings"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/logger"
)

func NewLogger(l logger.Loggerer) *Logger {
	return &Logger{
		Loggerer: l,
	}
}

type Logger struct {
	logger.Loggerer
}

func (l *Logger) Error(msg string, err error, fields watermill.LogFields) {
	if strings.Contains(msg, "read fail") && strings.Contains(err.Error(), "EOF") {
		l.Infof("%s, error: %s. %+v", msg, err.Error(), fields)
		return
	}

	l.Errorf("%s, error: %s. %+v", msg, err.Error(), fields)
}

func (l *Logger) Info(msg string, fields watermill.LogFields) {
	l.Infof("%s: %+v", msg, fields)
}

func (l *Logger) Debug(msg string, fields watermill.LogFields) {
	l.Debugf("%s: %+v", msg, fields)
}

func (l *Logger) Trace(msg string, fields watermill.LogFields) {}

func (l *Logger) Warn(msg string, fields watermill.LogFields) {
	l.Warnf(fmt.Sprintf("%s: %+v", msg, fields))
}

func (l *Logger) With(fields watermill.LogFields) watermill.LoggerAdapter {
	return l
}
