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

type workflowMocks struct {
	engine           *testutil.MockWorkflowEngine
	connectorService *testutil.MockConnectorService
	workflowRepo     *testutil.MockWorkflowRepository
	managedSystem    *testutil.MockManagedSystemRepository
	entitlement      *testutil.MockEntitlementCatalogRepository
	identity         *testutil.MockIdentityRepository
	account          *testutil.MockAccountRepository
}

func setupWorkflowService(t *testing.T) (*WorkflowService, workflowMocks) {
	ctrl := gomock.NewController(t)
	mocks := workflowMocks{
		engine:           testutil.NewMockWorkflowEngine(ctrl),
		connectorService: testutil.NewMockConnectorService(ctrl),
		workflowRepo:     testutil.NewMockWorkflowRepository(ctrl),
		managedSystem:    testutil.NewMockManagedSystemRepository(ctrl),
		entitlement:      testutil.NewMockEntitlementCatalogRepository(ctrl),
		identity:         testutil.NewMockIdentityRepository(ctrl),
		account:          testutil.NewMockAccountRepository(ctrl),
	}
	svc := NewWorkflowService(testutil.NewTestLogger(), mocks.engine, mocks.connectorService, mocks.workflowRepo, mocks.managedSystem, mocks.entitlement, mocks.identity, mocks.account)
	return svc, mocks
}

// buildAwsSystem is a managed system with an "iam-user" resource for accounts and entitlements.
func buildAwsSystem() *domain.ManagedSystem {
	return &domain.ManagedSystem{
		Id: "system-1", Name: "aws",
		ConnectorSpec: &domain.ConnectorSpec{
			Actions:      map[string]domain.AccountActions{"iam-user": {Create: &domain.Operation{}, Disable: &domain.Operation{}}},
			Entitlements: map[string]domain.UserEntitlements{"iam-user": {Grant: &domain.Operation{}, Revoke: &domain.Operation{}}},
		},
	}
}

func buildTenant(id uuid.UUID, name string) *domain.ManagedSystemTenant {
	return &domain.ManagedSystemTenant{Id: id.String(), ManagedSystemId: "system-1", Name: name}
}

func buildCreateStep(name string, after ...string) domain.WorkflowStep {
	return domain.WorkflowStep{
		Name:      name,
		Operation: domain.WorkflowOperation{System: "aws", Resource: "iam-user", Kind: domain.AccountActionKind},
		After:     after,
	}
}

func buildGrantStep(name string, entitlementId uuid.UUID, after ...string) domain.WorkflowStep {
	return domain.WorkflowStep{
		Name:      name,
		Operation: domain.WorkflowOperation{System: "aws", Resource: "iam-user", Kind: domain.EntitlementKind, Entitlement: entitlementId},
		After:     after,
	}
}

