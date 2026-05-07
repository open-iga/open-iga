package db

import (
	"github.com/open-iga/core/internal/domain"
)

func (i *Identity) ToDomain() *domain.Identity {
	return &domain.Identity{
		Id:        i.ID,
		FirstName: i.FirstName.String,
		LastName:  i.LastName.String,
		Email:     i.Email,
	}
}

func (s *Session) ToDomain() *domain.Session {
	return &domain.Session{
		Id:         s.ID,
		SessionId:  s.SessionID,
		IdentityId: s.IdentityID,
		Active:     s.Active,
		CreatedAt:  s.CreatedAt.Time,
		ExpiresAt:  s.ExpiresAt.Time,
	}
}
