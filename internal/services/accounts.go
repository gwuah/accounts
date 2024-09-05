package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gwuah/accounts/internal/models"
)

type accountRepository interface {
	GetTx(ctx context.Context) (*sql.Tx, error)
	Create(ctx context.Context, tx *sql.Tx, a *models.Account) error
}

type createAccountRequest struct {
	UserID int `json:"user_id"`
}

func (r createAccountRequest) validate() error {
	if r.UserID == 0 { // $1
		return errors.New("'user_id' is required, can't be empty")
	}
	return nil
}

func createAccountNumber() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1e9))
	return fmt.Sprintf("%d", n.Int64())
}

func createAccount(global *slog.Logger, accountRepo accountRepository, userRepo userRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := global.With("entity", "accounts")

		var req createAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Error("error reading request", "err", err)
			writeBadRequest(w, err)
			return
		}
		if err := req.validate(); err != nil {
			logger.Error("error validating request", "err", err)
			writeBadRequest(w, err)
			return
		}

		// check if user exists, before creating account.
		// very important validation

		account := &models.Account{
			UserID:        req.UserID,
			AccountNumber: createAccountNumber(),
		}

		tx, err := userRepo.GetTx(r.Context())
		if err != nil {
			logger.Error("failed to acquire transaction", "err", err)
			writeInternalServer(w, "failed to create account")
		}

		err = accountRepo.Create(r.Context(), tx, account)
		if err != nil {
			logger.Error("failed to create user", "err", err)
			writeInternalServer(w, "failed to create user")
		}

		err = tx.Commit()
		if err != nil {
			logger.Error("failed to commit tx", "err", err)
			writeInternalServer(w, "failed to create account")
		}

		writeOk(w, map[string]interface{}{
			"account": account,
		})
	}
}

func AddAccountRoutes(logger *slog.Logger, r *mux.Router, accountRepo accountRepository, userRepo userRepository) {
	r.Methods("POST").Path("/accounts").HandlerFunc(createAccount(logger, accountRepo, userRepo))
}
