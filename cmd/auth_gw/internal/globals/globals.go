package globals

import (
	"fmt"
	"go-gw-test/cmd/auth_gw/internal/types"
	"go-gw-test/pkg/configuration_manager"
	cmt "go-gw-test/pkg/configuration_manager/types"
	"os"
	"time"

	"go.uber.org/zap"
)

var Cfg types.AppConfig

func InitConfiguration() {
	var err error
	Cfg.StandardConfigs, err = configuration_manager.InitStandardConfigs(
		cmt.InitChecklist{
			DB:              true,
			Redis:           false,
			AutoMigrateList: []any{&types.UserRecord{}, &types.ServiceRecord{}},
		})
	if err != nil {
		fmt.Printf("failed init configs: %v\n", err)
		os.Exit(1)
	}
	zap.ReplaceGlobals(Cfg.StandardConfigs.Clients.Logger)

	// Signing key setup by current time
	Cfg.JwtSigningKey = []byte(time.Now().UTC().Format(time.RFC3339))
}
