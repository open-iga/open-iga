package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/open-iga/core/internal/api"
	"github.com/open-iga/core/internal/application"
	"github.com/open-iga/core/internal/common"
	"github.com/open-iga/core/internal/remote"
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

	runtimeRemote := remote.NewRemote(appConfig, logger)
	runtimeApplication := application.NewApplication(appConfig, runtimeRemote, logger)

	router := api.NewRouter(appConfig, runtimeApplication, logger)
	router.Start()
}
