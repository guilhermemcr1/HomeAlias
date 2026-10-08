package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/homealias/homealias/backend/internal/alert"
	"github.com/homealias/homealias/backend/internal/config"
	cf "github.com/homealias/homealias/backend/internal/dns/cloudflare"
	httpapi "github.com/homealias/homealias/backend/internal/http"
	"github.com/homealias/homealias/backend/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	logger := config.NewLogger()
	db, err := store.Open(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := store.BootstrapAdmin(ctx, db, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatalf("bootstrap: %v", err)
	}

	provider := cf.New()
	router := httpapi.NewRouter(cfg, db, provider)

	worker := &alert.Worker{
		DB: db, Log: logger,
		Telegram: alert.NewTelegram(cfg.TelegramBotToken),
		SMTP:     &alert.SMTP{Host: cfg.SMTPHost, Port: cfg.SMTPPort, User: cfg.SMTPUser, Pass: cfg.SMTPPass, From: cfg.SMTPFrom},
		Interval: time.Minute,
	}
	wctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go worker.Start(wctx)
	go func() {
		t := time.NewTicker(24 * time.Hour)
		defer t.Stop()
		for range t.C {
			_ = store.RunRetention(context.Background(), db, cfg.RetentionDays)
		}
	}()

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 16,
	}
	go func() {
		logger.Info("listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("serve: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	cancel()
	shctx, shcancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shcancel()
	_ = srv.Shutdown(shctx)
}
