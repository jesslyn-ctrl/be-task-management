package config

import (
	_logger "bitbucket.org/edts/go-task-management/pkg/logger"
	"log"
	"time"

	"github.com/spf13/viper"
)

var logs = _logger.GetContextLoggerf(nil)

// Config structure for the application
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWT            `mapstructure:"jwt"`
}

// AppConfig holds application-related settings
type AppConfig struct {
	Name string `mapstructure:"name"`
	Port string `mapstructure:"port"`
}

// DatabaseConfig holds db related settings
type DatabaseConfig struct {
	URL               string        `mapstructure:"url"`
	MaxConnections    int32         `mapstructure:"max_connections"`
	MinConnections    int32         `mapstructure:"min_connections"`
	MaxIdleTime       time.Duration `mapstructure:"max_idle_time"`
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`
}

// JWT holds jwt related settings
type JWT struct {
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
}

// Global variable to store the loaded config
var AppConfigInstance Config

// LoadConfig reads from config.yaml and sets values in AppConfigInstance
func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	// Search for config.yaml in the config root directory
	viper.AddConfigPath(".")

	// Override settings with env var if available
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	if err := viper.Unmarshal(&AppConfigInstance); err != nil {
		log.Fatalf("Error parsing config file: %v", err)
	}

	logs.Info("✅ Config loaded successfully!")
}
