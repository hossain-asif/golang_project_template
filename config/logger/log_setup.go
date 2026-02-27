package logger

import (
	config "go_project_structure/config/env"
	"os"

	"github.com/sirupsen/logrus"
)

var Log *LoggerWrapper

func InitLogger() {
	cfg := LogConfig{
		Environment: config.GetString("APP_ENV", "development"),
		Level:       config.GetString("LOG_LEVEL", "info"),
		LogFile:     os.Getenv("LOG_FILE"),
	}

	log := NewLogConfig(cfg)

	Log = &LoggerWrapper{Logger: log}

	log.WithFields(map[string]interface{}{
		"module": "logger",
		"step":   "log_setup",
	}).Info("Logger initialized")

}

func AddHook(hook logrus.Hook) {
	Log.Logger.AddHook(hook)
}
