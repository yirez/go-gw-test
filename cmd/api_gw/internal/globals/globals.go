package globals

import (
	"fmt"
	"github.com/yirez/go-gw-test/cmd/api_gw/internal/types"
	"github.com/yirez/go-gw-test/pkg/configuration_manager"
	cmt "github.com/yirez/go-gw-test/pkg/configuration_manager/types"
	"os"

	"go.uber.org/zap"
)

// Cfg holds global configuration for api_gw.
var Cfg types.AppConfig

// InitConfiguration loads configs and initializes global logger.
func InitConfiguration() {
	var err error
	Cfg.StandardConfigs, err = configuration_manager.InitStandardConfigs(
		cmt.InitChecklist{
			Auth:            true,
			Redis:           true,
			AutoMigrateList: nil,
		})
	if err != nil {
		fmt.Printf("failed init configs: %v\n", err)
		os.Exit(1)
	}

	err = configuration_manager.ReadCustomConfig("endpoint_configuration", &Cfg.EndpointConfiguration)
	if err != nil {
		fmt.Printf("failed load endpoint configuration: %v\n", err)
		os.Exit(1)
	}

	zap.ReplaceGlobals(Cfg.StandardConfigs.Clients.Logger)
}
