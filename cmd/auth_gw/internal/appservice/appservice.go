package appservice

import (
	"fmt"
	"go-gw-test/cmd/auth_gw/internal/router"
	cmt "go-gw-test/pkg/configuration_manager/types"
	"net/http"
	"time"

	"go-gw-test/cmd/auth_gw/internal/handler"
	"go-gw-test/cmd/auth_gw/internal/repo"
	"go-gw-test/cmd/auth_gw/internal/usecase"
)

// AppService wires dependencies and exposes HTTP router and server address.
type AppService struct {
	cfg    cmt.StandardConfig
	router http.Handler
}

// NewAppService builds an AppService with all auth_gw components.
func NewAppService(cfg cmt.StandardConfig) *AppService {
	authRepo := repo.NewAuthRepo(cfg.Clients.DB)
	authUsecase := usecase.NewAuthUseCase(authRepo, jwtSigningKey(), time.Hour)
	authHandler := handler.NewAuthHandler(authUsecase)
	loggingHandler := handler.NewLoggingHandler()

	appRouter := router.NewRouter(authHandler, loggingHandler)

	return &AppService{
		cfg:    cfg,
		router: appRouter,
	}
}

// Router returns the root HTTP handler.
func (s *AppService) Router() http.Handler {
	return s.router
}

// Address returns the host:port address for the HTTP server.
func (s *AppService) Address() string {
	return fmt.Sprintf(":%d", s.cfg.Port)
}

func jwtSigningKey() []byte {
	return []byte(time.Now().UTC().Format(time.RFC3339))
}
