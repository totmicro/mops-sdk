package sdk

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger provides a unified logging interface for plugins
type Logger struct {
	*logrus.Logger
	pluginName string
	logFile    *os.File
}

// LogLevel represents the logging level
type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

var (
	// Global logger instance for the plugin
	pluginLogger *Logger
)

// InitPluginLogger initializes the logger for a plugin
func InitPluginLogger(pluginName string) (*Logger, error) {
	if pluginLogger != nil {
		return pluginLogger, nil
	}

	logger := &Logger{
		Logger:     logrus.New(),
		pluginName: pluginName,
	}

	// Set up log file
	if err := logger.setupLogFile(); err != nil {
		return nil, fmt.Errorf("failed to setup log file: %w", err)
	}

	// Configure logger format and level
	logger.setupFormat()
	logger.setupLevel()

	pluginLogger = logger
	return logger, nil
}

// GetPluginLogger returns the global plugin logger instance
func GetPluginLogger() *Logger {
	return pluginLogger
}

// setupLogFile creates and opens the log file for the plugin
func (l *Logger) setupLogFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	logsDir := filepath.Join(homeDir, ".mops", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Create log file with pattern: pluginname-YYYY-MM-DD.log
	today := time.Now().Format("2006-01-02")
	logFileName := fmt.Sprintf("%s-%s.log", l.pluginName, today)
	logFilePath := filepath.Join(logsDir, logFileName)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", logFilePath, err)
	}

	l.logFile = logFile

	// Set output to both file and stderr for development
	l.SetOutput(io.MultiWriter(logFile, os.Stderr))

	return nil
}

// setupFormat configures the log format
func (l *Logger) setupFormat() {
	l.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		ForceColors:     false, // Disable colors for file output
	})
}

// setupLevel configures the log level based on environment variables
func (l *Logger) setupLevel() {
	// Check for plugin-specific log level first
	pluginLogLevel := os.Getenv(fmt.Sprintf("%s_LOG_LEVEL", strings.ToUpper(l.pluginName)))
	if pluginLogLevel == "" {
		// Fall back to MOPS global log level
		pluginLogLevel = os.Getenv("MOPS_LOG_LEVEL")
	}
	if pluginLogLevel == "" {
		// Default to info level
		pluginLogLevel = "info"
	}

	switch strings.ToLower(pluginLogLevel) {
	case "debug":
		l.SetLevel(logrus.DebugLevel)
	case "info":
		l.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		l.SetLevel(logrus.WarnLevel)
	case "error":
		l.SetLevel(logrus.ErrorLevel)
	default:
		l.SetLevel(logrus.InfoLevel)
	}
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// WithComponent adds a component field to log entries
func (l *Logger) WithComponent(component string) *logrus.Entry {
	return l.WithField("component", component)
}

// WithFields adds multiple fields to log entries
func (l *Logger) WithFields(fields map[string]interface{}) *logrus.Entry {
	return l.Logger.WithFields(logrus.Fields(fields))
}

// WithError adds an error field to log entries
func (l *Logger) WithError(err error) *logrus.Entry {
	return l.Logger.WithError(err)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.Logger.Debug(msg)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.Logger.Debugf(format, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.Logger.Info(msg)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.Logger.Infof(format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.Logger.Warn(msg)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.Logger.Warnf(format, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.Logger.Error(msg)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.Logger.Errorf(format, args...)
}

// Convenience functions for quick logging without getting logger instance

// LogDebug logs a debug message using the global plugin logger
func LogDebug(component, msg string) {
	if pluginLogger != nil {
		pluginLogger.WithComponent(component).Debug(msg)
	}
}

// LogDebugf logs a formatted debug message using the global plugin logger
func LogDebugf(component, format string, args ...interface{}) {
	if pluginLogger != nil {
		pluginLogger.WithComponent(component).Debugf(format, args...)
	}
}

// LogInfo logs an info message using the global plugin logger
func LogInfo(component, msg string) {
	if pluginLogger != nil {
		pluginLogger.WithComponent(component).Info(msg)
	}
}

// LogInfof logs a formatted info message using the global plugin logger
func LogInfof(component, format string, args ...interface{}) {
	if pluginLogger != nil {
		pluginLogger.WithComponent(component).Infof(format, args...)
	}
}

// LogWarn logs a warning message using the global plugin logger
func LogWarn(component, msg string) {
	if pluginLogger != nil {
		pluginLogger.WithComponent(component).Warn(msg)
	}
}

// LogWarnf logs a formatted warning message using the global plugin logger
func LogWarnf(component, format string, args ...interface{}) {
	if pluginLogger != nil {
		pluginLogger.WithComponent(component).Warnf(format, args...)
	}
}

// LogError logs an error message using the global plugin logger
func LogError(component, msg string) {
	if pluginLogger != nil {
		pluginLogger.WithComponent(component).Error(msg)
	}
}

// LogErrorf logs a formatted error message using the global plugin logger
func LogErrorf(component, format string, args ...interface{}) {
	if pluginLogger != nil {
		pluginLogger.WithComponent(component).Errorf(format, args...)
	}
}

// LogWithFields logs a message with custom fields using the global plugin logger
func LogWithFields(level LogLevel, component string, msg string, fields map[string]interface{}) {
	if pluginLogger == nil {
		return
	}

	entry := pluginLogger.WithComponent(component).WithFields(logrus.Fields(fields))
	
	switch level {
	case DebugLevel:
		entry.Debug(msg)
	case InfoLevel:
		entry.Info(msg)
	case WarnLevel:
		entry.Warn(msg)
	case ErrorLevel:
		entry.Error(msg)
	}
}
