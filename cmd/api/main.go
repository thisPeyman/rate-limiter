package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thisPeyman/rate-limiter/internal/limiter"
	"github.com/thisPeyman/rate-limiter/internal/platform/redis"
	"github.com/thisPeyman/rate-limiter/internal/server"
)

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb, err := redis.NewClient(redisAddr)
	if err != nil {
		log.Fatalf("Fatal: Could not connect to Redis: %v", err)
	}
	defer rdb.Close()

	rateLimiter := limiter.NewSlidingWindowLimiter(rdb, 1*time.Minute)
	rlMiddleware := server.NewRateLimitMiddleware(rateLimiter)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Request allowed: Resource accessed"))
	})

	mux.Handle("/api/resource", rlMiddleware.Handler(apiHandler))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		log.Printf("Server starting on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
