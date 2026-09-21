package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"TestTask/internal/config"
	"TestTask/internal/db"
	"TestTask/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPostgresPool(ctx, config.Cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Не удалось соединиться с базой: %v", err)
	}
	defer pool.Close()
	log.Println("Соединение с базой данных установлено")

	redisClient, err := db.NewRedisClient(ctx, config.Cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Не удалось соединиться с Redis: %v", err)
	}
	defer redisClient.Close()

	engine, err := server.New(config.Cfg.JWTSecret, pool, redisClient, config.Cfg.UploadDir)
	if err != nil {
		log.Fatalf("Не удалось поднять сервер: %v", err)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: engine,
	}

	go func() {
		log.Println("Сервер запущен на порту :8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Получен сигнал остановки, выполняется остановка")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Не удалось корректно завершить работу: %v", err)
	} else {
		log.Println("Сервер завершил работу корректно")
	}
}
