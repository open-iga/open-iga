package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupConnectorServiceWithMocks(t *testing.T) (*ConnectorService, *testutil.MockConnectorRuntime, *testutil.MockManagedSystemRepository) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	runtime := testutil.NewMockConnectorRuntime(ctrl)
	repo := testutil.NewMockManagedSystemRepository(ctrl)
	return NewConnectorService(runtime, repo, nil, nil, nil, testutil.NewTestLogger()), runtime, repo
}

func TestConnectorService_ValidateByUrl(t *testing.T) {
	t.Run("returns an id and records success once validation completes", func(t *testing.T) {
		cs, runtime, _ := setupConnectorServiceWithMocks(t)
		spec := &domain.ConnectorSpec{Name: "aws"}
		runtime.EXPECT().ValidateConnectorByURL(gomock.Any(), "https://conn", "hash").Return(spec, nil)

		id := cs.ValidateByUrl(context.TODO(), "https://conn", "hash")
		assert.NotEmpty(t, id)

		// validation runs in a goroutine; wait for the terminal state
		assert.Eventually(t, func() bool {
			res, ok := cs.GetOnboardedConnectorDetails(context.TODO(), id)
			return ok && res.Status == domain.ConnectorValidationSuccess
		}, time.Second, 5*time.Millisecond)

		res, _ := cs.GetOnboardedConnectorDetails(context.TODO(), id)
		assert.Equal(t, spec, res.ConnectorSpec)
		assert.NoError(t, res.Error)
	})

	t.Run("records failure when validation errors", func(t *testing.T) {
		cs, runtime, _ := setupConnectorServiceWithMocks(t)
		runtime.EXPECT().ValidateConnectorByURL(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("boom"))

		id := cs.ValidateByUrl(context.TODO(), "https://conn", "hash")

		assert.Eventually(t, func() bool {
			res, ok := cs.GetOnboardedConnectorDetails(context.TODO(), id)
			return ok && res.Status == domain.ConnectorValidationFailed
		}, time.Second, 5*time.Millisecond)

		res, _ := cs.GetOnboardedConnectorDetails(context.TODO(), id)
		assert.ErrorContains(t, res.Error, "boom")
		assert.Nil(t, res.ConnectorSpec)
	})
}

func TestConnectorService_LoadConnectorsOnStartup(t *testing.T) {
	t.Run("compiles connectors and skips ones with no spec", func(t *testing.T) {
		cs, runtime, repo := setupConnectorServiceWithMocks(t)
		required := true
		t.Setenv("OPEN_IGA_LOAD_ENV", "set")

		okSpec := &domain.ConnectorSpec{
			Name:   "ok",
			Config: []domain.Config{{Name: "OPEN_IGA_LOAD_ENV", Required: &required}},
		}

		systems := []*domain.ManagedSystem{
			{Id: "ok-id", Name: "ok", ConnectorUrl: "https://ok", ConnectorHash: "h1", ConnectorSpec: okSpec},
			{Id: "no-spec-id", Name: "no-spec", ConnectorUrl: "https://ns", ConnectorHash: "h4", ConnectorSpec: nil},
		}

		repo.EXPECT().ListManagedSystems(gomock.Any()).Return(systems, nil)
		// no-spec is skipped before compile; only ok is compiled.
		runtime.EXPECT().CompileConnector(gomock.Any(), "https://ok", "h1", gomock.Any())

		assert.NoError(t, cs.LoadConnectorsOnStartup(context.TODO()))
	})

	t.Run("aborts startup when a required env var is missing", func(t *testing.T) {
		cs, _, repo := setupConnectorServiceWithMocks(t)
		required := true
		spec := &domain.ConnectorSpec{
			Name:   "missing-env",
			Config: []domain.Config{{Name: "OPEN_IGA_LOAD_UNSET_ENV", Required: &required}},
		}
		repo.EXPECT().ListManagedSystems(gomock.Any()).Return([]*domain.ManagedSystem{
			{Id: "me-id", Name: "missing-env", ConnectorUrl: "https://me", ConnectorHash: "h3", ConnectorSpec: spec},
		}, nil)

		err := cs.LoadConnectorsOnStartup(context.TODO())

		assert.ErrorContains(t, err, "OPEN_IGA_LOAD_UNSET_ENV")
	})
}

