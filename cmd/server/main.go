package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/benchmark/go-ai-review-benchmark/internal/config"
	"github.com/benchmark/go-ai-review-benchmark/internal/handler"
	"github.com/benchmark/go-ai-review-benchmark/internal/repository"
	"github.com/benchmark/go-ai-review-benchmark/internal/service"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

func main() {
	cfg := config.LoadConfig()

	db, err := sqlx.Connect(cfg.DBDriver, cfg.DBSource)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userRepo := repository.NewUserRepository(db)
	if err := userRepo.InitSchema(ctx); err != nil {
		log.Fatalf("Failed to initialize user schema: %v", err)
	}

	batchRepo := repository.NewBatchRepository(db)
	if err := batchRepo.InitSchema(ctx); err != nil {
		log.Fatalf("Failed to initialize batch schema: %v", err)
	}

	userService := service.NewUserService(userRepo)
	batchProcessor := service.NewBatchProcessor(batchRepo, cfg.WorkerCount)

	userHandler := handler.NewUserHandler(userService)
	batchHandler := handler.NewBatchHandler(batchProcessor, batchRepo)

	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux)
	batchHandler.RegisterRoutes(mux)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %s...", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server listen error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}
