package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/open-iga/core/internal/api"
	"github.com/open-iga/core/internal/application"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/remote"
	"github.com/open-iga/core/internal/repository"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const banner = "\n" +
	" ██████╗ ██████╗ ███████╗███╗   ██╗    ██╗ ██████╗  █████╗ \n" +
	"██╔═══██╗██╔══██╗██╔════╝████╗  ██║    ██║██╔════╝ ██╔══██╗\n" +
	"██║   ██║██████╔╝█████╗  ██╔██╗ ██║    ██║██║  ███╗███████║\n" +
	"██║   ██║██╔═══╝ ██╔══╝  ██║╚██╗██║    ██║██║   ██║██╔══██║\n" +
	"╚██████╔╝██║     ███████╗██║ ╚████║    ██║╚██████╔╝██║  ██║\n" +
	" ╚═════╝ ╚═╝     ╚══════╝╚═╝  ╚═══╝    ╚═╝ ╚═════╝ ╚═╝  ╚═╝\n" +
	"                                                           \n"

func main() {
	fmt.Print(banner)

	appConfig := common.NewAppConfig()

	ctx := context.Background()
	logger, shutdownOTel, err := common.SetupOTel(ctx, appConfig.Telemetry)
	if err != nil {
		slog.Error("failed to set up OpenTelemetry", "err", err)
		os.Exit(1)
	}
	defer func() { _ = shutdownOTel(ctx) }()

	runtimeRepository, conn, err := repository.NewRepository(appConfig, logger)
	if err != nil {
		panic(err)
	}

	runtimeRemote := remote.NewRemote(appConfig, logger)
	runtimeApplication, err := application.NewApplication(ctx, appConfig, logger, runtimeRemote, runtimeRepository)
	if err != nil {
		logger.Error("failed to create application", "err", err)
		os.Exit(1)
	}

	router := api.NewRouter(appConfig, logger, runtimeApplication)

	// TODO: throwaway local test route for connector execution test
	mux := http.NewServeMux()
	mux.Handle("/", router)
	mux.HandleFunc("POST /debug/execute-create", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ConnectorName   string `json:"connectorName"`
			ManagedResource string `json:"managedResource"`
			TenantId        string `json:"tenantId"`
			IdentityId      string `json:"identityId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var tenantId uuid.UUID // stays empty when no tenantId is sent
		var tenantErr error
		if body.TenantId != "" {
			tenantId, tenantErr = uuid.Parse(body.TenantId)
		}
		identityId, identityErr := uuid.Parse(body.IdentityId)
		if err := errors.Join(tenantErr, identityErr); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		out, err := runtimeApplication.ConnectorService.Execute(r.Context(), body.ConnectorName, body.ManagedResource, domain.AccountActionKind, "create",
			domain.OperationRef{TenantId: tenantId, IdentityId: identityId, EntitlementId: uuid.Nil})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})

	server := &http.Server{
		Addr:    appConfig.Port,
		Handler: otelhttp.NewHandler(mux, "open-iga-core"),
		// Below timeouts are arbitrary
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	defer func() {
		logger.Info("Gracefully shutting down")
		conn.Close()
		logger.Info("Closed DB connection")
		err = server.Close()
		if err != nil {
			logger.Error("Error closing server")
		} else {
			logger.Info("Server shutdown gracefully")
		}
	}()

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
