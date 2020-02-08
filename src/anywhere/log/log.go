package log

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strconv"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func InitLogger(fileName string) {

	if log != nil {
		panic("logger already init")
	}
	log = logrus.New()
	log.Formatter = &logrus.TextFormatter{
		ForceColors:      false,
		DisableColors:    false,
		DisableTimestamp: false,
		FullTimestamp:    true,
		TimestampFormat:  "2006-01-02 15:04:05",
		DisableSorting:   false,
		CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
			return path.Base(frame.Function) + "()", path.Base(frame.File) + ":" + strconv.Itoa(frame.Line)

		},
	}
	log.SetReportCaller(true)
	if fileName != "" {
		// You could set this to any `io.Writer` such as a file
		file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.Out = file
		} else {
			log.Infof("Failed to log to file, using default stderr: %v", err)
		}
	} else {
		log.Infof("log to default stderr output")
	}

}

func GetCustomLogger(format string, a ...interface{}) *logrus.Entry {
	s := fmt.Sprintf(format, a...)
	return log.WithField("caller", s)
}

func GetDefaultLogger() *logrus.Logger {
	return log
}
