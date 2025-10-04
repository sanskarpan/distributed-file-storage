package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	
	if cfg.ListenAddr != ":3000" {
		t.Errorf("Expected default listen addr :3000, got %s", cfg.ListenAddr)
	}
	
	if cfg.StorageRoot != "storage" {
		t.Errorf("Expected default storage root 'storage', got %s", cfg.StorageRoot)
	}
	
	if cfg.LogLevel != "INFO" {
		t.Errorf("Expected default log level INFO, got %s", cfg.LogLevel)
	}
	
	if !cfg.EncryptionEnabled {
		t.Error("Expected encryption to be enabled by default")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name:        "valid config",
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name: "empty listen addr",
			config: &Config{
				ListenAddr:     "",
				StorageRoot:    "storage",
				LogLevel:       "INFO",
				MaxConnections: 10,
				ReadTimeout:    30,
				WriteTimeout:   30,
				MaxStorageSize: 1000,
				ReplicationFactor: 1,
			},
			expectError: true,
		},
		{
			name: "invalid log level",
			config: &Config{
				ListenAddr:     ":3000",
				StorageRoot:    "storage",
				LogLevel:       "INVALID",
				MaxConnections: 10,
				ReadTimeout:    30,
				WriteTimeout:   30,
				MaxStorageSize: 1000,
				ReplicationFactor: 1,
			},
			expectError: true,
		},
		{
			name: "negative max connections",
			config: &Config{
				ListenAddr:     ":3000",
				StorageRoot:    "storage",
				LogLevel:       "INFO",
				MaxConnections: -1,
				ReadTimeout:    30,
				WriteTimeout:   30,
				MaxStorageSize: 1000,
				ReplicationFactor: 1,
			},
			expectError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.Validate()
			if test.expectError && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !test.expectError && err != nil {
				t.Errorf("Expected no validation error but got: %v", err)
			}
		})
	}
}

func TestConfigLoadFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("FS_LISTEN_ADDR", ":8080")
	os.Setenv("FS_STORAGE_ROOT", "/tmp/storage")
	os.Setenv("FS_LOG_LEVEL", "DEBUG")
	defer func() {
		os.Unsetenv("FS_LISTEN_ADDR")
		os.Unsetenv("FS_STORAGE_ROOT")
		os.Unsetenv("FS_LOG_LEVEL")
	}()

	cfg := DefaultConfig()
	cfg.LoadFromEnv()

	if cfg.ListenAddr != ":8080" {
		t.Errorf("Expected listen addr :8080, got %s", cfg.ListenAddr)
	}
	
	if cfg.StorageRoot != "/tmp/storage" {
		t.Errorf("Expected storage root /tmp/storage, got %s", cfg.StorageRoot)
	}
	
	if cfg.LogLevel != "DEBUG" {
		t.Errorf("Expected log level DEBUG, got %s", cfg.LogLevel)
	}
}

func TestConfigSaveAndLoad(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ListenAddr = ":9000"
	cfg.LogLevel = "DEBUG"

	// Save to temporary file
	tmpFile := "/tmp/test_config.json"
	defer os.Remove(tmpFile)

	err := cfg.SaveToFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load from file
	loadedCfg, err := LoadFromFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.ListenAddr != ":9000" {
		t.Errorf("Expected listen addr :9000, got %s", loadedCfg.ListenAddr)
	}
	
	if loadedCfg.LogLevel != "DEBUG" {
		t.Errorf("Expected log level DEBUG, got %s", loadedCfg.LogLevel)
	}
}
