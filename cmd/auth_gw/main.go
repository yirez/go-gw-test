package main

import (
	"fmt"

	"go-gw-test/cmd/auth_gw/internal/appservice"
	"go-gw-test/cmd/auth_gw/internal/types"
	"go-gw-test/pkg/configuration_manager"
	cmt "go-gw-test/pkg/configuration_manager/types"

	"go-gw-test/pkg/rest_qol"

	"go.uber.org/zap"
)

// main initializes configuration, logger, and starts the auth_gw HTTP server.
func main() {

	cfg, err := configuration_manager.InitStandardConfigs(
		cmt.InitChecklist{
			DB:              true,
			Redis:           false,
			AutoMigrateList: []any{&types.UserRecord{}, &types.ServiceRecord{}},
		})
	if err != nil {
		fmt.Printf("failed init configs: %v\n", err)
		return
	}
	zap.ReplaceGlobals(cfg.Clients.Logger)

	app := appservice.NewAppService(cfg)

	err = rest_qol.RunHTTPServer(app.Address(), app.Router())
	if err != nil {
		zap.L().Error("server shutdown", zap.Error(err))
	}
}
