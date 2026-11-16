package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/open-iga/core/internal/api"
	"github.com/open-iga/core/internal/application"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/remote"
	"github.com/open-iga/core/internal/repository"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const (
	// workflowShutdownTimeout is how long running steps get to finish on shutdown; the rest resume on the next start.
	workflowShutdownTimeout = 10 * time.Second
	// serverShutdownTimeout is how long requests in flight get to finish on shutdown.
	serverShutdownTimeout = 10 * time.Second
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

	server := &http.Server{
		Addr:    appConfig.Port,
		Handler: otelhttp.NewHandler(router, "open-iga-core"),
		// Below timeouts are arbitrary
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	defer shutDown(logger, server, runtimeRemote.WorkflowEngine.Shutdown, conn.Close)

	waitForStop(logger, server)
}

// waitForStop serves until the server fails or a stop signal arrives; the deferred shutdown runs only when main returns, so the signal has to end the wait.
func waitForStop(logger *slog.Logger, server *http.Server) {
	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() { serverErr <- server.ListenAndServe() }()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	case received := <-stopSignal:
		logger.Info("Received a stop signal", "signal", received.String())
	}
}

// shutDown stops taking requests first, then lets the workflow steps finish, then closes the database.
func shutDown(logger *slog.Logger, server *http.Server, shutdownWorkflowEngine func(time.Duration), closeDatabase func()) {
	logger.Info("Gracefully shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error shutting down server", "err", err)
	} else {
		logger.Info("Server shutdown gracefully")
	}

	shutdownWorkflowEngine(workflowShutdownTimeout)

	closeDatabase()
	logger.Info("Closed DB connection")
}
