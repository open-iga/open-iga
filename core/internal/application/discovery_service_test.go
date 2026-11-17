package application

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type discoveryMocks struct {
	connectorService *testutil.MockConnectorService
	managedSystem    *testutil.MockManagedSystemRepository
	entitlement      *testutil.MockEntitlementCatalogRepository
}

func setupDiscoveryService(t *testing.T) (*DiscoveryService, discoveryMocks) {
	ctrl := gomock.NewController(t)
	mocks := discoveryMocks{
		connectorService: testutil.NewMockConnectorService(ctrl),
		managedSystem:    testutil.NewMockManagedSystemRepository(ctrl),
		entitlement:      testutil.NewMockEntitlementCatalogRepository(ctrl),
	}

	return NewDiscoveryService(mocks.connectorService, mocks.managedSystem, mocks.entitlement), mocks
}

func TestDiscoveryService_ListDiscoveryTargets(t *testing.T) {
	ctx := context.Background()
	firstTenantId, secondTenantId := uuid.New(), uuid.New()

	t.Run("a multi-tenant system gives one target per tenant and resource, a single-tenant system one per resource", func(t *testing.T) {
		svc, mocks := setupDiscoveryService(t)
		multiTenant := &domain.ManagedSystem{Id: "system-1", Name: "keycloak", ConnectorSpec: &domain.ConnectorSpec{
			Entitlements: map[string]domain.UserEntitlements{"role": {}, "group": {}},
		}}
		singleTenant := &domain.ManagedSystem{Id: "system-2", Name: "aws", ConnectorSpec: &domain.ConnectorSpec{
			Entitlements: map[string]domain.UserEntitlements{"iam-user": {}},
		}}
		mocks.managedSystem.EXPECT().ListManagedSystems(ctx).Return([]*domain.ManagedSystem{multiTenant, singleTenant}, nil)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-1").Return([]*domain.ManagedSystemTenant{buildTenant(firstTenantId, "realm-a"), buildTenant(secondTenantId, "realm-b")}, nil)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-2").Return(nil, nil)

		targets, err := svc.ListDiscoveryTargets(ctx)

		require.NoError(t, err)
		assert.Equal(t, []domain.DiscoveryTarget{
			{SystemId: "system-1", SystemName: "keycloak", TenantId: firstTenantId, Resource: "group"},
			{SystemId: "system-1", SystemName: "keycloak", TenantId: secondTenantId, Resource: "group"},
			{SystemId: "system-1", SystemName: "keycloak", TenantId: firstTenantId, Resource: "role"},
			{SystemId: "system-1", SystemName: "keycloak", TenantId: secondTenantId, Resource: "role"},
			{SystemId: "system-2", SystemName: "aws", TenantId: uuid.Nil, Resource: "iam-user"},
		}, targets)
	})

	t.Run("a system without a connector spec is skipped", func(t *testing.T) {
		svc, mocks := setupDiscoveryService(t)
		mocks.managedSystem.EXPECT().ListManagedSystems(ctx).Return([]*domain.ManagedSystem{{Id: "system-1", Name: "broken"}}, nil)

		targets, err := svc.ListDiscoveryTargets(ctx)

		require.NoError(t, err)
		assert.Empty(t, targets)
	})
}

func TestDiscoveryService_DiscoverPage(t *testing.T) {
	ctx := context.Background()
	tenantId := uuid.New()
	target := domain.DiscoveryTarget{SystemId: "system-1", SystemName: "aws", TenantId: tenantId, Resource: "iam-user"}
	admin := domain.DiscoveredEntitlement{EntitlementID: "role-admin", Name: "Admin"}
	reader := domain.DiscoveredEntitlement{EntitlementID: "role-reader", Name: "Reader"}

	t.Run("saves every entitlement of the page and returns the next cursor", func(t *testing.T) {
		svc, mocks := setupDiscoveryService(t)
		mocks.connectorService.EXPECT().
			Execute(ctx, "aws", "iam-user", domain.EntitlementKind, "discover", domain.OperationRef{TenantId: tenantId, Cursor: "p1"}).
			Return(domain.EntitlementDiscoverOutput{Entitlements: []domain.DiscoveredEntitlement{admin, reader}, NextCursor: "p2"}, nil)
		mocks.entitlement.EXPECT().UpsertEntitlementCatalog(ctx, "system-1", tenantId, "iam-user", admin).Return(&domain.EntitlementCatalog{}, nil)
		mocks.entitlement.EXPECT().UpsertEntitlementCatalog(ctx, "system-1", tenantId, "iam-user", reader).Return(&domain.EntitlementCatalog{}, nil)

		nextCursor, err := svc.DiscoverPage(ctx, target, "p1")

		require.NoError(t, err)
		assert.Equal(t, "p2", nextCursor)
	})

	t.Run("the last page returns an empty cursor", func(t *testing.T) {
		svc, mocks := setupDiscoveryService(t)
		mocks.connectorService.EXPECT().Execute(ctx, "aws", "iam-user", domain.EntitlementKind, "discover", gomock.Any()).
			Return(domain.EntitlementDiscoverOutput{Entitlements: []domain.DiscoveredEntitlement{admin}}, nil)
		mocks.entitlement.EXPECT().UpsertEntitlementCatalog(ctx, "system-1", tenantId, "iam-user", admin).Return(&domain.EntitlementCatalog{}, nil)

		nextCursor, err := svc.DiscoverPage(ctx, target, "")

		require.NoError(t, err)
		assert.Empty(t, nextCursor)
	})

	t.Run("a failed connector saves nothing", func(t *testing.T) {
		svc, mocks := setupDiscoveryService(t)
		mocks.connectorService.EXPECT().Execute(ctx, "aws", "iam-user", domain.EntitlementKind, "discover", gomock.Any()).Return(nil, errors.New("boom"))

		_, err := svc.DiscoverPage(ctx, target, "")

		require.Error(t, err)
	})
}
