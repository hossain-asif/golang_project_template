package logger

import (
	config "go_project_structure/config/env"
	"os"

	"github.com/sirupsen/logrus"
)

var Log *LoggerWrapper

func InitLogger() {
	cfg := LogConfig{
		Environment:  config.GetString("APP_ENV", "development"),
		Level:        config.GetString("LOG_LEVEL", "info"),
		LogDirectory: config.GetString("LOG_DIRECTORY", "logs"),
		LogFile:      os.Getenv("LOG_FILE"),
	}

	log := NewLogConfig(cfg)

	Log = &LoggerWrapper{Logger: log}

	log.WithFields(map[string]interface{}{
		"module":    "logger",
		"component": "log_setup",
	}).Info("Logger initialized")

}

func AddHook(hook logrus.Hook) {
	// add mongodb hook
	Log.Logger.AddHook(hook)

}


