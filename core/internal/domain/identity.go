package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/common"
)

type (
	// IdentityRole list of roles assigned to an identity
	// IdentityRole is not mixed with Identity, as Session(authn) and IdentityRole(authz) are separate concerns
	IdentityRole struct {
		IdentityId uuid.UUID
		Roles      []string
	}

	// Identity core model of domain
	Identity struct {
		Id        uuid.UUID
		FirstName string
		LastName  string
		Email     string
	}

	// IdentitySession To create session for an identity
	IdentitySession struct {
		SessionId string
		ExpiresAt time.Time
	}
)

var ErrNoIdentityFound = errors.New("no identity found")

const (
	sessionValidity     = 10 * time.Hour
	DefaultIdentityRole = "member"
	AdminRole           = "admin"
)

func (i *Identity) GenerateSession() (*IdentitySession, error) {
	sessionId, err := common.GenerateHighEntropyID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID %w", err)
	}

	expiresAt := time.Now().UTC().Add(sessionValidity)

	return &IdentitySession{SessionId: sessionId, ExpiresAt: expiresAt}, nil
}
