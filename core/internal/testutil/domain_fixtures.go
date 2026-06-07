package testutil

import (
	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
)

func WithOauthUser(oauthUser domain.OauthUser) func(identity *domain.Identity) {
	return func(identity *domain.Identity) {
		identity.Email = oauthUser.Email
		identity.FirstName = oauthUser.FirstName
		identity.LastName = oauthUser.LastName
	}
}

func NewIdentity(overwrite ...func(identity *domain.Identity)) domain.Identity {
	identity := domain.Identity{
		Id:        uuid.New(),
		FirstName: "firstname",
		LastName:  "lastname",
		Email:     "test@gmail.com",
	}

	for _, fn := range overwrite {
		fn(&identity)
	}

	return identity
}

func NewOauthUser() domain.OauthUser {
	return domain.OauthUser{
		FirstName: "firstname",
		LastName:  "lastname",
		Email:     "test@gmail.com",
	}
}
