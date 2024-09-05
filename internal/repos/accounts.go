package repos

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gwuah/accounts/internal/models"
)

type accountsRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewAccount(logger *slog.Logger, db *sql.DB) *accountsRepo {
	return &accountsRepo{
		db:     db,
		logger: logger,
	}
}

func (r *accountsRepo) GetTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *accountsRepo) GetByUserID(ctx context.Context, tx *sql.Tx, userID int) ([]*models.Account, error) {
	stmt, err := tx.Prepare("select id, user_id, account_number, created_at, update_at from accounts where user_id=$1;")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement. %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to exec query. %w", err)
	}

	return getMany(rows)
}

func (r *accountsRepo) GetAccounts(ctx context.Context, tx *sql.Tx, accountNumbers []string) ([]*models.Account, error) {
	if len(accountNumbers) == 0 {
		return nil, fmt.Errorf("no account numbers provided")
	}

	placeholders := make([]string, len(accountNumbers))
	args := make([]interface{}, len(accountNumbers))

	for i := range accountNumbers {
		placeholders[i] = fmt.Sprintf("$%d", i+1) // SQL placeholders start at $1
		args[i] = accountNumbers[i]
	}

	query := fmt.Sprintf(
		"SELECT id, user_id, account_number, created_at, updated_at FROM accounts WHERE account_number IN (%s);",
		strings.Join(placeholders, ","),
	)

	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to exec query: %w", err)
	}

	return getMany(rows)
}

func getMany(rows *sql.Rows) ([]*models.Account, error) {
	var out []*models.Account
	for rows.Next() {
		var a models.Account
		err := rows.Scan(&a.ID, &a.UserID, &a.AccountNumber, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			rows.Close()
			return nil, fmt.Errorf("failed to scan response. %w", err)
		}
		out = append(out, &a)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan response. %w", err)
	}
	return out, nil
}

func (r *accountsRepo) Create(ctx context.Context, tx *sql.Tx, a *models.Account) error {
	query := `insert into accounts (user_id, account_number) values ($1, $2) returning id, created_at, updated_at;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement. %w", err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(a.UserID, a.AccountNumber).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to exec query: %w", err)
	}

	return nil
}