func TestConnectorService_GetOnboardedConnectorDetails(t *testing.T) {
	t.Run("returns false for an unknown id", func(t *testing.T) {
		cs, _, _ := setupConnectorServiceWithMocks(t)

		res, ok := cs.GetOnboardedConnectorDetails(context.TODO(), "does-not-exist")

		assert.False(t, ok)
		assert.Nil(t, res)
	})

	t.Run("computes missing required env vars against the core environment on read", func(t *testing.T) {
		cs, _, _ := setupConnectorServiceWithMocks(t)
		required := true
		cs.onboardingResults["id"] = domain.ConnectorValidationResult{
			Status: domain.ConnectorValidationSuccess,
			ConnectorSpec: &domain.ConnectorSpec{
				Config: []domain.Config{{Name: "OPEN_IGA_TEST_UNSET_ENV", Required: &required}},
			},
		}

		res, ok := cs.GetOnboardedConnectorDetails(context.TODO(), "id")

		assert.True(t, ok)
		assert.Equal(t, []string{"OPEN_IGA_TEST_UNSET_ENV"}, res.ConnectorSpec.MissingRequiredEnvVars())
	})
}

func TestConnectorService_Execute_BuildsInputFromDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	runtime := testutil.NewMockConnectorRuntime(ctrl)
	systems := testutil.NewMockManagedSystemRepository(ctrl)
	identities := testutil.NewMockIdentityRepository(ctrl)
	accounts := testutil.NewMockAccountRepository(ctrl)
	service := NewConnectorService(runtime, systems, identities, accounts, testutil.NewMockEntitlementRepository(ctrl), testutil.NewTestLogger())

	tenantID, identityID := uuid.New(), uuid.New()
	systems.EXPECT().GetManagedSystemByName(gomock.Any(), "aws").Return(&domain.ManagedSystem{
		Id: "ms1", ConnectorUrl: "u", ConnectorHash: "h",
	}, nil)
	systems.EXPECT().GetManagedSystemTenant(gomock.Any(), tenantID).Return(&domain.ManagedSystemTenant{Id: tenantID.String(), ManagedSystemId: "ms1", ExternalId: "acct-1"}, nil)
	identities.EXPECT().GetIdentityById(gomock.Any(), identityID).Return(&domain.Identity{Email: "a@b.c", FirstName: "Ada", LastName: "L"}, nil)
	runtime.EXPECT().Execute(gomock.Any(), "u", "h", domain.ConnectorInvocation{
		Kind: domain.AccountActionKind, ManagedResource: "iam-user", Type: "create",
		Input: json.RawMessage(`{"tenant":{"id":"acct-1"},"email":"a@b.c","firstname":"Ada","lastname":"L"}`),
	}).Return([]byte(`{"accountId":"arn:1","created":true}`), nil)

	output, err := service.Execute(context.Background(), "aws", "iam-user", domain.AccountActionKind, "create", domain.OperationRef{TenantId: tenantID, IdentityId: identityID})

	assert.NoError(t, err)
	assert.Equal(t, domain.AccountCreateOutput{AccountID: "arn:1", Created: true}, output)
}

func TestConnectorService_Execute_RejectsTenantOfAnotherSystem(t *testing.T) {
	service, _, systems := setupConnectorServiceWithMocks(t)
	tenantID := uuid.New()
	systems.EXPECT().GetManagedSystemByName(gomock.Any(), "aws").Return(&domain.ManagedSystem{
		Id: "ms1",
	}, nil)
	systems.EXPECT().GetManagedSystemTenant(gomock.Any(), tenantID).Return(&domain.ManagedSystemTenant{ManagedSystemId: "other"}, nil)

	_, err := service.Execute(context.Background(), "aws", "iam-user", domain.AccountActionKind, "list", domain.OperationRef{TenantId: tenantID})

	assert.ErrorIs(t, err, domain.ErrTenantNotFound)
}

func TestConnectorService_Execute_SendsEmptyTenantWhenNoneGiven(t *testing.T) {
	ctrl := gomock.NewController(t)
	runtime := testutil.NewMockConnectorRuntime(ctrl)
	systems := testutil.NewMockManagedSystemRepository(ctrl)
	service := NewConnectorService(runtime, systems, nil, nil, nil, testutil.NewTestLogger())

	systems.EXPECT().GetManagedSystemByName(gomock.Any(), "aws").Return(&domain.ManagedSystem{Id: "ms1", ConnectorUrl: "u", ConnectorHash: "h"}, nil)
	runtime.EXPECT().Execute(gomock.Any(), "u", "h", domain.ConnectorInvocation{
		Kind: domain.AccountActionKind, ManagedResource: "iam-user", Type: "list",
		Input: json.RawMessage(`{"tenant":{"id":""}}`),
	}).Return([]byte(`{"accounts":[]}`), nil)

	_, err := service.Execute(context.Background(), "aws", "iam-user", domain.AccountActionKind, "list", domain.OperationRef{})

	assert.NoError(t, err)
}

