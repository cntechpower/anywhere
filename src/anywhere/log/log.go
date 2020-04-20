package log

import (
	"fmt"
	_log "log"
	"os"
	"path"
	"runtime"
	"time"
)

var l *Logger

func InitLogger(fileName string) {

	if l != nil {
		panic("Logger already init")
	}
	l = &Logger{l: &_log.Logger{}}
	if fileName != "" {
		// You could set this to any `io.Writer` such as a file
		file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			l.l.SetOutput(file)
		} else {
			l.l.SetOutput(os.Stderr)
			l.l.Printf("Failed to log to file, using default stderr: %v\n", err)
		}
	} else {
		l.l.SetOutput(os.Stderr)
		l.l.Println("log to default stderr output")
	}

}

type Logger struct {
	l *_log.Logger
}

type Level string

const (
	Info  Level = "INFO"
	Warn  Level = "WARN"
	Error Level = "ERROR"
	Fatal Level = "FATAL"
)

func getCaller() (string, int) {
	_, file, line, ok := runtime.Caller(3)
	if ok {
		return path.Base(file), line
	}
	return "unknown.go", 0
}

func (l *Logger) log(level Level, format string, a ...interface{}) {
	file, line := getCaller()
	l.l.Println(fmt.Sprintf("[%s] <%s> (%s:%v) %s", time.Now().Format(time.RFC3339), level, file, line, fmt.Sprintf(format, a...)))
}

func (l *Logger) Infof(format string, a ...interface{}) {
	l.log(Info, format, a...)
}

func (l *Logger) Errorf(format string, a ...interface{}) {
	l.log(Error, format, a...)
}

func (l *Logger) Warnf(format string, a ...interface{}) {
	l.log(Warn, format, a...)
}

func (l *Logger) Fatalf(format string, a ...interface{}) {
	l.log(Error, format, a...)
}

func Infof(format string, a ...interface{}) {
	l.Infof(format, a...)
}

func Errorf(format string, a ...interface{}) {
	l.Errorf(format, a...)
}

func Warnf(format string, a ...interface{}) {
	l.Warnf(format, a...)
}

func Fatalf(format string, a ...interface{}) {
	l.Fatalf(format, a...)
}
func GetDefaultLogger() *Logger {
	return l
}
