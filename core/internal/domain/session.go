package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNoActiveSession       = errors.New("no active session")
	ErrSessionNotFound       = errors.New("no session found")
	ErrInactiveSession       = errors.New("inactive session")
	ErrFailedToExpireSession = errors.New("failed to expire session")
	ErrExpiredSession        = errors.New("expired session")
)

type Session struct {
	Id         uuid.UUID
	SessionId  string
	IdentityId uuid.UUID
	Active     bool
	CreatedAt  time.Time
	ExpiresAt  time.Time // this is in UTC. Refer to GenerateSession on IdentitySession
}

func (s *Session) ValidityInSeconds() int {
	return int(s.ExpiresAt.Sub(time.Now().UTC()).Seconds())
}

func (s *Session) IsExpired() bool {
	return s.ExpiresAt.Sub(time.Now().UTC()) < 0
}
