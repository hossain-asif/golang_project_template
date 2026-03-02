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
		"module":    "logger",
		"component": "log_setup",
	}).Info("Logger initialized")

}

func AddHook(hook logrus.Hook) {
	// add mongodb hook
	Log.Logger.AddHook(hook)

	// add method hook
	// Log.Logger.AddHook(&MethodHook{})
}

type MethodHook struct{}

func (m *MethodHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// func (m *MethodHook) Fire(entry *logrus.Entry) error {
// 	// 6 is the typical depth for LogWrapper structure
// 	pc, _, _, ok := runtime.Caller(6)
// 	if ok {
// 		fn := runtime.FuncForPC(pc)
// 		if fn != nil {
// 			parts := strings.Split(fn.Name(), ".")
// 			entry.Data["method"] = parts[len(parts)-1]

// 		}
// 	}
// 	return nil
// }

// func (m *MethodHook) Fire(entry *logrus.Entry) error {

// 	for i := 1; i <= 15; i++ {
// 		pc, _, _, ok := runtime.Caller(i)
// 			if !ok { break }
// 			fn := runtime.FuncForPC(pc)
// 			if fn == nil { continue }

// 		name := fn.Name()

// 		// Skip logrus and your internal logger wrapper
// 		if !strings.Contains(name, "sirupsen/logrus") &&
// 			!strings.Contains(name, "config/logger") {
// 			parts := strings.Split(name, ".")
// 			entry.Data["method"] = parts[len(parts)-1]
// 			break
// 		}
// 	}

// 	return nil
// }
