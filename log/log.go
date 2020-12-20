package log

import (
	"fmt"
	_log "log"
	"os"
	"path"
	"runtime"
	"time"

	"github.com/cntechpower/anywhere/constants"
)

var globalLogger *_log.Logger

func InitLogger(fileName string) {
	if globalLogger != nil {
		panic("Logger already init")
	}
	globalLogger = &_log.Logger{}
	if fileName != "" {
		// You could set this to any `io.Writer` such as a file
		file, err := os.OpenFile("anywhere.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			globalLogger.SetOutput(file)
		} else {
			globalLogger.SetOutput(os.Stderr)
			globalLogger.Printf("Failed to log to file, using default stderr: %v\n", err)
		}
	} else {
		globalLogger.SetOutput(os.Stderr)
		//globalLogger.Println("log to default stderr output")
	}
}

type Level string

const (
	levelInfo  Level = "INFO"
	levelWarn  Level = "WARN"
	levelError Level = "ERROR"
	levelFatal Level = "FATAL"
)

func getCaller(skip int) (string, int) {
	_, file, line, ok := runtime.Caller(skip)
	if ok {
		return path.Base(file), line
	}
	return "unknown.go", 0
}

func log(skip int, h *Header, level Level, format string, a ...interface{}) {
	file, line := getCaller(skip)

	globalLogger.Println(fmt.Sprintf("[%s] <%s> |%s| (%s:%v) %s", time.Now().Format(constants.DefaultTimeFormat), level, h, file, line, fmt.Sprintf(format, a...)))
}

type Header struct {
	name string
}

func NewHeader(n string) *Header {
	if globalLogger == nil {
		panic("global logger not init")
	}
	return &Header{name: n}
}

func (h *Header) String() string {
	return h.name
}

type Logger interface {
	// Info logs routine messages about cron's operation.
	Info(msg string, keysAndValues ...interface{})
	// Error logs an error condition.
	Error(err error, msg string, keysAndValues ...interface{})
}

func (h *Header) Info(format string, a ...interface{}) {
	log(3, h, levelInfo, format, a...)
}
func (h *Header) Infof(format string, a ...interface{}) {
	log(3, h, levelInfo, format, a...)
}

func (h *Header) Errorf(format string, a ...interface{}) {
	log(3, h, levelError, format, a...)
}

func (h *Header) Error(err error, format string, a ...interface{}) {
	log(3, h, levelError, "%v", err)
	log(3, h, levelError, format, a...)
}

func (h *Header) Warnf(format string, a ...interface{}) {
	log(3, h, levelWarn, format, a...)
}

func (h *Header) Fatalf(format string, a ...interface{}) {
	log(3, h, levelFatal, format, a...)
	panic(nil)
}

func Infof(h *Header, format string, a ...interface{}) {
	log(3, h, levelInfo, format, a...)
}

func Errorf(h *Header, format string, a ...interface{}) {
	log(3, h, levelError, format, a...)
}

func Warnf(h *Header, format string, a ...interface{}) {
	log(3, h, levelWarn, format, a...)
}

func Fatalf(h *Header, format string, a ...interface{}) {
	log(3, h, levelError, format, a...)
}
