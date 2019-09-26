package log

import (
	"log"
	"os"
)

type Logger interface {
	logInfo(msg string, args ...interface{})
	logError(msg string, args ...interface{})
	logFatal(msg string, args ...interface{})
}

func Info(msg string, args ...interface{}) {
	getLogger().logInfo(msg, args)
}

func Error(msg string, args ...interface{}) {
	getLogger().logError(msg, args)
}

func Fatal(msg string, args ...interface{}) {
	getLogger().logFatal(msg, args)
}

var logger Logger

func getLogger() Logger {
	return logger
}

type stdLogger struct {
	logger *log.Logger
}

func InitStdLogger() {
	l := log.New(os.Stdout, "[anywhere]", log.LstdFlags)
	logger = &stdLogger{
		logger: l,
	}
}

func (l *stdLogger) logInfo(msg string, args ...interface{}) {
	l.logger.Printf(msg, args...)

}

func (l *stdLogger) logError(msg string, args ...interface{}) {
	l.logger.Printf(msg, args...)

}

func (l *stdLogger) logFatal(msg string, args ...interface{}) {
	l.logger.Fatalf(msg, args...)

}
