package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/oakproject-flink/oak-flink/oak-lib/certs"
	"github.com/oakproject-flink/oak-flink/oak-server/internal/grpc"
	"github.com/oakproject-flink/oak-flink/oak-server/internal/handlers"
)

func main() {
	// Initialize certificate manager (generates CA and server certs)
	log.Println("Initializing certificate manager...")
	certManager, err := certs.NewManager()
	if err != nil {
		log.Fatalf("Failed to initialize certificate manager: %v", err)
	}

	// Configuration (TODO: load from config file or env vars)
	apiKey := os.Getenv("OAK_API_KEY")
	if apiKey == "" {
		apiKey = "dev-api-key" // TODO: Remove default, require in production
		log.Println("WARNING: Using default API key (dev mode)")
	}

	httpPort := os.Getenv("OAK_HTTP_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	grpcPort := os.Getenv("OAK_GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Serve static files (CSS, JS, images)
	e.Static("/static", "web/static")

	// Web UI routes
	e.GET("/", handlers.Dashboard)
	e.GET("/jobs", handlers.Jobs)
	e.GET("/clusters", handlers.Clusters)
	e.GET("/metrics", handlers.Metrics)

	// API routes for HTMX
	api := e.Group("/api")
	api.GET("/jobs", handlers.APIJobs)
	api.GET("/metrics/stream", handlers.MetricsStream)

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"status":  "ok",
			"service": "oak-server",
			"version": "0.1.0",
		})
	})

	// Create gRPC server with mTLS and agent management
	log.Println("Initializing gRPC server...")
	grpcServer, err := grpc.NewServer(grpc.ServerConfig{
		Port:             grpcPort,
		CertManager:      certManager,
		HeartbeatTimeout: 90 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create gRPC server: %v", err)
	}

	// Start both servers
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Start HTTP server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("🌳 Oak Server (HTTP) starting on http://localhost:%s", httpPort)
		if err := e.Start(":" + httpPort); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Start gRPC server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("🔒 Oak Server (gRPC) starting on grpc://localhost:%s", grpcPort)
		if err := grpcServer.Start(); err != nil {
			errChan <- err
		}
	}()

	log.Println("✅ Oak Server started successfully")
	log.Printf("   Web UI:  http://localhost:%s", httpPort)
	log.Printf("   gRPC:    localhost:%s (mTLS)", grpcPort)
	log.Printf("   Health:  http://localhost:%s/health", httpPort)

	// Wait for interrupt signal or error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	select {
	case <-quit:
		log.Println("Received interrupt signal, shutting down...")
	case err := <-errChan:
		log.Printf("Server error: %v, shutting down...", err)
	}

	// Graceful shutdown
	log.Println("Shutting down servers...")

	// Shutdown HTTP server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Shutdown gRPC server
	grpcServer.Stop()

	// Wait for goroutines to finish
	wg.Wait()

	log.Println("✅ Servers stopped gracefully")
}