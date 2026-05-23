package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mcv0000/storeflow-api/internal/auth"
	"github.com/mcv0000/storeflow-api/internal/cache"
	"github.com/mcv0000/storeflow-api/internal/config"
	"github.com/mcv0000/storeflow-api/internal/db"
	"github.com/mcv0000/storeflow-api/internal/graphql"
	orderdomain "github.com/mcv0000/storeflow-api/internal/order"
	"github.com/mcv0000/storeflow-api/internal/platform/logger"
	httpmiddleware "github.com/mcv0000/storeflow-api/internal/platform/middleware"
	"github.com/mcv0000/storeflow-api/internal/product"
	"github.com/mcv0000/storeflow-api/internal/store"
	"github.com/mcv0000/storeflow-api/internal/user"
)

func main() {
	logger.Init()
	logger.InfoJSON("starting StoreFlow API", nil)

	cfg, err := config.Load()
	if err != nil {
		logger.Error.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	logger.InfoJSON("connecting to PostgreSQL", nil)
	dbPool, err := db.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	logger.InfoJSON("connected to PostgreSQL", nil)

	logger.InfoJSON("connecting to Redis", nil)
	redisClient, err := cache.NewRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		logger.Error.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	logger.InfoJSON("connected to Redis", nil)

	userRepo := user.NewPostgresRepository(dbPool)
	storeRepo := store.NewPostgresRepository(dbPool)
	productRepo := product.NewPostgresRepository(dbPool)
	orderRepo := orderdomain.NewPostgresRepository(dbPool)

	userService := user.NewService(userRepo, cfg.JWTSecret)
	storeService := store.NewService(storeRepo)
	productService := product.NewService(productRepo, storeRepo, redisClient)
	orderService := orderdomain.NewService(orderRepo, storeRepo, productService)

	resolver := graphql.NewResolver(userService, storeService, productService, orderService)

	mux := http.NewServeMux()

	mux.Handle("/graphql", auth.Middleware(cfg.JWTSecret)(graphql.NewHandler(resolver)))
	mux.Handle("/playground", graphql.NewPlaygroundHandler())

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := httpmiddleware.RequestID(httpmiddleware.Logging(mux))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.InfoJSON("server listening", logger.Fields{
			"port": cfg.Port,
		})

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.InfoJSON("shutting down server", nil)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.InfoJSON("server stopped", nil)
}
