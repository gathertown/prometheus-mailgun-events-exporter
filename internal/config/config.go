package config

import "os"

const (
	defaultLogLevel = "debug"
	defaultDomain   = "mail.gather.town"
	defaultApiKey   = "_this_is_a_secret_"
)

// Config contains all the available configuration options
type Config struct {
	Domain   string
	ApiKey   string
	LogLevel string
}

// FrontEnv populates the application configuration from environmental variables
func FromEnv() *Config {
	var (
		logLevel = getenv("LOGLEVEL", defaultLogLevel)
		domain   = getenv("DOMAIN", defaultDomain)
		apiKey   = getenv("MG_API_KEY", defaultApiKey)
	)

	c := &Config{
		LogLevel: logLevel,
		Domain:   domain,
		ApiKey:   apiKey,
	}
	return c
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
