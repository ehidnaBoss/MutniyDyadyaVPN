package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"mutinydydayvpn/internals/payments"
	"mutinydydayvpn/internals/storage"
	"mutinydydayvpn/telegram"
	"mutinydydayvpn/internals/user"
)

func main() {
	log.Println("Запуск VPN Bot...")

	// Инициализация базы данных
	db := storage.InitDB()
	defer db.Close()

	// Инициализация репозиториев
	paymentRepo := payments.NewPaymentRepository(db)
	userRepo := user.NewUserRepository(db)

	// Инициализация CloudPayments
	cpProvider := payments.NewCloudPaymentsProvider(
		getEnv("CLOUDPAYMENTS_PUBLIC_ID", "pk_test_123"), // тестовые ключи
		getEnv("CLOUDPAYMENTS_API_SECRET", "test_secret"),
	)

	// Инициализация сервисов
	paymentService := payments.NewService(paymentRepo, cpProvider)
	userService := user.NewService(userRepo)

	// Инициализация хендлеров
	paymentHandler := payments.NewHandler(paymentService)

	// Запуск Telegram бота в отдельной горутине
	go func() {
		botToken := getEnv("TELEGRAM_BOT_TOKEN", "7549844817:AAEL_SfARov4mU2Gyhjdni3igNpwrHjRfEw")
		bot := telegram.NewBot(botToken, paymentService, userService)
		if err := bot.Start(); err != nil {
			log.Fatal("Ошибка запуска Telegram бота:", err)
		}
	}()

	// Настройка HTTP сервера для webhook'ов и API
	router := mux.NewRouter()
	
	// API роуты для платежей
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/payments", paymentHandler.CreatePayment).Methods("POST")
	api.HandleFunc("/payments/{id}", paymentHandler.GetPayment).Methods("GET")
	api.HandleFunc("/webhooks/cloudpayments", paymentHandler.WebhookCloudPayments).Methods("POST")

	// Health check
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// HTTP сервер
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Запуск сервера в отдельной горутине
	go func() {
		log.Println("HTTP сервер запущен на :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Ошибка запуска HTTP сервера:", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Завершение работы...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Ошибка при завершении работы сервера:", err)
	}

	log.Println("Сервер остановлен")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}