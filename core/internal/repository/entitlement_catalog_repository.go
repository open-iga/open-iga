package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/repository/db"
)

type EntitlementCatalogRepository struct {
	queries *db.Queries
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func NewEntitlementCatalogRepository(queries *db.Queries) *EntitlementCatalogRepository {
	return &EntitlementCatalogRepository{queries: queries}
}

func (e *EntitlementCatalogRepository) GetEntitlementCatalog(ctx context.Context, entitlementId uuid.UUID, managedSystemId string, resource string) (*domain.EntitlementCatalog, error) {
	systemId, err := parseManagedSystemId(managedSystemId)
	if err != nil {
		return nil, err
	}

	row, err := e.queries.GetEntitlementCatalog(ctx, db.GetEntitlementCatalogParams{ID: entitlementId, ManagedSystemID: systemId, Resource: resource})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: %s of resource %q in system %s", domain.ErrEntitlementNotFound, entitlementId, resource, managedSystemId)
		}

		return nil, fmt.Errorf("failed to get entitlement: %w", err)
	}

	return toEntitlementCatalog(row)
}

func (e *EntitlementCatalogRepository) UpsertEntitlementCatalog(ctx context.Context, managedSystemId string, tenantId uuid.UUID, resource string, discovered domain.DiscoveredEntitlement) (*domain.EntitlementCatalog, error) {
	systemId, err := parseManagedSystemId(managedSystemId)
	if err != nil {
		return nil, err
	}

	metadata, err := json.Marshal(discovered.Metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal entitlement metadata: %w", err)
	}

	if discovered.Metadata == nil {
		metadata = []byte("{}")
	}

	var description string
	if discovered.Description != nil {
		description = *discovered.Description
	}

	row, err := e.queries.UpsertEntitlementCatalog(ctx, db.UpsertEntitlementCatalogParams{
		ManagedSystemID: systemId,
		TenantID:        toNullableTenantId(tenantId),
		Resource:        resource,
		ExternalID:      discovered.EntitlementID,
		DisplayName:     discovered.Name,
		Description:     pgtype.Text{String: description, Valid: discovered.Description != nil},
		Metadata:        metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save entitlement: %w", err)
	}

	return toEntitlementCatalog(row)
}

// SearchEntitlementCatalog matches the query as text, not as a pattern: % and _ in it stand for themselves.
// Matches that start with the query come first, then the closest by similarity.
func (e *EntitlementCatalogRepository) SearchEntitlementCatalog(ctx context.Context, search domain.EntitlementSearch) ([]*domain.EntitlementCatalog, int64, error) {
	rows, err := e.queries.SearchEntitlementCatalog(ctx, db.SearchEntitlementCatalogParams{
		ManagedSystemID: toNullableUuid(search.ManagedSystemId),
		TenantID:        toNullableUuid(search.TenantId),
		Resource:        toNullableText(search.Resource),
		Query:           toNullableText(search.Query),
		QueryPattern:    toNullableText(likeEscaper.Replace(search.Query)),
		PageOffset:      search.Offset,
		PageSize:        search.Limit,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search entitlements: %w", err)
	}

	var total int64
	entitlements := make([]*domain.EntitlementCatalog, 0, len(rows))
	for _, row := range rows {
		entitlement, err := toEntitlementCatalog(row.EntitlementCatalog)
		if err != nil {
			return nil, 0, err
		}

		total = row.Total
		entitlements = append(entitlements, entitlement)
	}

	return entitlements, total, nil
}

// toNullableUuid stores uuid.Nil, which here means "any", as NULL.
func toNullableUuid(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: id != uuid.Nil}
}

func toNullableText(text string) pgtype.Text {
	return pgtype.Text{String: text, Valid: text != ""}
}

func toEntitlementCatalog(row db.EntitlementCatalog) (*domain.EntitlementCatalog, error) {
	var metadata domain.Metadata
	if err := json.Unmarshal(row.Metadata, &metadata); err != nil {
		return nil, fmt.Errorf("unmarshal entitlement metadata: %w", err)
	}

	return &domain.EntitlementCatalog{
		Id:              row.ID.String(),
		ManagedSystemId: row.ManagedSystemID.String(),
		TenantId:        db.FormatNullableUUID(row.TenantID),
		Resource:        row.Resource,
		ExternalId:      row.ExternalID,
		DisplayName:     row.DisplayName,
		Description:     row.Description.String,
		Metadata:        metadata,
	}, nil
}
