package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	_ "github.com/joho/godotenv/autoload"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"github.com/yorukot/blind-party/internal/config"
	"github.com/yorukot/blind-party/internal/middleware"
	"github.com/yorukot/blind-party/internal/router"
	"github.com/yorukot/blind-party/pkg/logger"
	"github.com/yorukot/blind-party/pkg/response"
)

// @version 1.0
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
func main() {
	logger.InitLogger()

	_, err := config.InitConfig()
	if err != nil {
		zap.L().Fatal("Error initializing config", zap.Error(err))
		return
	}

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "https://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Requested-With", "Upgrade", "Connection", "Sec-WebSocket-Key", "Sec-WebSocket-Version", "Sec-WebSocket-Protocol"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.ZapLoggerMiddleware(zap.L()))
	r.Use(chiMiddleware.StripSlashes)

	setupRouter(r)

	zap.L().Info("Starting server on http://localhost:" + config.Env().Port)
	zap.L().Info("Environment: " + string(config.Env().AppEnv))

	err = http.ListenAndServe(":"+config.Env().Port, r)
	if err != nil {
		zap.L().Fatal("Failed to start server", zap.Error(err))
	}
}

// setupRouter sets up the router
func setupRouter(r chi.Router) {
	r.Route("/api", func(r chi.Router) {
		router.GameRouter(r)
	})

	if config.Env().AppEnv == config.AppEnvDev {
		r.Get("/swagger/*", httpSwagger.WrapHandler)
	}

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Not found handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.RespondWithError(w, http.StatusNotFound, "Not Found", "NOT_FOUND")
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.RespondWithError(w, http.StatusMethodNotAllowed, "Method Not Allowed", "METHOD_NOT_ALLOWED")
	})

	zap.L().Info("Router setup complete")
}
