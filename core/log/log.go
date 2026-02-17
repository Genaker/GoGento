package log

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// LogLevel represents the severity of the log message
type LogLevel int

const (
	INFO LogLevel = iota
	WARN
	ERROR
	FATAL
)

var (
	logFile *os.File
	logger  *log.Logger
	once    sync.Once
)

// Init initializes the logger and opens the log file
func Init() {
	once.Do(func() {
		var err error
		logFile, err = os.OpenFile("var/log/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %v", err)
		}
		logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
	})
}

// Close closes the log file
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

func logWithLevel(level LogLevel, format string, v ...interface{}) {
	if logger == nil {
		Init()
	}
	prefix := "[INFO] "
	switch level {
	case WARN:
		prefix = "[WARN] "
	case ERROR:
		prefix = "[ERROR] "
	case FATAL:
		prefix = "[FATAL] "
	}
	logger.SetPrefix(prefix)
	msg := fmt.Sprintf(format, v...)
	// 3 skips: logWithLevel -> Info/Error/etc -> your app code
	if level == FATAL {
		logger.Output(3, msg) // Output exits after printing if FATAL
		os.Exit(1)
	} else {
		logger.Output(3, msg)
	}
}

func Info(format string, v ...interface{})  { logWithLevel(INFO, format, v...) }
func Warn(format string, v ...interface{})  { logWithLevel(WARN, format, v...) }
func Error(format string, v ...interface{}) { logWithLevel(ERROR, format, v...) }
func Fatal(format string, v ...interface{}) { logWithLevel(FATAL, format, v...) }
