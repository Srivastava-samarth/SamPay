package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	SMTP     SMTPConfig
	Temporal TemporalConfig
	JWT      JWTConfig
}

type SMTPConfig struct {
	Host string
	Port string
	From string
}

type TemporalConfig struct {
	Host string
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

type JWTConfig struct {
	Secret        string
	AccessExpiry  string
	RefreshExpiry string
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
		SMTP: SMTPConfig{
			Host: os.Getenv("SMTP_HOST"),
			Port: os.Getenv("SMTP_PORT"),
			From: os.Getenv("SMTP_FROM"),
		},
		Temporal: TemporalConfig{
			Host: os.Getenv("SAMPAY_TEMPORAL_HOST"),
		},
		JWT: JWTConfig{
			Secret:        os.Getenv("JWT_SECRET"),
			AccessExpiry:  os.Getenv("JWT_ACCESS_EXPIRY"),
			RefreshExpiry: os.Getenv("JWT_REFRESH_EXPIRY"),
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
	if cfg.SMTP.Host == "" {
		return errors.New("missing required configuration: SAMPAY_SMTP_HOST")
	}
	if cfg.SMTP.Port == "" {
		return errors.New("missing required configuration: SAMPAY_SMTP_PORT")
	}
	if cfg.SMTP.From == "" {
		return errors.New("missing required configuration: SAMPAY_SMTP_FROM")
	}
	if cfg.Temporal.Host == "" {
		return errors.New("missing required configuration: SAMPAY_TEMPORAL_HOST")
	}
	if cfg.JWT.Secret == "" {
		return errors.New("missing required configuration: Secret_KEY")
	}
	if cfg.JWT.AccessExpiry == "" {
		return errors.New("missing required configuration: JWT_ACCESS_EXPIRY")
	}
	if cfg.JWT.RefreshExpiry == "" {
		return errors.New("missing required configuration: JWT_REFRESH_EXPIRY")
	}
	return nil
}
