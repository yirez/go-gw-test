package types

import cmt "go-gw-test/pkg/configuration_manager/types"

type AppConfig struct {
	JwtSigningKey   []byte
	StandardConfigs cmt.StandardConfig
}
