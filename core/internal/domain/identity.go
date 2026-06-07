package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/common"
)

type Identity struct {
	Id        uuid.UUID
	FirstName string
	LastName  string
	Email     string
}

const sessionValidity = 10 * time.Minute

type IdentitySession struct {
	SessionId string
	ExpiresAt time.Time
}

func (i *Identity) GenerateSession() (*IdentitySession, error) {
	sessionId, err := common.GenerateHighEntropyID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID %w", err)
	}

	expiresAt := time.Now().UTC().Add(sessionValidity)

	return &IdentitySession{SessionId: sessionId, ExpiresAt: expiresAt}, nil
}
