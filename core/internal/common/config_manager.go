package common

import (
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type OauthConfig struct {
	ClientId     string
	ClientSecret string
}

type EnvValue interface {
	~int | ~string | ~bool
}

type AppConfig struct {
	Port          string
	IsDevelopment bool
	HostUrl       string
	Oauth         struct {
		Google OauthConfig
	}
}

func mustEnv[T EnvValue](key string, converter func(val string) T) T {
	val, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Errorf("missing environment variable %s", key))
	}

	return converter(val)
}

func getEnvWithDefault[T EnvValue](key string, defaultValue T, converter ...func(val string) T) T {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	if len(converter) > 0 && converter[0] != nil {
		return converter[0](val)
	}

	return defaultValue
}

func NewAppConfig() *AppConfig {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	return &AppConfig{
		Port: getEnvWithDefault("PORT", ":8080", func(val string) string { return val }),
		HostUrl: mustEnv("HOST_URL", func(val string) string {
			_, err := url.ParseRequestURI(val)
			if err != nil {
				panic(fmt.Errorf("invalid HOST_URL: %w", err))
			}

			return val
		}),
		IsDevelopment: getEnvWithDefault("ENVIRONMENT", false, func(val string) bool { return val == "development" }),
		Oauth: struct{ Google OauthConfig }{Google: OauthConfig{
			ClientId:     mustEnv("GOOGLE_OAUTH_CLIENT_ID", func(val string) string { return val }),
			ClientSecret: mustEnv("GOOGLE_OAUTH_CLIENT_SECRET", func(val string) string { return val }),
		}},
	}
}
