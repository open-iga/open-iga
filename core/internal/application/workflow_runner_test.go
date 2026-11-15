package application

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestWorkflowService_ExecuteConnectorOperation(t *testing.T) {
	ctx := context.Background()
	identityId, tenantId, entitlementId := uuid.New(), uuid.New(), uuid.New()

	t.Run("create runs the operation on the only tenant and saves the account", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(buildAwsSystem(), nil)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-1").Return([]*domain.ManagedSystemTenant{buildTenant(tenantId, "main")}, nil)
		mocks.connectorService.EXPECT().
			Execute(ctx, "aws", "iam-user", domain.AccountActionKind, "create", domain.OperationRef{TenantId: tenantId, IdentityId: identityId, EntitlementId: uuid.Nil}).
			Return(domain.AccountCreateOutput{AccountID: "arn-1", Created: true, Metadata: nil}, nil)
		mocks.account.EXPECT().UpsertAccount(ctx, "system-1", tenantId, identityId, "arn-1").Return(&domain.Account{}, nil)

		require.NoError(t, svc.ExecuteConnectorOperation(ctx, buildCreateStep("create-aws"), identityId))
	})

	t.Run("a single-tenant system runs without a tenant and saves the account on the system", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(buildAwsSystem(), nil)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-1").Return(nil, nil)
		mocks.connectorService.EXPECT().
			Execute(ctx, "aws", "iam-user", domain.AccountActionKind, "create", domain.OperationRef{TenantId: uuid.Nil, IdentityId: identityId, EntitlementId: uuid.Nil}).
			Return(domain.AccountCreateOutput{AccountID: "arn-1", Created: true, Metadata: nil}, nil)
		mocks.account.EXPECT().UpsertAccount(ctx, "system-1", uuid.Nil, identityId, "arn-1").Return(&domain.Account{}, nil)

		require.NoError(t, svc.ExecuteConnectorOperation(ctx, buildCreateStep("create-aws"), identityId))
	})

	t.Run("a failed connector saves no account", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(buildAwsSystem(), nil)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-1").Return(nil, nil)
		mocks.connectorService.EXPECT().Execute(ctx, "aws", "iam-user", domain.AccountActionKind, "create", gomock.Any()).Return(nil, errors.New("boom"))

		require.Error(t, svc.ExecuteConnectorOperation(ctx, buildCreateStep("create-aws"), identityId))
	})

	t.Run("grant runs the operation with the entitlement and its tenant, saving nothing", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		otherTenantId := uuid.New()
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(buildAwsSystem(), nil)
		mocks.entitlement.EXPECT().GetEntitlementCatalog(ctx, entitlementId, "system-1", "iam-user").Return(&domain.EntitlementCatalog{Id: entitlementId.String(), TenantId: tenantId.String(), Resource: "iam-user"}, nil)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-1").
			Return([]*domain.ManagedSystemTenant{buildTenant(otherTenantId, "other"), buildTenant(tenantId, "main")}, nil)
		mocks.connectorService.EXPECT().
			Execute(ctx, "aws", "iam-user", domain.EntitlementKind, "grant", domain.OperationRef{TenantId: tenantId, IdentityId: identityId, EntitlementId: entitlementId}).
			Return(domain.EntitlementGrantOutput{}, nil)

		require.NoError(t, svc.ExecuteConnectorOperation(ctx, buildGrantStep("grant-admin", entitlementId), identityId))
	})

	t.Run("a named tenant picks one of several", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		otherTenantId := uuid.New()
		step := buildCreateStep("create-aws")
		step.Operation.Tenant = "main"
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(buildAwsSystem(), nil)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-1").
			Return([]*domain.ManagedSystemTenant{buildTenant(otherTenantId, "other"), buildTenant(tenantId, "main")}, nil)
		mocks.connectorService.EXPECT().
			Execute(ctx, "aws", "iam-user", domain.AccountActionKind, "create", domain.OperationRef{TenantId: tenantId, IdentityId: identityId, EntitlementId: uuid.Nil}).
			Return(domain.AccountCreateOutput{AccountID: "arn-1", Created: true, Metadata: nil}, nil)
		mocks.account.EXPECT().UpsertAccount(ctx, "system-1", tenantId, identityId, "arn-1").Return(&domain.Account{}, nil)

		require.NoError(t, svc.ExecuteConnectorOperation(ctx, step, identityId))
	})
}
