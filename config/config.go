package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/anthdm/foreverstore/logger"
)

// Config holds all configuration for the file server
type Config struct {
	// Server configuration
	ListenAddr    string   `json:"listen_addr"`
	StorageRoot   string   `json:"storage_root"`
	BootstrapNodes []string `json:"bootstrap_nodes"`
	
	// Logging configuration
	LogLevel string `json:"log_level"`
	LogFile  string `json:"log_file"`
	
	// Security configuration
	EncryptionEnabled bool   `json:"encryption_enabled"`
	EncryptionKey     string `json:"encryption_key"`
	
	// Performance configuration
	MaxConnections    int `json:"max_connections"`
	ReadTimeout       int `json:"read_timeout_seconds"`
	WriteTimeout      int `json:"write_timeout_seconds"`
	
	// Storage configuration
	MaxStorageSize    int64 `json:"max_storage_size_bytes"`
	ReplicationFactor int   `json:"replication_factor"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		ListenAddr:        ":3000",
		StorageRoot:       "storage",
		BootstrapNodes:    []string{},
		LogLevel:          "INFO",
		LogFile:           "",
		EncryptionEnabled: true,
		EncryptionKey:     "",
		MaxConnections:    100,
		ReadTimeout:       30,
		WriteTimeout:      30,
		MaxStorageSize:    1024 * 1024 * 1024, // 1GB
		ReplicationFactor: 2,
	}
}

// LoadFromFile loads configuration from a JSON file
func LoadFromFile(filename string) (*Config, error) {
	config := DefaultConfig()
	
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Warn("Config file %s not found, using defaults", filename)
			return config, nil
		}
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()
	
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}
	
	logger.Info("Loaded configuration from %s", filename)
	return config, nil
}

// SaveToFile saves the configuration to a JSON file
func (c *Config) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(c); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}
	
	logger.Info("Saved configuration to %s", filename)
	return nil
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv() {
	if val := os.Getenv("FS_LISTEN_ADDR"); val != "" {
		c.ListenAddr = val
	}
	if val := os.Getenv("FS_STORAGE_ROOT"); val != "" {
		c.StorageRoot = val
	}
	if val := os.Getenv("FS_BOOTSTRAP_NODES"); val != "" {
		c.BootstrapNodes = strings.Split(val, ",")
	}
	if val := os.Getenv("FS_LOG_LEVEL"); val != "" {
		c.LogLevel = val
	}
	if val := os.Getenv("FS_LOG_FILE"); val != "" {
		c.LogFile = val
	}
	if val := os.Getenv("FS_ENCRYPTION_ENABLED"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			c.EncryptionEnabled = enabled
		}
	}
	if val := os.Getenv("FS_ENCRYPTION_KEY"); val != "" {
		c.EncryptionKey = val
	}
	if val := os.Getenv("FS_MAX_CONNECTIONS"); val != "" {
		if maxConn, err := strconv.Atoi(val); err == nil {
			c.MaxConnections = maxConn
		}
	}
	if val := os.Getenv("FS_READ_TIMEOUT"); val != "" {
		if timeout, err := strconv.Atoi(val); err == nil {
			c.ReadTimeout = timeout
		}
	}
	if val := os.Getenv("FS_WRITE_TIMEOUT"); val != "" {
		if timeout, err := strconv.Atoi(val); err == nil {
			c.WriteTimeout = timeout
		}
	}
	if val := os.Getenv("FS_MAX_STORAGE_SIZE"); val != "" {
		if size, err := strconv.ParseInt(val, 10, 64); err == nil {
			c.MaxStorageSize = size
		}
	}
	if val := os.Getenv("FS_REPLICATION_FACTOR"); val != "" {
		if factor, err := strconv.Atoi(val); err == nil {
			c.ReplicationFactor = factor
		}
	}
}

// LoadFromFlags loads configuration from command line flags
func (c *Config) LoadFromFlags() {
	flag.StringVar(&c.ListenAddr, "listen", c.ListenAddr, "Address to listen on")
	flag.StringVar(&c.StorageRoot, "storage", c.StorageRoot, "Storage root directory")
	flag.StringVar(&c.LogLevel, "log-level", c.LogLevel, "Log level (DEBUG, INFO, WARN, ERROR, FATAL)")
	flag.StringVar(&c.LogFile, "log-file", c.LogFile, "Log file path (empty for stdout)")
	flag.BoolVar(&c.EncryptionEnabled, "encryption", c.EncryptionEnabled, "Enable encryption")
	flag.StringVar(&c.EncryptionKey, "encryption-key", c.EncryptionKey, "Encryption key (hex encoded)")
	flag.IntVar(&c.MaxConnections, "max-connections", c.MaxConnections, "Maximum number of connections")
	flag.IntVar(&c.ReadTimeout, "read-timeout", c.ReadTimeout, "Read timeout in seconds")
	flag.IntVar(&c.WriteTimeout, "write-timeout", c.WriteTimeout, "Write timeout in seconds")
	flag.Int64Var(&c.MaxStorageSize, "max-storage", c.MaxStorageSize, "Maximum storage size in bytes")
	flag.IntVar(&c.ReplicationFactor, "replication", c.ReplicationFactor, "Replication factor")
	
	// Custom flag for bootstrap nodes
	var bootstrapNodes string
	flag.StringVar(&bootstrapNodes, "bootstrap", "", "Comma-separated list of bootstrap nodes")
	
	flag.Parse()
	
	if bootstrapNodes != "" {
		c.BootstrapNodes = strings.Split(bootstrapNodes, ",")
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ListenAddr == "" {
		return fmt.Errorf("listen address cannot be empty")
	}
	
	if c.StorageRoot == "" {
		return fmt.Errorf("storage root cannot be empty")
	}
	
	validLogLevels := map[string]bool{
		"DEBUG": true,
		"INFO":  true,
		"WARN":  true,
		"ERROR": true,
		"FATAL": true,
	}
	
	if !validLogLevels[strings.ToUpper(c.LogLevel)] {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}
	
	if c.MaxConnections <= 0 {
		return fmt.Errorf("max connections must be positive")
	}
	
	if c.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}
	
	if c.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}
	
	if c.MaxStorageSize <= 0 {
		return fmt.Errorf("max storage size must be positive")
	}
	
	if c.ReplicationFactor <= 0 {
		return fmt.Errorf("replication factor must be positive")
	}
	
	return nil
}

// GetLogLevel returns the logger.LogLevel for the configured log level
func (c *Config) GetLogLevel() logger.LogLevel {
	switch strings.ToUpper(c.LogLevel) {
	case "DEBUG":
		return logger.DEBUG
	case "INFO":
		return logger.INFO
	case "WARN":
		return logger.WARN
	case "ERROR":
		return logger.ERROR
	case "FATAL":
		return logger.FATAL
	default:
		return logger.INFO
	}
}

// Load loads configuration from file, environment variables, and command line flags
func Load(configFile string) (*Config, error) {
	// Start with defaults
	config := DefaultConfig()
	
	// Load from file if specified
	if configFile != "" {
		fileConfig, err := LoadFromFile(configFile)
		if err != nil {
			return nil, err
		}
		config = fileConfig
	}
	
	// Override with environment variables
	config.LoadFromEnv()
	
	// Override with command line flags
	config.LoadFromFlags()
	
	// Validate the final configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	
	return config, nil
}
