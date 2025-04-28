package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
)

// Config represents the application configuration
type Config struct {
	Server      ServerConfig      `json:"server"`
	Database    DatabaseConfig    `json:"database"`
	Logging     LoggingConfig     `json:"logging"`
	Application ApplicationConfig `json:"application"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// ApplicationConfig holds application-related configuration
type ApplicationConfig struct {
	Secret string `json:"secret"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Driver   string `json:"driver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

var (
	instance *Config
	once     sync.Once
)

// LoadConfig loads the application configuration
func LoadConfig(configFile string) (*Config, error) {
	var err error

	once.Do(func() {
		// Initialize default configuration
		instance = &Config{
			Server: ServerConfig{
				Host: "127.0.0.1",
				Port: 8080,
			},
			Database: DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "postgres",
				Name:     "jiaxun",
			},
			Logging: LoggingConfig{
				Level:  "info",
				Format: "text",
			},
			Application: ApplicationConfig{
				Secret: "mysecret",
			},
		}

		// Load from file if provided
		if configFile != "" {
			if err = loadFromFile(configFile, instance); err != nil {
				err = fmt.Errorf("loading config: %w", err)
				return
			}
		}

		// Override with environment variables
		overrideFromEnv(instance)
	})

	return instance, err
}

// GetConfig returns the current configuration
func GetConfig() *Config {
	if instance == nil {
		_, _ = LoadConfig("") // Load default configuration
	}
	return instance
}

// loadFromFile loads configuration from a JSON file
func loadFromFile(file string, cfg *Config) error {
	// Check if the file exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		// Create the default config file if it does not exist
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling default config: %w", err)
		}
		if err := os.WriteFile(file, data, 0644); err != nil {
			return fmt.Errorf("writing default config: %w", err)
		}
		return nil
	}

	// Read and parse the file
	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("parsing config file: %w", err)
	}

	return nil
}

// overrideFromEnv overrides configuration with environment variables
func overrideFromEnv(cfg *Config) {
	// Server configuration
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}

	// Database configuration
	if driver := os.Getenv("DB_DRIVER"); driver != "" {
		cfg.Database.Driver = driver
	}
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Database.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}
	if name := os.Getenv("DB_NAME"); name != "" {
		cfg.Database.Name = name
	}

	// Logging configuration
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Logging.Level = level
	}
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		cfg.Logging.Format = format
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("database password is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}
	return nil
}
