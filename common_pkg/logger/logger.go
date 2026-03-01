package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConfig struct {
	Environment string
	Level       string
	LogFile     string
}

type OrderedFormatter struct {
	logrus.TextFormatter
	Order []string
}

func (f *OrderedFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(logrus.Fields)

	// add ordered fields first
	for _, key := range f.Order {
		if value, ok := entry.Data[key]; ok {
			data[key] = value
		}
	}

	// add any remaining fields
	for key, value := range entry.Data {
		if _, ok := data[key]; !ok {
			data[key] = value
		}
	}
	entry.Data = data
	return f.TextFormatter.Format(entry)
}

func NewLogConfig(cfg LogConfig) *logrus.Logger {
	log := logrus.New()

	level, err := logrus.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	if cfg.Environment == "production" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05Z07:00",
		})
	} else {
		log.SetFormatter(&OrderedFormatter{
			TextFormatter: logrus.TextFormatter{
				FullTimestamp:  true,
				DisableSorting: true,
			},
			Order: []string{"layer", "module", "component", "method"},
		})
	}

	if cfg.LogFile != "" {
		log.SetOutput(&lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    20, // MB
			MaxBackups: 5,
			MaxAge:     30, // days
			Compress:   true,
		})
	} else {
		// Recommended for Docker/K8s
		log.SetOutput(os.Stdout)
	}

	return log
}
