// Package logger provides utilities for creating and using structured loggers with logrus and Gin middleware
package logger

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	mkdirPerm    = 0o700
	openFilePerm = 0o600
)

// NewLogger creates and configures a new logrus logger with specified level and optional file output
func NewLogger(level, logFile string) *logrus.Logger {
	log := logrus.New()

	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		lvl = logrus.InfoLevel
	}
	log.SetLevel(lvl)

	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	if logFile != "" {
		dir := filepath.Dir(logFile)
		if err := os.MkdirAll(dir, mkdirPerm); err != nil {
			log.Warnf("Failed to create log directory: %v", err)
		} else {
			// #nosec G304 -- log path is trusted and not user-controlled
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, openFilePerm)
			if err != nil {
				log.Warnf("Failed to open log file: %v", err)
			} else {
				log.SetOutput(io.MultiWriter(os.Stdout, file))
			}
		}
	}

	return log
}

// GinLogger returns a Gin middleware handler that logs HTTP requests using the provided logger
func GinLogger(log *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		entry := log.WithFields(logrus.Fields{
			"status":     statusCode,
			"method":     method,
			"path":       path,
			"ip":         clientIP,
			"latency":    latency,
			"user_agent": c.Request.UserAgent(),
			"body_size":  bodySize,
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.String())
		} else {
			msg := "Request completed"
			switch {
			case statusCode >= http.StatusInternalServerError:
				entry.Error(msg)
			case statusCode >= http.StatusBadRequest:
				entry.Warn(msg)
			default:
				entry.Info(msg)
			}
		}
	}
}
