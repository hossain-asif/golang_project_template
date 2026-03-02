package logger

import (
	"context"
	"go_project_structure/utils/enums"

	"github.com/sirupsen/logrus"
)

type ScopeLogger struct {
	layer     string
	module    string
	component string
	method    string
	ctx       context.Context
}

// Scope returns a logrus.Entry with pre-defined fields for the layer, module, and component.
func (l *LoggerWrapper) Scope(layer, module, component string) *ScopeLogger {
	return &ScopeLogger{
		layer:     layer,
		module:    module,
		component: component,
	}
}

// WithContext returns a new ScopeLogger that will auto-inject reqId and userSlug
func (s *ScopeLogger) WithContext(ctx context.Context) *ScopeLogger {
	return &ScopeLogger{
		layer:     s.layer,
		module:    s.module,
		component: s.component,
		method:    s.method,
		ctx:       ctx,
	}
}

func (s *ScopeLogger) fields() logrus.Fields {
	fields := logrus.Fields{
		"module":    s.module,
		"component": s.component,
	}

	if s.layer != "" {
		fields["layer"] = s.layer
	}

	if s.method != "" {
		fields["method"] = s.method
	}

	if s.ctx != nil {
		if reqId, ok := s.ctx.Value(enums.CtxRequestID).(string); ok && reqId != "" {
			fields["reqId"] = reqId
		}
		if userSlug, ok := s.ctx.Value(enums.CtxUserSlug).(string); ok && userSlug != "" {
			fields["user_slug"] = userSlug
		}
	}
	return fields
}

// func (s *ScopeLogger) Method(method string) *logrus.Entry {

// 	f := s.fields()
// 	f["method"] = method
// 	return Log.Logger.WithFields(f)
// }

func (s *ScopeLogger) Method(method string) *ScopeLogger {
    return &ScopeLogger{
        layer:     s.layer,
        module:    s.module,
        component: s.component,
        method:    method,
        ctx:       s.ctx,
    }
}

func (s *ScopeLogger) Info(msg string) {
	Log.Logger.WithFields(s.fields()).Info(msg)
}

func (s *ScopeLogger) Infof(msg string, args ...interface{}) {

	Log.Logger.WithFields(s.fields()).Infof(msg, args...)
}

func (s *ScopeLogger) Error(msg string) {
	Log.Logger.WithFields(s.fields()).Error(msg)
}

func (s *ScopeLogger) Errorf(msg string, args ...interface{}) {
	Log.Logger.WithFields(s.fields()).Errorf(msg, args...)
}

func (s *ScopeLogger) Warn(msg string) {
	Log.Logger.WithFields(s.fields()).Warn(msg)
}

func (s *ScopeLogger) Warnf(msg string, args ...interface{}) {
	Log.Logger.WithFields(s.fields()).Warnf(msg, args...)
}

func (s *ScopeLogger) Debug(msg string) {
	Log.Logger.WithFields(s.fields()).Debug(msg)
}

func (s *ScopeLogger) Debugf(msg string, args ...interface{}) {
	Log.Logger.WithFields(s.fields()).Debugf(msg, args...)
}

func (s *ScopeLogger) Fatal(msg string) {
	Log.Logger.WithFields(s.fields()).Fatal(msg)
}

func (s *ScopeLogger) Fatalf(msg string, args ...interface{}) {
	Log.Logger.WithFields(s.fields()).Fatalf(msg, args...)
}

func (s *ScopeLogger) Trace(msg string) {
	Log.Logger.WithFields(s.fields()).Trace(msg)
}

func (s *ScopeLogger) Tracef(msg string, args ...interface{}) {
	Log.Logger.WithFields(s.fields()).Tracef(msg, args...)
}

func (s *ScopeLogger) WithFields(fields map[string]interface{}) *logrus.Entry {
	return Log.Logger.WithFields(s.fields()).WithFields(fields)
}

func (s *ScopeLogger) WithField(key string, value interface{}) *logrus.Entry {
	return Log.Logger.WithFields(s.fields()).WithField(key, value)
}

func (s *ScopeLogger) WithError(err error) *logrus.Entry {
	return Log.Logger.WithFields(s.fields()).WithError(err)
}
