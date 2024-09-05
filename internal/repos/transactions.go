package repos

import (
	"context"
	"database/sql"
	"log/slog"
)

type transactionsRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewTransactions(logger *slog.Logger, db *sql.DB) *transactionsRepo {
	return &transactionsRepo{
		db:     db,
		logger: logger,
	}
}

func (r *transactionsRepo) GetTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.Begin()
}
