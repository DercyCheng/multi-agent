package logger

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/multi-agent/api-gateway/internal/config"
)

var log *logrus.Logger

// Init initializes the logger
func Init(cfg config.LoggingConfig) {
	log = logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	// Set output
	if cfg.Output == "stdout" || cfg.Output == "" {
		log.SetOutput(os.Stdout)
	} else {
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.SetOutput(os.Stdout)
			log.Warnf("Failed to open log file %s, using stdout", cfg.Output)
		} else {
			log.SetOutput(file)
		}
	}

	// Set formatter
	if cfg.Format == "json" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...interface{}) {
	if log == nil {
		return
	}
	log.WithFields(parseFields(fields...)).Debug(msg)
}

// Info logs an info message
func Info(msg string, fields ...interface{}) {
	if log == nil {
		return
	}
	log.WithFields(parseFields(fields...)).Info(msg)
}

// Warn logs a warning message
func Warn(msg string, fields ...interface{}) {
	if log == nil {
		return
	}
	log.WithFields(parseFields(fields...)).Warn(msg)
}

// Error logs an error message
func Error(msg string, fields ...interface{}) {
	if log == nil {
		return
	}
	log.WithFields(parseFields(fields...)).Error(msg)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...interface{}) {
	if log == nil {
		return
	}
	log.WithFields(parseFields(fields...)).Fatal(msg)
}

// parseFields converts key-value pairs to logrus.Fields
func parseFields(fields ...interface{}) logrus.Fields {
	result := make(logrus.Fields)
	
	for i := 0; i < len(fields)-1; i += 2 {
		if key, ok := fields[i].(string); ok {
			result[key] = fields[i+1]
		}
	}
	
	return result
}