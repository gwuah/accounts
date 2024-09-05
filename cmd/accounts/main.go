package main

import (
	"context"
	"database/sql"
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
	"github.com/gwuah/accounts/internal/repos"
	"github.com/gwuah/accounts/internal/services"
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

func runSeeds(db *sql.DB) error {
	// create 1 user for the bank
	// create 1 account for the banks user
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("insert into users (email) values ($1) on conflict do nothing;", "primary@accounts.com")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("insert into accounts (user_id, account_number) values ($1,$2) on conflict do nothing;", 1, "000000000")
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
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

	db, err := database.New(ctx, cfg, database.POSTGRES)
	if err != nil {
		logger.Error("failed to setup db connection", "err", err)
		os.Exit(1)
	}

	err = runSeeds(db.Instance())
	if err != nil {
		logger.Error("failed to run seeds", "err", err)
		os.Exit(1)
	}
	ar := repos.NewAccount(logger, db.Instance())
	ur := repos.NewUsers(logger, db.Instance())
	// tr := repos.NewTransactions(logger, db.Instance())

	r := mux.NewRouter()
	r.Use(func(h http.Handler) http.Handler {
		return requestLogger(h, logger)
	})
	r.Path("/").Methods(http.MethodGet).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	services.AddUserRoutes(logger, r, ar, ur)
	services.AddAccountRoutes(logger, r, ar, ur)
	// services.AddTransactionRoutes(logger, r, ar, ur, tr)

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
