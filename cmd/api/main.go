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
	"github.com/mcv0000/storeflow-api/internal/product"
	"github.com/mcv0000/storeflow-api/internal/store"
	"github.com/mcv0000/storeflow-api/internal/user"
)

func main() {
	logger.Init()
	logger.Info.Println("Starting StoreFlow API...")

	cfg, err := config.Load()
	if err != nil {
		logger.Error.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()

	logger.Info.Println("Connecting to PostgreSQL...")
	dbPool, err := db.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()
	logger.Info.Println("Connected to PostgreSQL")

	logger.Info.Println("Connecting to Redis...")
	redisClient, err := cache.NewRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		logger.Error.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	logger.Info.Println("Connected to Redis")

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

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info.Printf("Server listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info.Println("Server stopped")
}
