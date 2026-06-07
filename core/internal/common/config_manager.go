package common

import (
	"fmt"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type EnvValue interface {
	~int | ~string | ~bool
}

type OauthConfig struct {
	ClientId     string
	ClientSecret string
}

type Oauth struct {
	Google OauthConfig
}

type Database struct {
	URL string
}
type Redirect struct {
	Home               string
	SignIn             string
	GoogleAuthCallback string
}

type AppConfig struct {
	Port     string
	HostUrl  string
	Oauth    Oauth
	Database Database
	Redirect Redirect
}

func mustEnv[T EnvValue](key string, converter func(val string) T) T {
	val, ok := os.LookupEnv(key)
	if !ok {
		panic(fmt.Errorf("missing environment variable %s", key))
	}

	return converter(val)
}

func envWithDefault[T EnvValue](key string, defaultValue T, converter ...func(val string) T) T {
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
	_ = godotenv.Load(".env")

	return &AppConfig{
		Port: envWithDefault("PORT", ":8080", func(val string) string { return val }),
		HostUrl: mustEnv("HOST_URL", func(val string) string {
			_, err := url.ParseRequestURI(val)
			if err != nil {
				panic(fmt.Errorf("invalid HOST_URL: %w", err))
			}

			return val
		}),
		Oauth: Oauth{
			Google: OauthConfig{
				ClientId:     mustEnv("GOOGLE_OAUTH_CLIENT_ID", func(val string) string { return val }),
				ClientSecret: mustEnv("GOOGLE_OAUTH_CLIENT_SECRET", func(val string) string { return val }),
			},
		},
		Database: Database{
			URL: mustEnv("DATABASE_URL", func(val string) string { return val }),
		},
		Redirect: Redirect{
			Home:               "/",
			SignIn:             "/auth/sign-in",
			GoogleAuthCallback: "/auth/google/callback",
		},
	}
}
