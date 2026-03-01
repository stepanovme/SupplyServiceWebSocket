package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"supplyservicews/internal/app"
	"supplyservicews/internal/config"
	"supplyservicews/internal/db"
	"supplyservicews/internal/ws"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	connections, err := db.NewConnections(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("db connections: %v", err)
	}
	defer connections.Close()

	hub := ws.NewHub()
	wsHandler := ws.NewHandler(hub, cfg.WSReadBufferSize, cfg.WSWriteBufferSize)

	repo := db.NewRepository(connections.Supply)
	watcher := app.NewEventWatcher(repo, hub, cfg.WatcherPoll)
	if err := watcher.Init(ctx); err != nil {
		log.Fatalf("watcher init: %v", err)
	}
	go watcher.Run(ctx)

	mux := http.NewServeMux()
	mux.Handle("/ws", wsHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
	}()

	log.Printf("websocket server started at :%s", cfg.AppPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
}
