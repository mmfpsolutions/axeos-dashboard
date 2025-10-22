package logger

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Module represents the component that is logging
type Module string

const (
	ModuleMain       Module = "main"
	ModuleDatabase   Module = "database"
	ModuleScheduler  Module = "scheduler"
	ModuleConfig     Module = "config"
	ModuleHandler    Module = "handler"
	ModuleMiddleware Module = "middleware"
	ModuleService    Module = "service"
	ModuleAuth       Module = "auth"
)

// Logger provides structured logging functionality
type Logger struct {
	module Module
	logger *log.Logger
}

// New creates a new logger for the specified module
func New(module Module) *Logger {
	return &Logger{
		module: module,
		logger: log.New(os.Stdout, "", 0), // No prefix, we'll format ourselves
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// formatMessage formats the log message with standard format:
// [timestamp] [client_ip] [module] action
func (l *Logger) formatMessage(clientIP, action string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	if clientIP != "" {
		return fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, clientIP, l.module, action)
	}
	return fmt.Sprintf("[%s] [system] [%s] %s", timestamp, l.module, action)
}

// Info logs an informational message (system-level, no client IP)
func (l *Logger) Info(format string, args ...interface{}) {
	action := fmt.Sprintf(format, args...)
	msg := l.formatMessage("", action)
	l.logger.Println(msg)
}

// InfoWithRequest logs an informational message with client IP from request
func (l *Logger) InfoWithRequest(r *http.Request, format string, args ...interface{}) {
	action := fmt.Sprintf(format, args...)
	clientIP := getClientIP(r)
	msg := l.formatMessage(clientIP, action)
	l.logger.Println(msg)
}

// Error logs an error message (system-level, no client IP)
func (l *Logger) Error(format string, args ...interface{}) {
	action := fmt.Sprintf(format, args...)
	msg := l.formatMessage("", action)
	l.logger.Println(msg)
}

// ErrorWithRequest logs an error message with client IP from request
func (l *Logger) ErrorWithRequest(r *http.Request, format string, args ...interface{}) {
	action := fmt.Sprintf(format, args...)
	clientIP := getClientIP(r)
	msg := l.formatMessage(clientIP, action)
	l.logger.Println(msg)
}

// Fatal logs a fatal error and exits the program
func (l *Logger) Fatal(format string, args ...interface{}) {
	action := fmt.Sprintf(format, args...)
	msg := l.formatMessage("", action)
	l.logger.Fatal(msg)
}

// Warn logs a warning message (system-level, no client IP)
func (l *Logger) Warn(format string, args ...interface{}) {
	action := fmt.Sprintf(format, args...)
	msg := l.formatMessage("", action)
	l.logger.Println(msg)
}

// WarnWithRequest logs a warning message with client IP from request
func (l *Logger) WarnWithRequest(r *http.Request, format string, args ...interface{}) {
	action := fmt.Sprintf(format, args...)
	clientIP := getClientIP(r)
	msg := l.formatMessage(clientIP, action)
	l.logger.Println(msg)
}

// Debug logs a debug message (system-level, no client IP)
func (l *Logger) Debug(format string, args ...interface{}) {
	action := fmt.Sprintf(format, args...)
	msg := l.formatMessage("", action)
	l.logger.Println(msg)
}
