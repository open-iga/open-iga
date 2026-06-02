package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertPanicWithError(t *testing.T, errorMessage string) {
	r := recover()
	require.NotNil(t, r)

	err, ok := r.(error)
	require.Truef(t, ok, "expected error panic, got %T", r)
	assert.EqualError(t, err, errorMessage)
}

func TestConfigManager(t *testing.T) {
	t.Run("panics when HOST_URL is missing in env ", func(t *testing.T) {
		defer assertPanicWithError(t, "missing environment variable HOST_URL")

		NewAppConfig()
	})

	t.Run("panics when GOOGLE_OAUTH_CLIENT_ID is missing in env", func(t *testing.T) {
		t.Setenv("HOST_URL", "http://localhost:8080")
		defer assertPanicWithError(t, "missing environment variable GOOGLE_OAUTH_CLIENT_ID")

		NewAppConfig()
	})

	t.Run("panics when GOOGLE_OAUTH_CLIENT_SECRET is missing in env", func(t *testing.T) {
		t.Setenv("HOST_URL", "http://localhost:8080")
		t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "dummy-client-id")
		defer assertPanicWithError(t, "missing environment variable GOOGLE_OAUTH_CLIENT_SECRET")

		NewAppConfig()
	})

	t.Run("panics when DB url is missing in env", func(t *testing.T) {
		t.Setenv("HOST_URL", "http://localhost:8080")
		t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "dummy-client-id")
		t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "dummy-client-secret")
		defer assertPanicWithError(t, "missing environment variable DATABASE_URL")

		NewAppConfig()
	})

	t.Run("returns AppConfig with correct values when all env variables are set", func(t *testing.T) {
		t.Setenv("HOST_URL", "http://localhost:8080")
		t.Setenv("GOOGLE_OAUTH_CLIENT_ID", "dummy-client-id")
		t.Setenv("GOOGLE_OAUTH_CLIENT_SECRET", "dummy-client-secret")
		t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/open_iga")

		config := NewAppConfig()

		assert.Equal(t, &AppConfig{
			Port:    ":8080",
			HostUrl: "http://localhost:8080",
			Oauth: struct{ Google OauthConfig }{
				Google: OauthConfig{
					ClientId:     "dummy-client-id",
					ClientSecret: "dummy-client-secret",
				},
			},
			Database: Database{
				URL: "postgres://test:test@localhost:5432/open_iga",
			},
			Redirect: Redirect{
				Home:               "/",
				SignIn:             "/auth/sign-in",
				GoogleAuthCallback: "/auth/google/callback",
			},
		}, config)
	})
}
