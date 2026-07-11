package common

import (
	"errors"
	"fmt"
	"os"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/joho/godotenv"
)

type (
	OauthConfig struct {
		ClientId     string
		ClientSecret string
	}

	Oauth struct {
		Google OauthConfig
	}

	Database struct {
		URL string
	}

	Redirect struct {
		Home               string
		SignIn             string
		GoogleAuthCallback string
	}

	AdminUser struct {
		Email string
	}

	AppConfig struct {
		AdminUser AdminUser
		Port      string
		HostUrl   string
		Oauth     Oauth
		Database  Database
		Redirect  Redirect
	}
)

type (
	EnvField[T any] struct {
		envKey       string
		defaultValue T
		value        T
		// This description will be used to inform the users on why this is required and other details
		description string
		rules       []validation.Rule
	}

	ConfigFromEnv struct {
		AdminUserEmail, Port, HostUrl, GoogleOauthClientId, GoogleOauthClientSecret, DatabaseUrl *EnvField[any]
	}
)

func (e *EnvField[T]) IsRequired() *EnvField[T] {
	e.rules = append(e.rules, validation.Required)
	return e
}

func (e *EnvField[T]) WithDefault(val T) *EnvField[T] {
	e.defaultValue = val
	return e
}

func (e *EnvField[T]) WithDescription(desc string) *EnvField[T] {
	e.description = desc
	return e
}

func (e *EnvField[T]) WithValidationRules(rules ...validation.Rule) *EnvField[T] {
	e.rules = append(e.rules, rules...)
	return e
}

func (e *EnvField[T]) Validate() error {
	var valueToParse any

	raw, ok := os.LookupEnv(e.envKey)
	if !ok {
		valueToParse = e.defaultValue
	} else {
		valueToParse = raw
	}

	err := validation.Validate(valueToParse, e.rules...)
	if err != nil {
		return fmt.Errorf("%s: %v. Details: %s", e.envKey, err, e.description)
	}

	e.value = valueToParse.(T)
	return nil
}

func Var[T any](key string) *EnvField[T] {
	return &EnvField[T]{envKey: key}
}

var (
	AdminUserEmail          = Var[any]("ADMIN_USER_EMAIL").WithDefault("").WithDescription("Admin user email. This is required during the first run").WithValidationRules(is.Email)
	Port                    = Var[any]("PORT").WithDefault("8080").WithDescription("Port to listen on").WithValidationRules(is.Port).IsRequired()
	HostUrl                 = Var[any]("HOST_URL").WithDescription("URL at which this server is running. This is used for Oauth callback").WithValidationRules(is.URL).IsRequired()
	GoogleOauthClientId     = Var[any]("GOOGLE_OAUTH_CLIENT_ID").WithDescription("Google Client ID for Oauth2 flow").IsRequired()
	GoogleOauthClientSecret = Var[any]("GOOGLE_OAUTH_CLIENT_SECRET").WithDescription("Google Client Secret for Oauth2 flow").IsRequired()
	DatabaseUrl             = Var[any]("DATABASE_URL").WithDescription("PostgreSQL connection string").IsRequired()

	configFromEnv = ConfigFromEnv{AdminUserEmail, Port, HostUrl, GoogleOauthClientId, GoogleOauthClientSecret, DatabaseUrl}
)

func validateEnv() (*ConfigFromEnv, error) {
	var validationError error

	for _, value := range StructToMap(configFromEnv) {
		envField := value.(*EnvField[any])
		if err := envField.Validate(); err != nil {
			validationError = errors.Join(validationError, err)
		}
	}

	if validationError != nil {
		return nil, validationError
	}

	return &configFromEnv, nil
}

func NewAppConfig() *AppConfig {
	_ = godotenv.Load(".env")
	validatedEnv, err := validateEnv()
	if err != nil {
		panic(err)
	}

	return &AppConfig{
		AdminUser: AdminUser{
			Email: validatedEnv.AdminUserEmail.value.(string),
		},
		Port:    ":" + validatedEnv.Port.value.(string),
		HostUrl: validatedEnv.HostUrl.value.(string),
		Oauth: Oauth{
			Google: OauthConfig{
				ClientId:     validatedEnv.GoogleOauthClientId.value.(string),
				ClientSecret: validatedEnv.GoogleOauthClientSecret.value.(string),
			},
		},
		Database: Database{
			URL: validatedEnv.DatabaseUrl.value.(string),
		},
		Redirect: Redirect{
			Home:               "/",
			SignIn:             "/auth/sign-in",
			GoogleAuthCallback: "/auth/google/callback",
		},
	}
}
