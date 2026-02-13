package main

import (
	"fmt"
	g "go-gw-test/cmd/orders_gw/internal/globals"
	"go-gw-test/pkg/rest_qol"

	"go.uber.org/zap"
)

//go:generate swag init -g main.go -o docs --parseDependency --parseInternal
// @title Orders Gateway API
// @version 1.0
// @description Read-only orders gateway for order and item retrieval.
// @BasePath /

// main initializes configuration and starts the orders_gw HTTP server.
func main() {
	g.InitConfiguration()

	router := NewRouter()

	address := fmt.Sprintf(":%d", g.Cfg.StandardConfigs.Port)
	err := rest_qol.RunHTTPServer(address, router)
	if err != nil {
		zap.L().Error("server shutdown", zap.Error(err))
	}
}
