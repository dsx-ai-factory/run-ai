package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimezap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func InitLogger(logLevel int) {
	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zapcore.Level(logLevel)) //nolint:gosec
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapConfig.EncoderConfig.TimeKey = "timestamp"
	zapConfig.Encoding = "console"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := zapConfig.Build()
	zap.ReplaceGlobals(logger)

	// required for imported code from workload-controller
	opts := runtimezap.Options{
		Development:     true,
		StacktraceLevel: zapcore.LevelEnabler(zapcore.FatalLevel),
		TimeEncoder:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")}
	ctrl.SetLogger(runtimezap.New(runtimezap.UseFlagOptions(&opts)))
}
