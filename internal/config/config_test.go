package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	configContent := `
server:
  port: 8080
  read_timeout: 10s
  write_timeout: 10s

database:
  host: localhost
  port: 5432
  user: test
  password: test
  name: testdb
  ssl_mode: disable
  max_conns: 10

logging:
  level: info
  file: /tmp/app.log
  max_size: 100
  max_backups: 3
  max_age: 28
`

	tmpFile, err := os.CreateTemp("", "config_test_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, 10*time.Second, cfg.Server.ReadTimeout)
	assert.Equal(t, 10*time.Second, cfg.Server.WriteTimeout)

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "test", cfg.Database.User)
	assert.Equal(t, "test", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.Equal(t, "disable", cfg.Database.SSLMode)
	assert.Equal(t, int32(10), cfg.Database.MaxConns)

	assert.Equal(t, "info", cfg.Logging.Level)
	assert.Equal(t, "/tmp/app.log", cfg.Logging.File)
	assert.Equal(t, 100, cfg.Logging.MaxSize)
	assert.Equal(t, 3, cfg.Logging.MaxBackups)
	assert.Equal(t, 28, cfg.Logging.MaxAge)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	cfg, err := LoadConfig("nonexistent.yaml")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "invalid_config_*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString("invalid: yaml: content:")
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	assert.Error(t, err)
	assert.Nil(t, cfg)
}
