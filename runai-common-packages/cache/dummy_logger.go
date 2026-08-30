package cache

type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type dummyLogger struct{}

func (*dummyLogger) Debugf(format string, args ...any) {}
func (*dummyLogger) Infof(format string, args ...any)  {}
func (*dummyLogger) Warnf(format string, args ...any)  {}
func (*dummyLogger) Errorf(format string, args ...any) {}
