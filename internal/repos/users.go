package repos

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/gwuah/accounts/internal/models"
)

type usersRepo struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewUsers(logger *slog.Logger, db *sql.DB) *usersRepo {
	return &usersRepo{
		db:     db,
		logger: logger,
	}
}

func (r *usersRepo) GetTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *usersRepo) GetByID(ctx context.Context, tx *sql.Tx, userID int) (*models.User, error) {
	stmt, err := tx.Prepare("select id, email, created_at, update_at from users where id=?")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement. %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to exec query. %w", err)
	}

	var out []*models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			rows.Close()
			return nil, fmt.Errorf("failed to scan response. %w", err)
		}
		out = append(out, &u)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan response. %w", err)
	}

	return out[0], nil
}

func (r *usersRepo) Create(ctx context.Context, tx *sql.Tx, u *models.User) error {
	query := `insert into users (email) values (?);`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Email)
	return err
}
