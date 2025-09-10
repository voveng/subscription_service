package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode)
}

func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	if err := viper.BindEnv("server.port", "PORT"); err != nil {
		return nil, fmt.Errorf("failed to bind server port: %w", err)
	}
	if err := viper.BindEnv("database.host", "DB_HOST"); err != nil {
		return nil, fmt.Errorf("failed to bind database host: %w", err)
	}
	if err := viper.BindEnv("database.port", "DB_PORT"); err != nil {
		return nil, fmt.Errorf("failed to bind database port: %w", err)
	}
	if err := viper.BindEnv("database.user", "DB_USER"); err != nil {
		return nil, fmt.Errorf("failed to bind database user: %w", err)
	}
	if err := viper.BindEnv("database.password", "DB_PASSWORD"); err != nil {
		return nil, fmt.Errorf("failed to bind database password: %w", err)
	}
	if err := viper.BindEnv("database.dbname", "DB_NAME"); err != nil {
		return nil, fmt.Errorf("failed to bind database name: %w", err)
	}
	if err := viper.BindEnv("database.sslmode", "DB_SSLMODE"); err != nil {
		return nil, fmt.Errorf("failed to bind database sslmode: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
