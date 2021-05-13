// Package config contains the Config struct that is used for configuring the Application.
package config

import (
	"github.com/spf13/viper"
)

// Config contains all parameter for configuring the Application.
type Config struct {
	LogLevel    string `mapstructure:"log_level"`
	LogEncoding string `mapstructure:"log_encoding"`
	DBHost      string `mapstructure:"db_host"`
	DBPort      int    `mapstructure:"db_port"`
	DBUser      string `mapstructure:"db_user"`
	DBPassword  string `mapstructure:"db_password"`
	DBName      string `mapstructure:"db_name"`
	HttpPort    int    `mapstructure:"http_port"`
}

// NewConfig creates a new Config instance with parameters parsed by viber.
func NewConfig() (*Config, error) {
	config := &Config{}
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	viper.SetDefault("log_level", "info")
	viper.SetDefault("log_encoding", "json")
	viper.SetDefault("db_host", "localhost")
	viper.SetDefault("db_port", 5432)
	viper.SetDefault("db_user", "postgres")
	viper.SetDefault("db_password", "postgres")
	viper.SetDefault("db_name", "postgres")
	viper.SetDefault("http_port", 8080)

	_ = viper.ReadInConfig()

	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
