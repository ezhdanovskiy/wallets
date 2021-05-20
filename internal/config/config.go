// Package config contains Config struct that is used for configuring application.
package config

import (
	"github.com/spf13/viper"
)

// Config contains all parameter for configuring application.
type Config struct {
	LogLevel    string `mapstructure:"log_level"`
	LogEncoding string `mapstructure:"log_encoding"` // json/console
	HttpPort    int    `mapstructure:"http_port"`
	DB          DB
}

// DB contains parameter for configuring repository.
type DB struct {
	Host           string `mapstructure:"db_host"`
	Port           int    `mapstructure:"db_port"`
	User           string `mapstructure:"db_user"`
	Password       string `mapstructure:"db_password"`
	DBName         string `mapstructure:"db_name"`
	MigrationsPath string `mapstructure:"migrations_path"`
}

// NewConfig creates a new Config instance with parameters parsed by viber.
func NewConfig() (*Config, error) {
	config := &Config{}
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_encoding", "json")
	viper.SetDefault("http_port", 8080)

	viper.SetDefault("db_host", "localhost")
	viper.SetDefault("db_port", 5432)
	viper.SetDefault("db_user", "postgres")
	viper.SetDefault("db_password", "postgres")
	viper.SetDefault("db_name", "postgres")
	viper.SetDefault("migrations_path", "migrations")

	_ = viper.ReadInConfig()

	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&config.DB); err != nil {
		return nil, err
	}

	return config, nil
}
