package logger

type dummyLogger struct{}

func NewDummyLogger() *dummyLogger {
	return &dummyLogger{}
}

func (l *dummyLogger) Debugf(format string, args ...any) {
}

func (l *dummyLogger) Infof(format string, args ...any) {
}

func (l *dummyLogger) Warnf(format string, args ...any) {
}

func (l *dummyLogger) Errorf(format string, args ...any) {}
