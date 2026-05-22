package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpdelivery "payment-platform-server/internal/delivery/http"
	"payment-platform-server/internal/repository"
	"payment-platform-server/internal/usecase"
)

func main() {
	repo := repository.NewStaticMessageRepository("Hello!")
	uc := usecase.NewTestUseCase(repo)
	handler := httpdelivery.NewTestHandler(uc)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("Сервер запущен на http://localhost:8080/test")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Ошибка: %v\n", err)
		}
	}()

	<-stop

	fmt.Println("\nВыключаю сервер...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Ошибка при выключении: %v\n", err)
	}

	fmt.Println("Сервер успешно остановлен.")
}