func TestWorkflowService_CreateWorkflow(t *testing.T) {
	ctx := context.Background()
	tenantId, entitlementId := uuid.New(), uuid.New()
	definition := domain.WorkflowDefinition{Name: "onboarding", Steps: []domain.WorkflowStep{
		buildCreateStep("create-aws"),
		buildGrantStep("grant-admin", entitlementId, "create-aws"),
	}}

	t.Run("saves a workflow whose systems, resources and entitlements exist", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(buildAwsSystem(), nil).Times(2)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-1").DoAndReturn(func(context.Context, string) ([]*domain.ManagedSystemTenant, error) {
			return []*domain.ManagedSystemTenant{buildTenant(tenantId, "main")}, nil
		}).Times(2)
		mocks.entitlement.EXPECT().GetEntitlementCatalog(ctx, entitlementId, "system-1", "iam-user").Return(&domain.EntitlementCatalog{Id: entitlementId.String(), TenantId: tenantId.String(), Resource: "iam-user"}, nil)
		mocks.workflowRepo.EXPECT().CreateWorkflow(ctx, definition).Return(nil)

		require.NoError(t, svc.CreateWorkflow(ctx, definition))
	})

	t.Run("saves a workflow on a system without tenants and adds no tenant", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(buildAwsSystem(), nil)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-1").Return(nil, nil)
		single := domain.WorkflowDefinition{Name: "onboarding", Steps: []domain.WorkflowStep{buildCreateStep("create-aws")}}
		mocks.workflowRepo.EXPECT().CreateWorkflow(ctx, single).Return(nil)

		require.NoError(t, svc.CreateWorkflow(ctx, single))
	})

	t.Run("rejects a named tenant on a system without tenants", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(buildAwsSystem(), nil)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-1").Return(nil, nil)
		step := buildCreateStep("create-aws")
		step.Operation.Tenant = "main"

		err := svc.CreateWorkflow(ctx, domain.WorkflowDefinition{Name: "w", Steps: []domain.WorkflowStep{step}})
		assert.ErrorIs(t, err, domain.ErrTenantNotFound)
	})

	t.Run("refuses a name that is taken", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(buildAwsSystem(), nil)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-1").Return(nil, nil)
		single := domain.WorkflowDefinition{Name: "onboarding", Steps: []domain.WorkflowStep{buildCreateStep("create-aws")}}
		mocks.workflowRepo.EXPECT().CreateWorkflow(ctx, single).Return(domain.ErrWorkflowAlreadyExists)

		assert.ErrorIs(t, svc.CreateWorkflow(ctx, single), domain.ErrWorkflowAlreadyExists)
	})

	t.Run("rejects a kind a workflow cannot run", func(t *testing.T) {
		svc, _ := setupWorkflowService(t)
		step := buildCreateStep("other-aws")
		step.Operation.Kind = "other"

		err := svc.CreateWorkflow(ctx, domain.WorkflowDefinition{Name: "w", Steps: []domain.WorkflowStep{step}})
		assert.ErrorIs(t, err, domain.ErrInvalidWorkflow)
	})

	t.Run("rejects two create steps on the same system and tenant", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(buildAwsSystem(), nil).Times(2)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-1").Return(nil, nil).Times(2)

		twice := domain.WorkflowDefinition{Name: "w", Steps: []domain.WorkflowStep{buildCreateStep("create-one"), buildCreateStep("create-two")}}
		assert.ErrorIs(t, svc.CreateWorkflow(ctx, twice), domain.ErrInvalidWorkflow)
	})

	t.Run("rejects a system that does not exist", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(nil, domain.ErrManagedSystemNotFound)

		err := svc.CreateWorkflow(ctx, domain.WorkflowDefinition{Name: "w", Steps: []domain.WorkflowStep{buildCreateStep("create-aws")}})
		assert.ErrorIs(t, err, domain.ErrManagedSystemNotFound)
	})

	t.Run("rejects a resource the connector does not have", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		system := buildAwsSystem()
		system.ConnectorSpec.Actions = map[string]domain.AccountActions{"other": {}}
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(system, nil)

		err := svc.CreateWorkflow(ctx, domain.WorkflowDefinition{Name: "w", Steps: []domain.WorkflowStep{buildCreateStep("create-aws")}})
		assert.ErrorIs(t, err, domain.ErrInvalidWorkflow)
	})

	t.Run("rejects an entitlement of another tenant", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(buildAwsSystem(), nil)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-1").Return([]*domain.ManagedSystemTenant{buildTenant(tenantId, "main")}, nil)
		mocks.entitlement.EXPECT().GetEntitlementCatalog(ctx, entitlementId, "system-1", "iam-user").Return(&domain.EntitlementCatalog{Id: entitlementId.String(), TenantId: uuid.NewString(), Resource: "iam-user"}, nil)

		err := svc.CreateWorkflow(ctx, domain.WorkflowDefinition{Name: "w", Steps: []domain.WorkflowStep{buildGrantStep("grant", entitlementId)}})
		assert.ErrorIs(t, err, domain.ErrTenantNotFound)
	})

	t.Run("rejects an entitlement that names no tenant on a system that has tenants", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(buildAwsSystem(), nil)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-1").Return([]*domain.ManagedSystemTenant{buildTenant(tenantId, "main")}, nil)
		mocks.entitlement.EXPECT().GetEntitlementCatalog(ctx, entitlementId, "system-1", "iam-user").Return(&domain.EntitlementCatalog{Id: entitlementId.String(), TenantId: "", Resource: "iam-user"}, nil)

		err := svc.CreateWorkflow(ctx, domain.WorkflowDefinition{Name: "w", Steps: []domain.WorkflowStep{buildGrantStep("grant", entitlementId)}})
		assert.ErrorIs(t, err, domain.ErrEntitlementNotFound)
		assert.ErrorContains(t, err, "belongs to no tenant")
	})

	t.Run("rejects an entitlement that is not in the catalog for the step's resource", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(buildAwsSystem(), nil)
		mocks.entitlement.EXPECT().GetEntitlementCatalog(ctx, entitlementId, "system-1", "iam-user").Return(nil, domain.ErrEntitlementNotFound)

		err := svc.CreateWorkflow(ctx, domain.WorkflowDefinition{Name: "w", Steps: []domain.WorkflowStep{buildGrantStep("grant", entitlementId)}})
		assert.ErrorIs(t, err, domain.ErrEntitlementNotFound)
	})

	t.Run("rejects a system with several tenants when none is named", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.managedSystem.EXPECT().GetManagedSystemByName(ctx, "aws").Return(buildAwsSystem(), nil)
		mocks.managedSystem.EXPECT().ListManagedSystemTenants(ctx, "system-1").Return([]*domain.ManagedSystemTenant{buildTenant(uuid.New(), "a"), buildTenant(uuid.New(), "b")}, nil)

		err := svc.CreateWorkflow(ctx, domain.WorkflowDefinition{Name: "w", Steps: []domain.WorkflowStep{buildCreateStep("create-aws")}})
		assert.ErrorIs(t, err, domain.ErrInvalidWorkflow)
	})
}

