package temporal

import (
	"go.temporal.io/sdk/log"
	"go.uber.org/zap"
)

// TemporalLogger wraps zap.Logger to implement Temporal's Logger interface
type TemporalLogger struct {
	logger *zap.Logger
}

// NewTemporalLogger creates a new Temporal logger wrapper
func NewTemporalLogger(logger *zap.Logger) log.Logger {
	return &TemporalLogger{
		logger: logger,
	}
}

// Debug logs a debug message
func (l *TemporalLogger) Debug(msg string, keyvals ...interface{}) {
	l.logger.Debug(msg, l.convertKeyvals(keyvals...)...)
}

// Info logs an info message
func (l *TemporalLogger) Info(msg string, keyvals ...interface{}) {
	l.logger.Info(msg, l.convertKeyvals(keyvals...)...)
}

// Warn logs a warning message
func (l *TemporalLogger) Warn(msg string, keyvals ...interface{}) {
	l.logger.Warn(msg, l.convertKeyvals(keyvals...)...)
}

// Error logs an error message
func (l *TemporalLogger) Error(msg string, keyvals ...interface{}) {
	l.logger.Error(msg, l.convertKeyvals(keyvals...)...)
}

// convertKeyvals converts key-value pairs to zap fields
func (l *TemporalLogger) convertKeyvals(keyvals ...interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(keyvals)/2)
	
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			key := keyvals[i]
			value := keyvals[i+1]
			
			if keyStr, ok := key.(string); ok {
				fields = append(fields, zap.Any(keyStr, value))
			}
		}
	}
	
	return fields
}