package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gwuah/accounts/internal/models"
	"github.com/gwuah/accounts/internal/repos"
)

const (
	Deposit              string = "deposit"
	Transfer             string = "transfer"
	GenesisAccountNumber string = "000000000"
)

type TransactionRepository interface {
	GetTx(ctx context.Context) (*sql.Tx, error)
	Create(ctx context.Context, tx *sql.Tx, t *models.Transaction) error
	CreateTransactionLine(ctx context.Context, tx *sql.Tx, t *models.TransactionLine) error
	GetBalance(ctx context.Context, tx *sql.Tx, accountID int) (int64, error)
}

type createTransactionRequest struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Type      string `json:"type"`
	Amount    int    `json:"amount"`
	Reference string `json:"reference"`
}

func (r createTransactionRequest) validate() error {
	switch r.Type {
	case Deposit:
		if r.To == "" {
			return errors.New("destination account is required for 'deposit'")
		}
		if r.To == GenesisAccountNumber {
			return errors.New("action not allowed for this account number")
		}
	case Transfer:
		if r.From == "" || r.To == "" {
			return errors.New("origin/destination accounts are required for 'transfer'")
		}
	default:
		return errors.New("transaction 'type' is required")
	}
	if r.Amount == 0 {
		return errors.New("amount is required. (non-zero value)")
	}

	return nil
}

func createTransaction(global *slog.Logger, accountRepo AccountRepository, userRepo UserRepository, transactionRepo TransactionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := global.With("entity", "transactions")

		var req createTransactionRequest
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

		// check if user exists, before creating account (very important validation)

		transaction := &models.Transaction{
			Reference: req.Reference,
		}

		tx, err := transactionRepo.GetTx(r.Context())
		if err != nil {
			logger.Error("failed to acquire transaction", "err", err)
			writeInternalServer(w, "failed to create transaction")
			return
		}

		// a deposit is like any transfer, except we debit the genesis account, with id 0
		if req.Type == Deposit {
			req.From = GenesisAccountNumber
		}

		accounts, err := accountRepo.GetAccounts(r.Context(), tx, []string{
			req.From,
			req.To,
		})
		if err != nil {
			logger.Error("failed to get accounts", "err", err)
			writeInternalServer(w, "failed to create transaction")
			return
		}

		if len(accounts) != 2 {
			logger.Error("uneven number of accounts", "err", err, "count", len(accounts))
			writeInternalServer(w, "failed to create transaction")
			return
		}

		err = transactionRepo.Create(r.Context(), tx, transaction)
		if err != nil {
			tx.Rollback()
			if strings.Contains(err.Error(), "duplicate key value") && strings.Contains(err.Error(), "transactions_reference_key") {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "duplicate transaction request",
				})
				return
			}
			logger.Error("failed to create payment transaction", "err", err)
			writeInternalServer(w, "failed to create transaction")
			return
		}

		// before performing this debit/credit, we need to verify if the origin account has enough balance for this transaction.
		// we however exclude the genesis account, which has an account number of 0, since it's a special account that only hold risks.
		if req.From != GenesisAccountNumber {
			account := getAccountByAccountNumber(accounts, req.From)
			balance, err := transactionRepo.GetBalance(r.Context(), tx, account.ID)
			if err != nil {
				tx.Rollback()
				logger.Error("failed to get balance", "err", err)
				writeInternalServer(w, "failed to create transaction")
				return
			}

			if balance < int64(req.Amount) {
				tx.Rollback()
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusUnprocessableEntity)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error": "insufficient balance",
				})
				return
			}

		}

		debit := models.TransactionLine{
			TransactionID: transaction.ID,
			AccountID:     getAccountByAccountNumber(accounts, req.From).ID,
			Amount:        req.Amount * 100,
			Purpose:       string(repos.DEBIT),
		}

		credit := models.TransactionLine{
			TransactionID: transaction.ID,
			AccountID:     getAccountByAccountNumber(accounts, req.To).ID,
			Amount:        req.Amount * 100,
			Purpose:       string(repos.CREDIT),
		}

		err = transactionRepo.CreateTransactionLine(r.Context(), tx, &debit)
		if err != nil {
			tx.Rollback()
			logger.Error("failed to create debit transaction", "err", err)
			writeInternalServer(w, "failed to create transaction")
			return
		}

		err = transactionRepo.CreateTransactionLine(r.Context(), tx, &credit)
		if err != nil {
			tx.Rollback()
			logger.Error("failed to create credit transaction", "err", err)
			writeInternalServer(w, "failed to create transaction")
			return
		}

		err = tx.Commit()
		if err != nil {
			logger.Error("failed to commit tx", "err", err)
			writeInternalServer(w, "failed to create transaction")
			return
		}

		writeOk(w, map[string]interface{}{
			"status": "ok",
		})
	}
}

func getAccountByAccountNumber(accounts []*models.Account, accountNumber string) *models.Account {
	for _, acc := range accounts {
		if acc.AccountNumber == accountNumber {
			return acc
		}
	}
	return nil
}

func AddTransactionRoutes(logger *slog.Logger, r *mux.Router, accountRepo AccountRepository, userRepo UserRepository, transactionRepo TransactionRepository) {
	r.Methods("POST").Path("/transactions").HandlerFunc(createTransaction(logger, accountRepo, userRepo, transactionRepo))
}
