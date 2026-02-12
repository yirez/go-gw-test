package main

import (
	"fmt"
	g "go-gw-test/cmd/api_gw/internal/globals"
	"go-gw-test/pkg/rest_qol"

	"go.uber.org/zap"
)

// main initializes configuration and starts the api_gw HTTP server.
func main() {
	g.InitConfiguration()

	router := NewRouter()
	address := fmt.Sprintf(":%d", g.Cfg.StandardConfigs.Port)
	err := rest_qol.RunHTTPServer(address, router)
	if err != nil {
		zap.L().Error("server shutdown", zap.Error(err))
	}
}
