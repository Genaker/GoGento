package config

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// AppConfig holds global application configuration
var AppConfig *Config
var once sync.Once

type Config struct {
	AppName  string
	Port     string
	Env      string
	Debug    bool
	MediaUrl string
	// Add more fields as needed
}

// GetBasePath returns the project's root directory
func GetBasePath() string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	return filepath.Dir(basepath) // Adjust based on your project structure
}

// LoadAppConfig initializes the global AppConfig variable
func LoadAppConfig() {
	once.Do(func() {
		AppConfig = &Config{
			AppName:  os.Getenv("APP_NAME"),
			Port:     os.Getenv("PORT"),
			Env:      os.Getenv("APP_ENV"),
			Debug:    os.Getenv("DEBUG") == "true",
			MediaUrl: GetEnv("MEDIA_URL", "http://localhost/media/"),
		}
	})
}

// GetEnvOrDefault returns the value of the environment variable key, or defaultVal if not set or empty.
func GetEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
