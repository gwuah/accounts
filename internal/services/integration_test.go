package services_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	"github.com/gwuah/accounts/internal/config"
	"github.com/gwuah/accounts/internal/database"
	"github.com/gwuah/accounts/internal/models"
	"github.com/gwuah/accounts/internal/repos"
	"github.com/gwuah/accounts/internal/services"
	"github.com/gwuah/accounts/pkg"
	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) (context.Context, *mux.Router, *database.DB, *slog.Logger, func()) {
	ctx := context.Background()

	dir, err := ioutil.TempDir("", "accounts-sqlite")
	require.NoError(t, err)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg := config.New()
	cfg.DB_URL = filepath.Join(dir, "accounts.db")

	db, err := database.New(ctx, cfg, database.SQLITE)
	require.NoError(t, err)

	err = database.RunSeeds(db.Instance())
	require.NoError(t, err)

	ar := repos.NewAccount(logger, db.Instance())
	ur := repos.NewUsers(logger, db.Instance())
	tr := repos.NewTransactions(logger, db.Instance())

	r := mux.NewRouter()
	services.AddUserRoutes(logger, r, ar, ur)
	services.AddAccountRoutes(logger, r, ar, ur, tr)
	services.AddTransactionRoutes(logger, r, ar, ur, tr)

	teardown := func() {
		os.Remove(filepath.Join(dir, "accounts.db"))
	}

	return ctx, r, db, logger, teardown
}

func performRequestAndGetResponse[T any](r *mux.Router, t *testing.T) func(req *http.Request, input *T) *httptest.ResponseRecorder {
	return func(req *http.Request, input *T) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		bd, err := io.ReadAll(w.Body)
		require.NoError(t, err)

		// t.Errorf("response %v", string(bd))

		err = json.Unmarshal(bd, input)
		require.NoError(t, err)

		return w

	}
}

func TestCreateUserCreateAccountDepositTransfer(t *testing.T) {
	_, r, _, _, teardown := setup(t)
	defer teardown()

	// create user
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer([]byte(`{"email": "1@gmail.com"}'`)))

	var uResponse map[string]models.User
	w := performRequestAndGetResponse[map[string]models.User](r, t)(req, &uResponse)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "1@gmail.com", uResponse["user"].Email)

	// create 2 accounts for user
	reqBody := fmt.Sprintf(`{"user_id": %d}`, uResponse["user"].ID)

	req = httptest.NewRequest("POST", "/accounts", bytes.NewBuffer([]byte(reqBody)))
	var a1Response map[string]models.Account
	w = performRequestAndGetResponse[map[string]models.Account](r, t)(req, &a1Response)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, uResponse["user"].ID, a1Response["account"].UserID)

	req = httptest.NewRequest("POST", "/accounts", bytes.NewBuffer([]byte(reqBody)))
	var a2Response map[string]models.Account
	w = performRequestAndGetResponse[map[string]models.Account](r, t)(req, &a2Response)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, uResponse["user"].ID, a2Response["account"].UserID)

	// deposit 200 usd into account 1
	reqBody = fmt.Sprintf(`{"to":"%s","type":"deposit","amount":200,"reference":"%s"}`, a1Response["account"].AccountNumber, pkg.CreateAccountNumber())
	req = httptest.NewRequest("POST", "/transactions", bytes.NewBuffer([]byte(reqBody)))
	var dResponse map[string]string
	w = performRequestAndGetResponse[map[string]string](r, t)(req, &dResponse)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "ok", dResponse["status"])

	// transfer 200 to account 2, in 2 transactions of 100
	reqBody = fmt.Sprintf(`{"from":"%s","to":"%s","type":"transfer","amount":100,"reference":"%s"}`, a1Response["account"].AccountNumber, a2Response["account"].AccountNumber, pkg.CreateAccountNumber())
	req = httptest.NewRequest("POST", "/transactions", bytes.NewBuffer([]byte(reqBody)))
	var t1Response map[string]string
	w = performRequestAndGetResponse[map[string]string](r, t)(req, &t1Response)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "ok", t1Response["status"])

	reqBody = fmt.Sprintf(`{"from":"%s","to":"%s","type":"transfer","amount":100,"reference":"%s"}`, a1Response["account"].AccountNumber, a2Response["account"].AccountNumber, pkg.CreateAccountNumber())
	req = httptest.NewRequest("POST", "/transactions", bytes.NewBuffer([]byte(reqBody)))
	var t2Response map[string]string
	w = performRequestAndGetResponse[map[string]string](r, t)(req, &t2Response)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "ok", t2Response["status"])

	// verify that performing another transaction from account 1 -> 2 fails with insufficient balance error
	reqBody = fmt.Sprintf(`{"from":"%s","to":"%s","type":"transfer","amount":100,"reference":"%s"}`, a1Response["account"].AccountNumber, a2Response["account"].AccountNumber, pkg.CreateAccountNumber())
	req = httptest.NewRequest("POST", "/transactions", bytes.NewBuffer([]byte(reqBody)))
	var t3Response map[string]string
	w = performRequestAndGetResponse[map[string]string](r, t)(req, &t3Response)
	require.Equal(t, http.StatusUnprocessableEntity, w.Code)
	require.Equal(t, "insufficient balance", t3Response["error"])

	// verify that account 1 has balance of 0
	reqBody = fmt.Sprintf(`{"from":"%s","to":"%s","type":"transfer","amount":100,"reference":"%s"}`, a1Response["account"].AccountNumber, a2Response["account"].AccountNumber, pkg.CreateAccountNumber())
	req = httptest.NewRequest("GET", fmt.Sprintf("/accounts/%s", a1Response["account"].AccountNumber), bytes.NewBuffer([]byte(reqBody)))
	var finalA1Response map[string]models.Account
	w = performRequestAndGetResponse[map[string]models.Account](r, t)(req, &finalA1Response)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, float64(0), finalA1Response["account"].Balance)

	// verify that account 2 has balance of 200
	reqBody = fmt.Sprintf(`{"from":"%s","to":"%s","type":"transfer","amount":100,"reference":"%s"}`, a1Response["account"].AccountNumber, a2Response["account"].AccountNumber, pkg.CreateAccountNumber())
	req = httptest.NewRequest("GET", fmt.Sprintf("/accounts/%s", a2Response["account"].AccountNumber), bytes.NewBuffer([]byte(reqBody)))
	var finalA2Response map[string]models.Account
	w = performRequestAndGetResponse[map[string]models.Account](r, t)(req, &finalA2Response)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, float64(200), finalA2Response["account"].Balance)
}
