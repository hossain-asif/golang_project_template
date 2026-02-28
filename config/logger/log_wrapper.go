package logger

import (
	"github.com/sirupsen/logrus"
)

type LoggerWrapper struct {
	Logger *logrus.Logger
}

func (l *LoggerWrapper) Info(msg string) {
	l.Logger.Info(msg)
}

func (l *LoggerWrapper) Error(msg string) {
	l.Logger.Error(msg)
}

func (l *LoggerWrapper) Warn(msg string) {
	l.Logger.Warn(msg)
}

func (l *LoggerWrapper) Debug(msg string) {
	l.Logger.Debug(msg)
}

func (l *LoggerWrapper) Fatal(msg string) {
	l.Logger.Fatal(msg)
}

func (l *LoggerWrapper) Trace(msg string) {
	l.Logger.Trace(msg)
}

func (l *LoggerWrapper) WithFields(fields logrus.Fields) *logrus.Entry {
	return l.Logger.WithFields(fields)
}

// Scope returns a logrus.Entry with pre-defined fields for the layer, module, and component.
func (l *LoggerWrapper) Scope(layer, module, component, method string) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields{
		"layer":     layer,
		"module":    module,
		"component": component,
		"method": method,
	})
}