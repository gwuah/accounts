package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gwuah/accounts/internal/models"
)

type userRepository interface {
	GetTx(ctx context.Context) (*sql.Tx, error)
	Create(ctx context.Context, tx *sql.Tx, u *models.User) error
	GetByID(ctx context.Context, tx *sql.Tx, userID int) (*models.User, error)
}

type createUserRequest struct {
	Email string `json:"email"`
}

func (r createUserRequest) validate() error {
	if strings.TrimSpace(r.Email) == "" { // $1
		return errors.New("'email' is required, can't be empty")
	}
	return nil
}

func stringToInt(s string) int {
	intValue, _ := strconv.Atoi(s)
	// check for err
	return intValue
}

func writeBadRequest(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func writeInternalServer(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": msg,
	})
}

func writeOk(w http.ResponseWriter, data map[string]interface{}) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(data)
}

func createUser(global *slog.Logger, userRepo userRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := global.With("entity", "users")

		var req createUserRequest
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

		tx, err := userRepo.GetTx(r.Context())
		if err != nil {
			logger.Error("failed to acquire transaction", "err", err)
			writeInternalServer(w, "failed to create user")
		}

		err = userRepo.Create(r.Context(), tx, &models.User{
			Email: req.Email,
		})
		if err != nil {
			logger.Error("failed to create user", "err", err)
			writeInternalServer(w, "failed to create user")
		}

		err = tx.Commit()
		if err != nil {
			logger.Error("failed to commit tx", "err", err)
			writeInternalServer(w, "failed to create user")
		}

		writeOk(w, map[string]interface{}{
			"user": &models.User{
				Email: req.Email,
			},
		})
	}
}

func findUser(global *slog.Logger, userRepo userRepository, accountRepo accountRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := global.With("entity", "users")
		id := mux.Vars(r)["id"]

		tx, err := userRepo.GetTx(r.Context())
		if err != nil {
			logger.Error("failed to acquire transaction", "err", err)
			writeInternalServer(w, "failed to get user")
		}

		user, err := userRepo.GetByID(r.Context(), tx, stringToInt(id))
		if err != nil {
			logger.Error("failed to get user", "err", err)
			writeInternalServer(w, "failed to get user")
		}

		err = tx.Commit()
		if err != nil {
			logger.Error("failed to commit tx", "err", err)
			writeInternalServer(w, "failed to get user")
		}

		writeOk(w, map[string]interface{}{
			"user": user,
		})
	}
}

func AddUserRoutes(logger *slog.Logger, r *mux.Router, accountRepo accountRepository, userRepo userRepository) {
	r.Methods("GET").Path("/users/{id}").HandlerFunc(findUser(logger, userRepo, accountRepo))
	r.Methods("POST").Path("/users").HandlerFunc(createUser(logger, userRepo))
}
