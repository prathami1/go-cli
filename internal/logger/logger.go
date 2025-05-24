package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// InitLogger initializes the global logger with appropriate settings
func InitLogger(verbose bool) {
	Log = logrus.New()
	Log.SetOutput(os.Stdout)

	if verbose {
		Log.SetLevel(logrus.DebugLevel)
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	} else {
		Log.SetLevel(logrus.InfoLevel)
		Log.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp: true,
		})
	}
}

// Info logs an info message
func Info(args ...interface{}) {
	if Log != nil {
		Log.Info(args...)
	}
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	if Log != nil {
		Log.Infof(format, args...)
	}
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	if Log != nil {
		Log.Debug(args...)
	}
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	if Log != nil {
		Log.Debugf(format, args...)
	}
}

// Error logs an error message
func Error(args ...interface{}) {
	if Log != nil {
		Log.Error(args...)
	}
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	if Log != nil {
		Log.Errorf(format, args...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	if Log != nil {
		Log.Fatal(args...)
	}
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	if Log != nil {
		Log.Fatalf(format, args...)
	}
}
