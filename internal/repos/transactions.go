package repos

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/gwuah/accounts/internal/models"
)

type TransactionPurpose string

const (
	DEBIT  TransactionPurpose = "debit"
	CREDIT TransactionPurpose = "credit"
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

func (r *transactionsRepo) GetBalance(ctx context.Context, tx *sql.Tx, accountID int) (int64, error) {
	stmt, err := tx.Prepare("select transaction_id, purpose, account_id, amount, created_at from transaction_lines where account_id=$1;")
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement. %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, accountID)
	if err != nil {
		return 0, fmt.Errorf("failed to exec query. %w", err)
	}

	var out []*models.TransactionLine
	for rows.Next() {
		var u models.TransactionLine
		err := rows.Scan(&u.TransactionID, &u.Purpose, &u.AccountID, &u.Amount, &u.CreatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			rows.Close()
			return 0, fmt.Errorf("failed to scan response. %w", err)
		}
		out = append(out, &u)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("failed to scan response. %w", err)
	}

	total := 0
	for _, line := range out {
		if line.Purpose == string(CREDIT) {
			total += line.Amount
		}
		if line.Purpose == string(DEBIT) {
			total -= line.Amount
		}
	}

	return int64(total), nil
}

func (r *transactionsRepo) Create(ctx context.Context, tx *sql.Tx, t *models.Transaction) error {
	query := `insert into transactions (reference) values ($1) returning id, created_at, updated_at;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(t.Reference).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return err
	}

	return err
}

func (r *transactionsRepo) CreateTransactionLine(ctx context.Context, tx *sql.Tx, t *models.TransactionLine) error {
	query := `insert into transaction_lines (transaction_id, purpose, account_id, amount) values ($1, $2, $3, $4) returning id, created_at, updated_at;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	err = stmt.QueryRow(t.TransactionID, t.Purpose, t.AccountID, t.Amount).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return err
	}

	return err
}