func TestConnectorService_Execute_GrantUsesAccountAndEntitlementFromDB(t *testing.T) {
	ctrl := gomock.NewController(t)
	runtime := testutil.NewMockConnectorRuntime(ctrl)
	systems := testutil.NewMockManagedSystemRepository(ctrl)
	identities := testutil.NewMockIdentityRepository(ctrl)
	accounts := testutil.NewMockAccountRepository(ctrl)
	entitlements := testutil.NewMockEntitlementRepository(ctrl)
	service := NewConnectorService(runtime, systems, identities, accounts, entitlements, testutil.NewTestLogger())

	tenantID, identityID, entitlementID := uuid.New(), uuid.New(), uuid.New()
	systems.EXPECT().GetManagedSystemByName(gomock.Any(), "aws").Return(&domain.ManagedSystem{Id: "ms1", ConnectorUrl: "u", ConnectorHash: "h"}, nil)
	systems.EXPECT().GetManagedSystemTenant(gomock.Any(), tenantID).Return(&domain.ManagedSystemTenant{Id: tenantID.String(), ManagedSystemId: "ms1", ExternalId: "acct-1"}, nil)
	identities.EXPECT().GetIdentityById(gomock.Any(), identityID).Return(&domain.Identity{}, nil)
	accounts.EXPECT().GetAccountByTenantAndIdentity(gomock.Any(), tenantID, identityID).Return(&domain.Account{ExternalId: "arn:user"}, nil)
	entitlements.EXPECT().GetEntitlement(gomock.Any(), entitlementID).Return(&domain.Entitlement{TenantId: tenantID.String(), Name: "group-1"}, nil)
	runtime.EXPECT().Execute(gomock.Any(), "u", "h", domain.ConnectorInvocation{
		Kind: domain.EntitlementKind, ManagedResource: "iam-user-groups", Type: "grant",
		Input: json.RawMessage(`{"tenant":{"id":"acct-1"},"accountId":"arn:user","entitlementId":"group-1"}`),
	}).Return([]byte(`{"granted":true}`), nil)

	output, err := service.Execute(context.Background(), "aws", "iam-user-groups", domain.EntitlementKind, "grant",
		domain.OperationRef{TenantId: tenantID, IdentityId: identityID, EntitlementId: entitlementID})

	assert.NoError(t, err)
	assert.Equal(t, domain.EntitlementGrantOutput{Granted: true}, output)
}

func TestConnectorService_Execute_RejectsEntitlementOfAnotherTenant(t *testing.T) {
	ctrl := gomock.NewController(t)
	systems := testutil.NewMockManagedSystemRepository(ctrl)
	identities := testutil.NewMockIdentityRepository(ctrl)
	accounts := testutil.NewMockAccountRepository(ctrl)
	entitlements := testutil.NewMockEntitlementRepository(ctrl)
	service := NewConnectorService(testutil.NewMockConnectorRuntime(ctrl), systems, identities, accounts, entitlements, testutil.NewTestLogger())

	tenantID, identityID, entitlementID := uuid.New(), uuid.New(), uuid.New()
	systems.EXPECT().GetManagedSystemByName(gomock.Any(), "aws").Return(&domain.ManagedSystem{Id: "ms1"}, nil)
	systems.EXPECT().GetManagedSystemTenant(gomock.Any(), tenantID).Return(&domain.ManagedSystemTenant{Id: tenantID.String(), ManagedSystemId: "ms1"}, nil)
	identities.EXPECT().GetIdentityById(gomock.Any(), identityID).Return(&domain.Identity{}, nil)
	accounts.EXPECT().GetAccountByTenantAndIdentity(gomock.Any(), tenantID, identityID).Return(&domain.Account{ExternalId: "arn:user"}, nil)
	entitlements.EXPECT().GetEntitlement(gomock.Any(), entitlementID).Return(&domain.Entitlement{TenantId: uuid.NewString()}, nil)

	_, err := service.Execute(context.Background(), "aws", "iam-user-groups", domain.EntitlementKind, "grant",
		domain.OperationRef{TenantId: tenantID, IdentityId: identityID, EntitlementId: entitlementID})

	assert.ErrorIs(t, err, domain.ErrEntitlementNotFound)
}
