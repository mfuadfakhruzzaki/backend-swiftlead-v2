package logger

import (
	"log"
	"os"
)

// Level represents log level
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

// Logger provides structured logging
type Logger struct {
	level Level
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

// New creates a new logger
func New(level string) *Logger {
	l := &Logger{
		level: parseLevel(level),
		debug: log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
		info:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime),
		warn:  log.New(os.Stdout, "[WARN] ", log.Ldate|log.Ltime),
		error: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
	}
	return l
}

func parseLevel(level string) Level {
	switch level {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	default:
		return INFO
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= DEBUG {
		l.debug.Printf(format, v...)
	}
}

// Info logs an info message
func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= INFO {
		l.info.Printf(format, v...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= WARN {
		l.warn.Printf(format, v...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= ERROR {
		l.error.Printf(format, v...)
	}
}

// Fatal logs an error and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.error.Fatalf(format, v...)
}

// Default logger instance
var defaultLogger = New("info")

// SetDefault sets the default logger
func SetDefault(l *Logger) {
	defaultLogger = l
}

// Package-level logging functions
func Debug(format string, v ...interface{}) { defaultLogger.Debug(format, v...) }
func Info(format string, v ...interface{})  { defaultLogger.Info(format, v...) }
func Warn(format string, v ...interface{})  { defaultLogger.Warn(format, v...) }
func Error(format string, v ...interface{}) { defaultLogger.Error(format, v...) }
func Fatal(format string, v ...interface{}) { defaultLogger.Fatal(format, v...) }
