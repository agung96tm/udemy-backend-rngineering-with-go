package env

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// InitEnv loads environment variables from .env (if exists)
// and enables Viper to read OS environment variables.
func InitEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using system environment variables")
	}
	viper.AutomaticEnv()
}

// GetString returns string value from env or default if missing
func GetString(key string, defaultValue string) string {
	if !viper.IsSet(key) {
		return defaultValue
	}
	return viper.GetString(key)
}

// GetInt returns int value from env or default if missing
func GetInt(key string, defaultValue int) int {
	if !viper.IsSet(key) {
		return defaultValue
	}
	return viper.GetInt(key)
}

// GetBool returns bool value from env or default if missing
func GetBool(key string, defaultValue bool) bool {
	if !viper.IsSet(key) {
		return defaultValue
	}
	return viper.GetBool(key)
}
