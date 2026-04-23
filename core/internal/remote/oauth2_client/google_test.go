package oauth2_client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

type mockOauth2Config struct {
	token       *oauth2.Token
	exchangeErr error
	httpClient  *http.Client
}

func (m *mockOauth2Config) AuthCodeURL(state string, _ ...oauth2.AuthCodeOption) string {
	return "https://accounts.google.com/auth?state=" + state
}

func (m *mockOauth2Config) Exchange(_ context.Context, _ string, _ ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return m.token, m.exchangeErr
}

func (m *mockOauth2Config) Client(_ context.Context, _ *oauth2.Token) *http.Client {
	return m.httpClient
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestGoogleOauth2Client_FetchOauthUser(t *testing.T) {
	t.Run("returns error when exchange fails", func(t *testing.T) {
		client := &GoogleOauth2Client{
			config: &mockOauth2Config{exchangeErr: errors.New("exchange failed")},
		}

		user, err := client.FetchOauthUser(context.Background(), "auth-code")

		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("returns error when fetching user info fails", func(t *testing.T) {
		client := &GoogleOauth2Client{
			config: &mockOauth2Config{
				token: &oauth2.Token{AccessToken: "fake-token"},
				httpClient: &http.Client{
					Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
						return nil, errors.New("network error")
					}),
				},
			},
		}

		user, err := client.FetchOauthUser(context.Background(), "auth-code")

		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("returns error when google responds with non-2xx for user info response", func(t *testing.T) {
		client := &GoogleOauth2Client{
			config: &mockOauth2Config{
				token: &oauth2.Token{AccessToken: "fake-token"},
				httpClient: &http.Client{
					Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 422,
						}, nil
					}),
				},
			},
		}

		user, err := client.FetchOauthUser(context.Background(), "auth-code")

		assert.EqualError(t, err, "unexpected status code from google for user info: 422")
		assert.Nil(t, user)
	})

	t.Run("returns user when exchange and user info succeed", func(t *testing.T) {
		userinfo, _ := json.Marshal(googleOauth2UserinfoDto{
			Email:      "user@example.com",
			GivenName:  "John",
			FamilyName: "Doe",
		})

		client := &GoogleOauth2Client{
			config: &mockOauth2Config{
				token: &oauth2.Token{AccessToken: "fake-token"},
				httpClient: &http.Client{
					Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       io.NopCloser(strings.NewReader(string(userinfo))),
						}, nil
					}),
				},
			},
		}

		user, err := client.FetchOauthUser(context.Background(), "auth-code")

		assert.NoError(t, err)
		assert.Equal(t, "user@example.com", user.Email)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
	})
}
