package logger

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger_DefaultLevel(t *testing.T) {
	log := NewLogger("invalid-level", "")
	assert.Equal(t, logrus.InfoLevel, log.Level)
}

func TestNewLogger_WithFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "app.log")

	log := NewLogger("debug", logFile)
	assert.Equal(t, logrus.DebugLevel, log.Level)

	log.Info("test log")

	info, err := os.Stat(logFile)
	assert.NoError(t, err)
	assert.False(t, info.IsDir())
}

func TestNewLogger_FailToCreateDir(t *testing.T) {
	badPath := "/root/forbidden-dir/app.log"
	log := NewLogger("info", badPath)
	assert.Equal(t, logrus.InfoLevel, log.Level)
}

func TestGinLogger_MiddlewareLogs(t *testing.T) {
	var buf bytes.Buffer
	log := logrus.New()
	log.SetOutput(&buf)
	log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339})

	r := gin.New()
	r.Use(GinLogger(log))

	r.GET("/success", func(c *gin.Context) {
		c.String(200, "ok")
	})
	r.GET("/badrequest", func(c *gin.Context) {
		c.String(400, "bad request")
	})
	r.GET("/servererror", func(c *gin.Context) {
		c.String(500, "server error")
	})

	tests := []struct {
		path           string
		expectedLevel  string
		expectedStatus int
	}{
		{"/success", "info", 200},
		{"/badrequest", "warning", 400},
		{"/servererror", "error", 500},
	}

	for _, tt := range tests {
		buf.Reset()
		req := httptest.NewRequest("GET", tt.path, http.NoBody)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		logOutput := buf.String()
		assert.Contains(t, logOutput, `"level":"`+tt.expectedLevel+`"`)
		assert.Contains(t, logOutput, `"status":`+strconv.Itoa(tt.expectedStatus))
		assert.Contains(t, logOutput, `"path":"`+tt.path+`"`)
		assert.Equal(t, tt.expectedStatus, w.Code)
	}
}

func TestGinLogger_LogsErrors(t *testing.T) {
	var buf bytes.Buffer
	log := logrus.New()
	log.SetOutput(&buf)
	log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339})

	r := gin.New()
	r.Use(GinLogger(log))

	r.GET("/error", func(c *gin.Context) {
		c.Error(&gin.Error{
			Err:  io.EOF,
			Type: gin.ErrorTypePrivate,
		})
		c.String(200, "ok with error")
	})

	req := httptest.NewRequest("GET", "/error", http.NoBody)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	logOutput := buf.String()
	assert.Contains(t, logOutput, `"level":"error"`)
	assert.Contains(t, logOutput, "EOF")
	assert.Equal(t, 200, w.Code)
}
