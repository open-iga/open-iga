package application

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type discoveryMocks struct {
	managedSystem *testutil.MockManagedSystemRepository
}

func setupDiscoveryService(t *testing.T) (*DiscoveryService, discoveryMocks) {
	ctrl := gomock.NewController(t)
	mocks := discoveryMocks{
		managedSystem: testutil.NewMockManagedSystemRepository(ctrl),
	}

	return NewDiscoveryService(mocks.managedSystem), mocks
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
			{SystemId: "system-1", SystemName: "keycloak", TenantId: firstTenantId, TenantName: "realm-a", Resource: "group"},
			{SystemId: "system-1", SystemName: "keycloak", TenantId: secondTenantId, TenantName: "realm-b", Resource: "group"},
			{SystemId: "system-1", SystemName: "keycloak", TenantId: firstTenantId, TenantName: "realm-a", Resource: "role"},
			{SystemId: "system-1", SystemName: "keycloak", TenantId: secondTenantId, TenantName: "realm-b", Resource: "role"},
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