func TestWorkflowService_StartWorkflow(t *testing.T) {
	ctx := context.Background()
	identityId := uuid.New()
	definition := domain.WorkflowDefinition{Name: "onboarding", Steps: []domain.WorkflowStep{buildCreateStep("create-aws")}}

	t.Run("starts a run of the saved workflow for the identity", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.workflowRepo.EXPECT().GetWorkflow(ctx, "onboarding").Return(definition, nil)
		mocks.identity.EXPECT().GetIdentityById(ctx, identityId).Return(&domain.Identity{}, nil)
		mocks.engine.EXPECT().Start(ctx, gomock.Any(), domain.WorkflowRun{Definition: definition, IdentityId: identityId}).Return(nil)

		runId, err := svc.StartWorkflow(ctx, "onboarding", identityId)
		require.NoError(t, err)
		assert.NotEmpty(t, runId)
	})

	t.Run("fails when the workflow does not exist", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.workflowRepo.EXPECT().GetWorkflow(ctx, "nope").Return(domain.WorkflowDefinition{}, domain.ErrWorkflowNotFound)

		_, err := svc.StartWorkflow(ctx, "nope", identityId)
		assert.ErrorIs(t, err, domain.ErrWorkflowNotFound)
	})

	t.Run("does not start a run for an unknown identity", func(t *testing.T) {
		svc, mocks := setupWorkflowService(t)
		mocks.workflowRepo.EXPECT().GetWorkflow(ctx, "onboarding").Return(definition, nil)
		mocks.identity.EXPECT().GetIdentityById(ctx, identityId).Return(nil, domain.ErrNoIdentityFound)

		_, err := svc.StartWorkflow(ctx, "onboarding", identityId)
		assert.ErrorIs(t, err, domain.ErrNoIdentityFound)
	})
}

func TestWorkflowService_GetRun(t *testing.T) {
	svc, mocks := setupWorkflowService(t)
	mocks.engine.EXPECT().Get(gomock.Any(), "run-1").Return(domain.WorkflowRunState{Id: "run-1", Name: "onboarding"}, nil).Times(2)

	state, err := svc.GetRun(context.Background(), "onboarding", "run-1")
	require.NoError(t, err)
	assert.Equal(t, "run-1", state.Id)

	_, err = svc.GetRun(context.Background(), "rehire", "run-1")
	assert.ErrorIs(t, err, domain.ErrWorkflowRunNotFound)
}
