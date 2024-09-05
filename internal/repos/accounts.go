package repos

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

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
	query := `insert into accounts (user_id, account_number) values ($1, $2);`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(a.UserID, a.AccountNumber)
	return err
}
