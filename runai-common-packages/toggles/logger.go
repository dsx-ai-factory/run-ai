package toggles

// Logger is the interface for logging within the toggles package.
// SetLogger should be called before Init, otherwise a no-op logger is used.
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type dummyLogger struct{}

func (d *dummyLogger) Debugf(format string, args ...interface{}) {}
func (d *dummyLogger) Infof(format string, args ...interface{})  {}
func (d *dummyLogger) Warnf(format string, args ...interface{})  {}
func (d *dummyLogger) Errorf(format string, args ...interface{}) {}
