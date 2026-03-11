package globals

import (
	"fmt"
	"github.com/yirez/go-gw-test/cmd/users_gw/internal/types"
	"github.com/yirez/go-gw-test/pkg/configuration_manager"
	cmt "github.com/yirez/go-gw-test/pkg/configuration_manager/types"
	"os"

	"go.uber.org/zap"
)

// Cfg holds global configuration for users_gw.
var Cfg types.AppConfig

// InitConfiguration loads standard configs and initializes global logger.
func InitConfiguration() {
	var err error
	Cfg.StandardConfigs, err = configuration_manager.InitStandardConfigs(
		cmt.InitChecklist{
			DB:              true,
			Redis:           false,
			AutoMigrateList: []any{&types.UserProfile{}, &types.UserContactInfo{}},
		})
	if err != nil {
		fmt.Printf("failed init configs: %v\n", err)
		os.Exit(1)
	}

	zap.ReplaceGlobals(Cfg.StandardConfigs.Clients.Logger)
}
