package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/contract"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/testutil"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestHandler_SearchEntitlements(t *testing.T) {
	setup := func(t *testing.T) (http.Handler, *testutil.MockDiscoveryService) {
		t.Helper()
		ctrl := gomock.NewController(t)
		discovery := testutil.NewMockDiscoveryService(ctrl)
		handler := NewHandler(testutil.NewTestAppConfig(), testutil.NewTestLogger(), &contract.RuntimeApplication{DiscoveryService: discovery})

		return testutil.NewMockRouter(handler), discovery
	}
	systemId, tenantId := uuid.New(), uuid.New()

	t.Run("passes the filters and the page to the search and returns the page with its total", func(t *testing.T) {
		router, discovery := setup(t)
		discovery.EXPECT().SearchEntitlements(gomock.Any(), domain.EntitlementSearch{Query: "adm", ManagedSystemId: systemId, TenantId: tenantId, Resource: "role", Limit: 5, Offset: 10}).
			Return([]*domain.EntitlementCatalog{{Id: "e-1", ManagedSystemId: systemId.String(), TenantId: tenantId.String(), Resource: "role", ExternalId: "role-admin", DisplayName: "Admin", Description: ""}}, int64(11), nil)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/entitlements?q=adm&systemId="+systemId.String()+"&tenantId="+tenantId.String()+"&resource=role&limit=5&offset=10", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"total": 11, "entitlements": [{"id": "e-1", "managedSystemId": "`+systemId.String()+`", "tenantId": "`+tenantId.String()+`", "resource": "role", "externalId": "role-admin", "displayName": "Admin"}]}`, rec.Body.String())
	})

	t.Run("uses the default page size when none is given", func(t *testing.T) {
		router, discovery := setup(t)
		discovery.EXPECT().SearchEntitlements(gomock.Any(), domain.EntitlementSearch{Limit: 20}).Return(nil, int64(0), nil)

		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/entitlements", nil))

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.JSONEq(t, `{"total": 0, "entitlements": []}`, rec.Body.String())
	})

	t.Run("refuses a bad id or page", func(t *testing.T) {
		for _, query := range []string{"systemId=x", "tenantId=x", "limit=0", "limit=101", "offset=-1"} {
			router, _ := setup(t)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/entitlements?"+query, nil))

			assert.Equal(t, http.StatusBadRequest, rec.Code, query)
		}
	})
}
