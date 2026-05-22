package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	httpdelivery "payment-platform-server/internal/delivery/http"
	"payment-platform-server/internal/repository"
	"payment-platform-server/internal/usecase"
)

const (
	defaultAddress  = ":8080"
	shutdownTimeout = 10 * time.Second
)

type databaseConfig struct {
	host     string
	port     string
	user     string
	password string
	name     string
	sslMode  string
}

func main() {
	logger := log.New(os.Stdout, "payment-platform-server: ", log.LstdFlags|log.Lmicroseconds)

	dbConfig := loadDatabaseConfig()

	db, err := sql.Open("postgres", dbConfig.connectionString())
	if err != nil {
		logger.Fatalf("failed to open database connection: %v", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			logger.Printf("failed to close database connection: %v", closeErr)
		}
	}()

	pingContext, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	if err := db.PingContext(pingContext); err != nil {
		cancelPing()
		logger.Fatalf("failed to connect to database: %v", err)
	}
	cancelPing()

	repo := repository.NewPostgresMessageRepository(db, "Hello!")
	initContext, cancelInit := context.WithTimeout(context.Background(), 5*time.Second)
	if err := repo.InitSchema(initContext); err != nil {
		cancelInit()
		logger.Fatalf("failed to initialize database schema: %v", err)
	}
	cancelInit()

	uc := usecase.NewTestUseCase(repo, getEnv("JWT_SECRET", "local-development-secret"))
	handler := httpdelivery.NewTestHandler(uc)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	serverAddress := getEnv("SERVER_ADDRESS", defaultAddress)
	server := &http.Server{
		Addr:              serverAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		logger.Printf("server started on %s", serverAddress)
		logger.Println("GET  http://localhost:8080/test")
		logger.Println("POST http://localhost:8080/dbtest")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
			return
		}
		serverErrors <- nil
	}()

	select {
	case err := <-serverErrors:
		if err != nil {
			logger.Fatalf("server failed: %v", err)
		}
		return
	case <-shutdownSignal.Done():
		stop()
		logger.Println("shutdown signal received")
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Printf("graceful shutdown failed: %v", err)
		if closeErr := server.Close(); closeErr != nil {
			logger.Printf("forced close failed: %v", closeErr)
		}
		return
	}

	if err := <-serverErrors; err != nil {
		logger.Printf("server stopped with error: %v", err)
		return
	}

	logger.Println("server stopped gracefully")
}

func loadDatabaseConfig() databaseConfig {
	return databaseConfig{
		host:     getEnv("DB_HOST", "localhost"),
		port:     getEnv("DB_PORT", "5432"),
		user:     getEnv("DB_USER", "postgres"),
		password: getEnv("DB_PASSWORD", "postgres"),
		name:     getEnv("DB_NAME", "payment_platform"),
		sslMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

func (c databaseConfig) connectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.host,
		c.port,
		c.user,
		c.password,
		c.name,
		c.sslMode,
	)
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
