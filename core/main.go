package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/open-iga/core/internal/api"
	"github.com/open-iga/core/internal/application"
	"github.com/open-iga/core/internal/common"
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
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)
	fmt.Print(banner)

	appConfig := common.NewAppConfig()

	runtimeRepository, err := repository.NewRepository(appConfig, logger)

	if err != nil {
		panic(err)
	}
	runtimeRemote := remote.NewRemote(appConfig, logger)
	runtimeApplication := application.NewApplication(appConfig, logger, runtimeRemote, runtimeRepository)

	router := api.NewRouter(appConfig, logger, runtimeApplication)

	err = http.ListenAndServe(appConfig.Port, router)
	if err != nil {
		panic(err)
	}
}
