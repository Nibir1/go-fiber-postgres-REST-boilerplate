package util

import (
	"time"

	// Used for configuration loading from various sources (YAML, environment variables, etc.)
	"github.com/spf13/viper"
)

// Config represents the application configuration loaded from a file or environment variables
type Config struct {
	DBDriver            string        `mapstructure:"DB_DRIVER"`             // Database driver to use (e.g., "mysql", "postgres")
	DBSource            string        `mapstructure:"DB_SOURCE"`             // Database connection source string
	ServerAddress       string        `mapstructure:"SERVER_ADDRESS"`        // Server address where the application listens (e.g., "localhost:8080")
	TokenSymmetricKey   string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`   // Symmetric key used for token signing (should be kept secret)
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"` // Duration for which access tokens are valid
}

// LoadConfig reads the application configuration from a specified file or environment variables
func LoadConfig(path string) (config Config, err error) {
	// Add the provided path as a potential location for the configuration file
	viper.AddConfigPath(path)

	// Set the configuration file name to "app"
	viper.SetConfigName("app")

	// Set the configuration file type to environment variables (".env")
	viper.SetConfigType("env")

	// Automatically map environment variables with a "APP_" prefix to configuration keys (e.g., APP_DB_DRIVER becomes DB_DRIVER)
	viper.AutomaticEnv()

	// Attempt to read the configuration file
	err = viper.ReadInConfig()
	if err != nil {
		// Return an error if the configuration file cannot be read
		return
	}

	// Unmarshal the loaded configuration data into the Config struct
	err = viper.Unmarshal(&config)
	return
}
