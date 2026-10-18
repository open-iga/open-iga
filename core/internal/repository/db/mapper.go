package db

import (
	"encoding/json"
	"log/slog"

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

func (m *ManagedSystem) ToDomain(logger *slog.Logger) *domain.ManagedSystem {
	// connector_spec is our own marshaled output, so this normally round-trips; log if a stored value ever fails to decode
	var spec *domain.ConnectorSpec
	if len(m.ConnectorSpec) > 0 {
		if err := json.Unmarshal(m.ConnectorSpec, &spec); err != nil {
			logger.Error("failed to unmarshal connector spec", "managedSystemId", m.ID.String(), "error", err)
		}
	}
	return &domain.ManagedSystem{
		Id:            m.ID.String(),
		Name:          m.Name,
		ConnectorUrl:  m.ConnectorUrl,
		ConnectorHash: m.ConnectorHash,
		ConnectorSpec: spec,
		CreatedAt:     m.CreatedAt.Time,
	}
}

func (t *ManagedSystemTenant) ToDomain() *domain.ManagedSystemTenant {
	return &domain.ManagedSystemTenant{
		Id:              t.ID.String(),
		ManagedSystemId: t.ManagedSystemID.String(),
		ExternalId:      t.ExternalID,
		Name:            t.Name,
		CreatedAt:       t.CreatedAt.Time,
	}
}
