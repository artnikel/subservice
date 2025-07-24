// Package main is an entry point to application
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/artnikel/subservice/docs"
	"github.com/artnikel/subservice/internal/config"
	"github.com/artnikel/subservice/internal/handlers"
	"github.com/artnikel/subservice/internal/logger"
	"github.com/artnikel/subservice/internal/repository"
	"github.com/artnikel/subservice/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	dbConnectTimeout      = 5 * time.Second
	serverShutdownTimeout = 30 * time.Second
)

// @title Subscriptions API
// @version 1.0
// @description REST API for managing user subscriptions
// @host localhost:8080
// @BasePath /api/v1
func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	log := logger.NewLogger(cfg.Logging.Level, cfg.Logging.File)

	db, err := connectPGX(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Info("Connected to database successfully")

	repo := repository.NewSubscriptionRepository(db)
	svc := service.NewSubscriptionService(repo, log)
	handler := handlers.NewSubscriptionHandler(svc, log)

	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger.GinLogger(log))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := router.Group("/api/v1")
	{
		subscriptions := api.Group("/subscriptions")
		subscriptions.POST("", handler.CreateSubscription)
		subscriptions.GET("", handler.ListSubscriptions)
		subscriptions.GET("/:id", handler.GetSubscription)
		subscriptions.PUT("/:id", handler.UpdateSubscription)
		subscriptions.DELETE("/:id", handler.DeleteSubscription)

		api.GET("/cost/summary", handler.GetCostSummary)
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		log.Infof("Starting server on port %d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		//nolint:gocritic
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}

func connectPGX(cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	config.MaxConns = cfg.MaxConns

	ctx, cancel := context.WithTimeout(context.Background(), dbConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return pool, nil
}
