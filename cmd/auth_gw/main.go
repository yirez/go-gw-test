package main

import (
	"fmt"
	"go-gw-test/cmd/auth_gw/internal/appservice"
	"go-gw-test/pkg/rest_qol"

	"go-gw-test/pkg/configuration_manager"

	"go.uber.org/zap"
)

// main initializes configuration, logger, and starts the auth_gw HTTP server.
func main() {
	configPath := "config.hcl"

	cfg, err := configuration_manager.InitStandardConfigs(configPath)
	if err != nil {
		fmt.Printf("init configs: %v\n", err)
		return
	}
	zap.ReplaceGlobals(cfg.Clients.Logger)

	app := appservice.NewAppService(cfg)

	err = rest_qol.RunHTTPServer(app.Address(), app.Router())
	if err != nil {
		zap.L().Error("server shutdown", appservice.WrapError(err))
	}
}
