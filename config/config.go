package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	TimeZone string
}

// Load reads the configuration from environment variables and returns a Config struct.
func Load() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Database: DatabaseConfig{
			Host:     os.Getenv("SAMPAY_DB_HOST"),
			Port:     os.Getenv("SAMPAY_DB_PORT"),
			User:     os.Getenv("SAMPAY_DB_USER"),
			Password: os.Getenv("SAMPAY_DB_PASSWORD"),
			Name:     os.Getenv("SAMPAY_DB_NAME"),
			SSLMode:  os.Getenv("SAMPAY_DB_SSLMODE"),
			TimeZone: os.Getenv("SAMPAY_DB_TIMEZONE"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (cfg Config) validate() error {
	if cfg.Database.Host == "" {
		return errors.New("missing required configuration: SAMPAY_DB_HOST")
	}
	if cfg.Database.Port == "" {
		return errors.New("missing required configuration: SAMPAY_DB_PORT")
	}
	if cfg.Database.User == "" {
		return errors.New("missing required configuration: SAMPAY_DB_USER")
	}
	if cfg.Database.Password == "" {
		return errors.New("missing required configuration: SAMPAY_DB_PASSWORD")
	}
	if cfg.Database.Name == "" {
		return errors.New("missing required configuration: SAMPAY_DB_NAME")
	}
	if cfg.Database.SSLMode == "" {
		return errors.New("missing required configuration: SAMPAY_DB_SSLMODE")
	}
	if cfg.Database.TimeZone == "" {
		return errors.New("missing required configuration: SAMPAY_DB_TIMEZONE")
	}
	return nil
}