package log

import (
	"go.uber.org/zap"
)

var (
	log *Logger
)

// Logger provides logging methods.
type Logger struct {
	*zap.SugaredLogger
	// fields map[string]interface{}
}

func init() {
	// logger, _ := zap.NewProduction()
	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // flushes buffer, if any
	logger = logger.WithOptions(zap.AddCallerSkip(1))
	log = &Logger{logger.Sugar()}
}

// Debugf logs at debug level.
func Debugf(s string, args ...interface{}) {
	log.Debugf(s, args...)
}

// Infof logs at info level.
func Infof(s string, args ...interface{}) {
	log.Infof(s, args...)
}

// Warnf logs at warning level.
func Warnf(s string, args ...interface{}) {
	log.Warnf(s, args...)
}

// Error logs at error level.
func Error(args ...interface{}) {
	log.Error(args...)
}

// Errorf logs at error level.
func Errorf(s string, args ...interface{}) {
	log.Errorf(s, args...)
}

// Fatalf logs at error level and panics.
func Fatalf(s string, args ...interface{}) {
	log.Fatalf(s, args...)
}

// Fatal logs at error level and panics.
func Fatal(arg interface{}) {
	log.Fatal(arg)
}
