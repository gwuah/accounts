package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gwuah/accounts/internal/config"
	"github.com/gwuah/accounts/internal/database"
	"github.com/gwuah/accounts/internal/models"
)

func requestLogger(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		path, _ := mux.CurrentRoute(r).GetPathTemplate()
		logger.Info("new request",
			"method", r.Method,
			"path", path,
			"timestamp", start,
			"duration", time.Since(start),
		)
	})
}

func main() {
	doneCh := make(chan os.Signal, 1)
	signal.Notify(doneCh, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-doneCh
		cancel()
	}()

	cfg := config.New()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	db, err := database.NewPGConnection(cfg)
	if err != nil {
		logger.Error("failed to setup db connection", "err", err)
		os.Exit(1)
	}

	if cfg.ENV == config.ENV_LOCAL {
		db = db.Debug()
	}

	err = database.RunMigrations(db,
		models.Account{},
		models.Transaction{},
		models.User{},
	)
	if err != nil {
		logger.Error("failed to run migration", "err", err)
		os.Exit(1)
	}

	r := mux.NewRouter()
	r.Use(func(h http.Handler) http.Handler {
		return requestLogger(h, logger)
	})
	r.Path("/").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Handler: r,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", cfg.PORT))
	if err != nil {
		logger.Error("failed to setup tcp listener for service", "err", err)
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf("accounts.svc listening on %s", listener.Addr()))

	go func() {
		server.Serve(listener)
	}()

	<-ctx.Done()

	cancelCtx, cancelFn := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFn()

	if err := server.Shutdown(cancelCtx); err != nil {
		logger.Error("accounts.svc shutdown failed", "err", err)
		os.Exit(1)
	}

	logger.Info("accounts.svc shutdown")
	os.Exit(0)
}