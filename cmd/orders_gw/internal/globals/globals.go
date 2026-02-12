package globals

import (
	"fmt"
	"go-gw-test/cmd/orders_gw/internal/types"
	"go-gw-test/pkg/configuration_manager"
	cmt "go-gw-test/pkg/configuration_manager/types"
	"os"

	"go.uber.org/zap"
)

// Cfg holds global configuration for orders_gw.
var Cfg types.AppConfig

// InitConfiguration loads standard configs and initializes global logger.
func InitConfiguration() {
	var err error
	Cfg.StandardConfigs, err = configuration_manager.InitStandardConfigs(
		cmt.InitChecklist{
			DB:              true,
			Redis:           false,
			AutoMigrateList: []any{&types.OrderRecord{}, &types.OrderItem{}},
		})
	if err != nil {
		fmt.Printf("failed init configs: %v\n", err)
		os.Exit(1)
	}

	zap.ReplaceGlobals(Cfg.StandardConfigs.Clients.Logger)
}
