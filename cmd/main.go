package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DeadlyParkour777/pr-service/internal/config"
	"github.com/DeadlyParkour777/pr-service/internal/handler"
	"github.com/DeadlyParkour777/pr-service/internal/service"
	"github.com/DeadlyParkour777/pr-service/internal/store"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.NewConfig()
	if err != nil {
		return err
	}

	store, err := store.NewStore(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer store.Close()

	deps := service.Dependencies{
		TeamRepo:  store.Team(),
		UserRepo:  store.User(),
		PRRepo:    store.PR(),
		StatsRepo: store.PR(),
	}

	service := service.NewService(deps)
	handler := handler.NewHandler(service, cfg.JWTSecret, cfg.OpenAPISpecPath)
	router := handler.InitRoutes()

	server := &http.Server{
		Addr:    ":" + cfg.HTTP_PORT,
		Handler: router,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Printf("Server started at port %s", cfg.HTTP_PORT)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v, ", err)
		}
	}()

	<-stop

	log.Println("Shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return err
	}

	log.Println("Server stopped")
	return nil
}
