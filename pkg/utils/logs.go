package utils

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	defaultLogger *log.Logger
)

//DefaultLogger provides the main log instance
func DefaultLogger() *log.Logger {
	if defaultLogger != nil {
		return defaultLogger
	}
	defaultLogger = log.New()

	defaultLogger.SetOutput(os.Stdout)
	logLevel := viper.GetInt("LogLevel")
	if logLevel >= 5 {
		defaultLogger.SetLevel(log.TraceLevel)
	} else if logLevel >= 4 {
		defaultLogger.SetLevel(log.DebugLevel)
	} else if logLevel >= 3 {
		defaultLogger.SetLevel(log.WarnLevel)
	} else if logLevel >= 2 {
		defaultLogger.SetLevel(log.ErrorLevel)
	} else if logLevel >= 1 {
		defaultLogger.SetLevel(log.FatalLevel)
	} else if logLevel >= 0 {
		defaultLogger.SetLevel(log.PanicLevel)
	}
	if logLevel > 3 {
		defaultLogger.Infof("Log level set %v", defaultLogger.GetLevel())
	}

	return defaultLogger
}
