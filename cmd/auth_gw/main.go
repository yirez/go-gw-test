package main

import (
	"fmt"
	g "go-gw-test/cmd/auth_gw/internal/globals"
	"go-gw-test/pkg/rest_qol"

	"go.uber.org/zap"
)

// main initializes configuration, logger, and starts the auth_gw HTTP server.
func main() {
	g.InitConfiguration()

	router := NewRouter()
	err := rest_qol.RunHTTPServer(fmt.Sprintf(":%d", g.Cfg.StandardConfigs.Port), router)
	if err != nil {
		zap.L().Error("server shutdown", zap.Error(err))
	}
}
