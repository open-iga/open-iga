package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	NoActiveSession       = errors.New("no active session")
	SessionNotFound       = errors.New("no session found")
	InactiveSession       = errors.New("inactive session")
	FailedToExpireSession = errors.New("failed to expire session")
	ExpiredSession        = errors.New("expired session")
)

type Session struct {
	Id         uuid.UUID
	SessionId  string
	IdentityId uuid.UUID
	Active     bool
	CreatedAt  time.Time
	ExpiresAt  time.Time // this is in UTC. Refer to IdentitySession
}

func (s *Session) ValidityInSeconds() int {
	return int(s.ExpiresAt.Sub(time.Now().UTC()).Seconds())
}

func (s *Session) IsExpired() bool {
	return s.ExpiresAt.Sub(time.Now().UTC()) < 0
}
