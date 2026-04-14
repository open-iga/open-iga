package adapter

import (
	"context"

	"github.com/open-iga/core/internal/domain"
)

type ConsentPageDetails struct {
	AuthCodeURL string
	State       string
}

type Client interface {
	// GetConsentPageDetails returns auth code URL for the consent page
	GetConsentPageDetails(ctx context.Context) *ConsentPageDetails

	// FetchOauthUser Once the user provides consent to the required data, this method is invoked with the authorization code
	// This function should perform token exchange with authorization code and provide the user information
	// State check for CSRF is not required as the state is generated and stored in the session by the API layer.
	//The API layer will only invoke this method if the state check is successful
	FetchOauthUser(ctx context.Context, code string) (*domain.OauthUser, error)
}
