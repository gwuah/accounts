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

type AccountRepository interface {
	GetTx(ctx context.Context) (*sql.Tx, error)
	Create(ctx context.Context, tx *sql.Tx, a *models.Account) error
	GetAccounts(ctx context.Context, tx *sql.Tx, accountNumbers []string) ([]*models.Account, error)
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
	return fmt.Sprintf("%09d", n.Int64())
}

func createAccount(global *slog.Logger, accountRepo AccountRepository, userRepo UserRepository) http.HandlerFunc {
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
			return
		}

		err = accountRepo.Create(r.Context(), tx, account)
		if err != nil {
			logger.Error("failed to create account", "err", err)
			writeInternalServer(w, "failed to create account")
			return
		}

		err = tx.Commit()
		if err != nil {
			logger.Error("failed to commit tx", "err", err)
			writeInternalServer(w, "failed to create account")
			return
		}

		writeOk(w, map[string]interface{}{
			"account": account,
		})
	}
}

func getAccount(global *slog.Logger, accountRepo AccountRepository, userRepo UserRepository, transactionRepo TransactionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		logger := global.With("entity", "users")
		accountNumber := mux.Vars(r)["accountNumber"]

		tx, err := userRepo.GetTx(r.Context())
		if err != nil {
			logger.Error("failed to acquire transaction", "err", err)
			writeInternalServer(w, "failed to get account")
			return
		}

		accounts, err := accountRepo.GetAccounts(r.Context(), tx, []string{accountNumber, GenesisAccountNumber})
		if err != nil {
			logger.Error("failed to get accounts", "err", err)
			writeInternalServer(w, "failed to get accounts")
			return
		}

		account := getAccountByAccountNumber(accounts, accountNumber)

		balance, err := transactionRepo.GetBalance(r.Context(), tx, account.ID)
		if err != nil {
			logger.Error("failed to get accounts", "err", err)
			writeInternalServer(w, "failed to get accounts")
			return
		}

		account.Balance = balance

		writeOk(w, map[string]interface{}{
			"account": account,
		})
	}
}

func AddAccountRoutes(logger *slog.Logger, r *mux.Router, accountRepo AccountRepository, userRepo UserRepository, transactionRepo TransactionRepository) {
	r.Methods("POST").Path("/accounts").HandlerFunc(createAccount(logger, accountRepo, userRepo))
	r.Methods("GET").Path("/accounts/{accountNumber}").HandlerFunc(getAccount(logger, accountRepo, userRepo, transactionRepo))

}
