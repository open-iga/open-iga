package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/open-iga/core/internal/api"
	"github.com/open-iga/core/internal/application"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/domain"
	"github.com/open-iga/core/internal/remote"
	"github.com/open-iga/core/internal/repository"
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
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	fmt.Print(banner)

	appConfig := common.NewAppConfig()

	runtimeRepository, conn, err := repository.NewRepository(appConfig, logger)
	if err != nil {
		panic(err)
	}

	hasAdmin, err := runtimeRepository.IdentityRepository.HasAdmin(context.Background())
	if err != nil {
		logger.Error("failed to check if admin role exists", "err", err)
		os.Exit(1)
	}

	if !hasAdmin && appConfig.AdminUser.Email != "" {
		_, err = runtimeRepository.IdentityRepository.FindOrCreateWithRole(context.Background(), domain.NewOauthUser("", "", appConfig.AdminUser.Email), domain.AdminRole)
		if err != nil {
			logger.Error("failed to onboard admin", "err", err)
			os.Exit(1)
		}
	}

	if !hasAdmin && appConfig.AdminUser.Email == "" {
		logger.Error("ADMIN_USER is required to start the application. Make sure to set a valid email")
		os.Exit(1)
	}

	runtimeRemote := remote.NewRemote(appConfig, logger)
	runtimeApplication := application.NewApplication(appConfig, logger, runtimeRemote, runtimeRepository)

	router := api.NewRouter(appConfig, logger, runtimeApplication)

	server := &http.Server{
		Addr:    appConfig.Port,
		Handler: router,
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
