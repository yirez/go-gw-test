package main

import (
	"fmt"
	"go-gw-test/pkg/rest_qol"

	"go-gw-test/internal/auth_gw/appservice"
	"go-gw-test/pkg/configuration_manager"

	"go.uber.org/zap"
)

// main initializes configuration, logger, and starts the auth_gw HTTP server.
func main() {
	configPath := "config.hcl"

	cfg, logger, db, err := configuration_manager.InitStandardConfigs(configPath)
	if err != nil {
		fmt.Printf("init configs: %v\n", err)
		return
	}
	zap.ReplaceGlobals(logger)

	app := appservice.NewAppService(cfg, db)

	err = rest_qol.RunHTTPServer(app.Address(), app.Router())
	if err != nil {
		zap.L().Error("server shutdown", appservice.WrapError(err))
	}
}
