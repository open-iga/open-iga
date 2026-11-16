package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupRouterWithMockWorkflowService(t *testing.T) (*chi.Mux, *testutil.MockWorkflowService) {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	workflowServiceMock := testutil.NewMockWorkflowService(ctrl)
	applicationMock := &contract.RuntimeApplication{WorkflowService: workflowServiceMock}

	return testutil.NewMockRouter(NewHandler(testutil.NewTestAppConfig(), testutil.NewTestLogger(), applicationMock)), workflowServiceMock
}

func sendRequest(router *chi.Mux, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

const createStepJSON = `{"name": "create-aws", "operation": {"system": "aws", "resource": "iam-user", "kind": "account-action"}}`

var createStep = domain.WorkflowStep{
	Name:      "create-aws",
	Operation: domain.WorkflowOperation{System: "aws", Resource: "iam-user", Kind: domain.AccountActionKind},
}

func TestHandler_CreateWorkflow(t *testing.T) {
	body := `{"name": "onboarding", "type": "onboarding", "steps": [` + createStepJSON + `]}`
	definition := domain.WorkflowDefinition{Name: "onboarding", Type: domain.WorkflowTypeOnboarding, Steps: []domain.WorkflowStep{createStep}}

	t.Run("returns 201 when the workflow is created", func(t *testing.T) {
		router, service := setupRouterWithMockWorkflowService(t)
		service.EXPECT().CreateWorkflow(gomock.Any(), definition).Return(nil)

		assert.Equal(t, http.StatusCreated, sendRequest(router, http.MethodPost, "/api/v1/workflows", body).Code)
	})

	t.Run("returns 409 when the name is taken", func(t *testing.T) {
		router, service := setupRouterWithMockWorkflowService(t)
		service.EXPECT().CreateWorkflow(gomock.Any(), definition).Return(domain.ErrWorkflowAlreadyExists)

		assert.Equal(t, http.StatusConflict, sendRequest(router, http.MethodPost, "/api/v1/workflows", body).Code)
	})

	for name, err := range map[string]error{
		"the workflow is invalid":    domain.ErrInvalidWorkflow,
		"the system does not exist":  domain.ErrManagedSystemNotFound,
		"the tenant does not exist":  domain.ErrTenantNotFound,
		"the entitlement is unknown": domain.ErrEntitlementNotFound,
	} {
		t.Run("returns 400 when "+name, func(t *testing.T) {
			router, service := setupRouterWithMockWorkflowService(t)
			service.EXPECT().CreateWorkflow(gomock.Any(), definition).Return(err)

			assert.Equal(t, http.StatusBadRequest, sendRequest(router, http.MethodPost, "/api/v1/workflows", body).Code)
		})
	}
}

func TestHandler_GetWorkflow(t *testing.T) {
	t.Run("returns the workflow", func(t *testing.T) {
		router, service := setupRouterWithMockWorkflowService(t)
		service.EXPECT().GetWorkflow(gomock.Any(), "onboarding").Return(domain.WorkflowDefinition{Name: "onboarding", Type: domain.WorkflowTypeOnboarding, Steps: []domain.WorkflowStep{createStep}}, nil)

		rec := sendRequest(router, http.MethodGet, "/api/v1/workflows/onboarding", "")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"name": "onboarding", "type": "onboarding", "steps": [`+createStepJSON+`]}`, rec.Body.String())
	})

	t.Run("returns 404 when there is no such workflow", func(t *testing.T) {
		router, service := setupRouterWithMockWorkflowService(t)
		service.EXPECT().GetWorkflow(gomock.Any(), "nope").Return(domain.WorkflowDefinition{}, domain.ErrWorkflowNotFound)

		assert.Equal(t, http.StatusNotFound, sendRequest(router, http.MethodGet, "/api/v1/workflows/nope", "").Code)
	})
}

func TestHandler_StartWorkflow(t *testing.T) {
	identityId := uuid.New()
	body := `{"identityId": "` + identityId.String() + `"}`

	t.Run("returns 201 with the run id", func(t *testing.T) {
		router, service := setupRouterWithMockWorkflowService(t)
		service.EXPECT().StartWorkflow(gomock.Any(), "onboarding", identityId).Return("run-1", nil)

		rec := sendRequest(router, http.MethodPost, "/api/v1/workflows/onboarding/runs", body)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.JSONEq(t, `{"id": "run-1"}`, rec.Body.String())
	})

	t.Run("returns 400 for an identity id that is not an id", func(t *testing.T) {
		router, _ := setupRouterWithMockWorkflowService(t)

		assert.Equal(t, http.StatusBadRequest, sendRequest(router, http.MethodPost, "/api/v1/workflows/onboarding/runs", `{"identityId": "x"}`).Code)
	})

	t.Run("returns 400 when no identity is sent", func(t *testing.T) {
		router, _ := setupRouterWithMockWorkflowService(t)

		assert.Equal(t, http.StatusBadRequest, sendRequest(router, http.MethodPost, "/api/v1/workflows/onboarding/runs", `{}`).Code)
	})

	for name, err := range map[string]error{"workflow": domain.ErrWorkflowNotFound, "identity": domain.ErrNoIdentityFound} {
		t.Run("returns 404 for an unknown "+name, func(t *testing.T) {
			router, service := setupRouterWithMockWorkflowService(t)
			service.EXPECT().StartWorkflow(gomock.Any(), "onboarding", identityId).Return("", err)

			assert.Equal(t, http.StatusNotFound, sendRequest(router, http.MethodPost, "/api/v1/workflows/onboarding/runs", body).Code)
		})
	}
}

func TestHandler_GetWorkflowRun(t *testing.T) {
	t.Run("returns the state and the times of every step", func(t *testing.T) {
		router, service := setupRouterWithMockWorkflowService(t)
		created := time.Date(2026, 11, 15, 10, 0, 0, 0, time.UTC)
		started := created.Add(time.Second)
		service.EXPECT().GetRun(gomock.Any(), "onboarding", "run-1").Return(domain.WorkflowRunState{Id: "run-1", Steps: []domain.WorkflowStepState{
			{Name: "create-aws", Status: domain.WorkflowStepSucceeded, CreatedAt: created, UpdatedAt: started, StartedAt: started},
			{Name: "grant-admin", Status: domain.WorkflowStepFailed, Error: "boom", CreatedAt: created, UpdatedAt: created},
		}}, nil)

		rec := sendRequest(router, http.MethodGet, "/api/v1/workflows/onboarding/runs/run-1", "")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"id": "run-1", "steps": [
			{"name": "create-aws", "status": "succeeded", "createdAt": "2026-11-15T10:00:00Z", "updatedAt": "2026-11-15T10:00:01Z", "startedAt": "2026-11-15T10:00:01Z"},
			{"name": "grant-admin", "status": "failed", "error": "boom", "createdAt": "2026-11-15T10:00:00Z", "updatedAt": "2026-11-15T10:00:00Z"}
		]}`, rec.Body.String())
	})

	t.Run("returns 404 for an unknown run", func(t *testing.T) {
		router, service := setupRouterWithMockWorkflowService(t)
		service.EXPECT().GetRun(gomock.Any(), "onboarding", "run-1").Return(domain.WorkflowRunState{}, domain.ErrWorkflowRunNotFound)

		assert.Equal(t, http.StatusNotFound, sendRequest(router, http.MethodGet, "/api/v1/workflows/onboarding/runs/run-1", "").Code)
	})
}
