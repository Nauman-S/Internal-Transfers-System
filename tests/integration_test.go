package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/Nauman-S/Internal-Transfers-System/api"
	"github.com/Nauman-S/Internal-Transfers-System/config"
	"github.com/Nauman-S/Internal-Transfers-System/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestServer struct {
	Server *httptest.Server
	DB     *storage.DB
	Config *config.ApplicationConfig
}

func SetupTestServer(t *testing.T) *TestServer {
	testConfig := storage.Config{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvIntOrDefault("TEST_DB_PORT", 5432),
		Database: getEnvOrDefault("TEST_DB_NAME", "transfers_db"),
		Username: getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "password"),
		SSLMode:  getEnvOrDefault("TEST_DB_SSL_MODE", "disable"),
	}

	ctx := context.Background()
	db, err := storage.InitDB(ctx, testConfig)
	require.NoError(t, err, "Failed to initialize test database")

	err = db.Start(ctx)
	require.NoError(t, err, "Failed to start test database")

	time.Sleep(1 * time.Second)

	appConfig := &config.ApplicationConfig{
		Ctx:                ctx,
		DB:                 db,
		AccountRepository:  storage.NewAccountRepository(db),
		TransferRepository: storage.NewTransferRepository(db),
	}

	router := api.InitRouter(appConfig)

	server := httptest.NewServer(router)

	return &TestServer{
		Server: server,
		DB:     db,
		Config: appConfig,
	}
}

func (ts *TestServer) Cleanup() {
	ts.Server.Close()
	ts.DB.Close()
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}


type CreateAccountRequest struct {
	AccountID      int    `json:"account_id"`
	InitialBalance string `json:"initial_balance"`
}

type CreateAccountResponse struct{}

type GetAccountResponse struct {
	AccountID int    `json:"account_id"`
	Balance   string `json:"balance"`
}

type CreateTransactionRequest struct {
	SourceAccountID      int    `json:"source_account_id"`
	DestinationAccountID int    `json:"destination_account_id"`
	Amount              string `json:"amount"`
}

type CreateTransactionResponse struct {
	TransactionID       int    `json:"transaction_id"`
	Status             string `json:"status"`
	SourceBalance      string `json:"source_balance"`
	DestinationBalance string `json:"destination_balance"`
	Amount             string `json:"amount"`
	CreatedAt          string `json:"created_at"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}


func TestCreateAccount(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	baseID := int(time.Now().Unix()) % 100000

	tests := []struct {
		name           string
		request        CreateAccountRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid account creation",
			request: CreateAccountRequest{
				AccountID:      baseID + 1,
				InitialBalance: "100.50",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Account with high precision balance",
			request: CreateAccountRequest{
				AccountID:      baseID + 2,
				InitialBalance: "100.12345678",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Duplicate account creation",
			request: CreateAccountRequest{
				AccountID:      baseID + 1, // Same as first test
				InitialBalance: "200.00",
			},
			expectedStatus: http.StatusConflict,
			expectError:    true,
		},
		{
			name: "Negative balance",
			request: CreateAccountRequest{
				AccountID:      baseID + 3,
				InitialBalance: "-50.00",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			resp, err := http.Post(
				fmt.Sprintf("%s/accounts", ts.Server.URL),
				"application/json",
				bytes.NewBuffer(reqBody),
			)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectError {
				var errorResp ErrorResponse
				err = json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.NotEqual(t, 0, errorResp.Code)
			} else {
				var successResp CreateAccountResponse
				err = json.NewDecoder(resp.Body).Decode(&successResp)
				require.NoError(t, err)
			}
		})
	}
}

func TestGetAccount(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	baseID := int(time.Now().Unix()) % 100000

	createReq := CreateAccountRequest{
		AccountID:      baseID + 100,
		InitialBalance: "100.12345678",
	}
	reqBody, err := json.Marshal(createReq)
	require.NoError(t, err)

	createResp, err := http.Post(
		fmt.Sprintf("%s/accounts", ts.Server.URL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	require.NoError(t, err)
	createResp.Body.Close()
	assert.Equal(t, http.StatusOK, createResp.StatusCode)

	tests := []struct {
		name           string
		accountID      int
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid account query",
			accountID:      baseID + 100,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Non-existent account",
			accountID:      baseID + 999,
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "Invalid account ID",
			accountID:      -1,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("%s/accounts/%d", ts.Server.URL, tt.accountID))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectError {
				var errorResp ErrorResponse
				err = json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.NotEqual(t, 0, errorResp.Code)
			} else {
				var accountResp GetAccountResponse
				err = json.NewDecoder(resp.Body).Decode(&accountResp)
				require.NoError(t, err)
				assert.Equal(t, tt.accountID, accountResp.AccountID)
				assert.Equal(t, "100.12345678", accountResp.Balance)
			}
		})
	}
}

// Test Transaction Processing
func TestCreateTransaction(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	baseID := int(time.Now().Unix()) % 100000

	accounts := []CreateAccountRequest{
		{AccountID: baseID + 200, InitialBalance: "1000.00"},
		{AccountID: baseID + 201, InitialBalance: "500.00"},
	}

	for _, account := range accounts {
		reqBody, err := json.Marshal(account)
		require.NoError(t, err)

		resp, err := http.Post(
			fmt.Sprintf("%s/accounts", ts.Server.URL),
			"application/json",
			bytes.NewBuffer(reqBody),
		)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	tests := []struct {
		name           string
		request        CreateTransactionRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid transaction",
			request: CreateTransactionRequest{
				SourceAccountID:      baseID + 200,
				DestinationAccountID: baseID + 201,
				Amount:              "100.50",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "High precision transaction",
			request: CreateTransactionRequest{
				SourceAccountID:      baseID + 200,
				DestinationAccountID: baseID + 201,
				Amount:              "0.12345678",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Insufficient funds",
			request: CreateTransactionRequest{
				SourceAccountID:      baseID + 200,
				DestinationAccountID: baseID + 201,
				Amount:              "2000.00",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "Same account transfer",
			request: CreateTransactionRequest{
				SourceAccountID:      baseID + 200,
				DestinationAccountID: baseID + 200,
				Amount:              "100.00",
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "Non-existent source account",
			request: CreateTransactionRequest{
				SourceAccountID:      baseID + 999,
				DestinationAccountID: baseID + 201,
				Amount:              "100.00",
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name: "Non-existent destination account",
			request: CreateTransactionRequest{
				SourceAccountID:      baseID + 200,
				DestinationAccountID: baseID + 999,
				Amount:              "100.00",
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			resp, err := http.Post(
				fmt.Sprintf("%s/transactions", ts.Server.URL),
				"application/json",
				bytes.NewBuffer(reqBody),
			)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectError {
				var errorResp ErrorResponse
				err = json.NewDecoder(resp.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.NotEqual(t, 0, errorResp.Code)
			} else {
				var transactionResp CreateTransactionResponse
				err = json.NewDecoder(resp.Body).Decode(&transactionResp)
				require.NoError(t, err)
				assert.Equal(t, "COMPLETED", transactionResp.Status)
				assert.Equal(t, tt.request.Amount, transactionResp.Amount)
				assert.NotEmpty(t, transactionResp.TransactionID)
			}
		})
	}
}

func TestCompleteWorkflow(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	baseID := int(time.Now().Unix()) % 100000

	accounts := []CreateAccountRequest{
		{AccountID: baseID + 1001, InitialBalance: "1000.12345678"},
		{AccountID: baseID + 1002, InitialBalance: "500.98765432"},
	}

	for _, account := range accounts {
		reqBody, err := json.Marshal(account)
		require.NoError(t, err)

		resp, err := http.Post(
			fmt.Sprintf("%s/accounts", ts.Server.URL),
			"application/json",
			bytes.NewBuffer(reqBody),
		)
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	for _, accountID := range []int{baseID + 1001, baseID + 1002} {
		resp, err := http.Get(fmt.Sprintf("%s/accounts/%d", ts.Server.URL, accountID))
		require.NoError(t, err)
		defer resp.Body.Close()

		var accountResp GetAccountResponse
		err = json.NewDecoder(resp.Body).Decode(&accountResp)
		require.NoError(t, err)
		assert.Equal(t, accountID, accountResp.AccountID)
	}

	transactions := []CreateTransactionRequest{
		{SourceAccountID: baseID + 1001, DestinationAccountID: baseID + 1002, Amount: "100.12345678"},
		{SourceAccountID: baseID + 1002, DestinationAccountID: baseID + 1001, Amount: "50.98765432"},
		{SourceAccountID: baseID + 1001, DestinationAccountID: baseID + 1002, Amount: "25.55555555"},
	}

	for i, transaction := range transactions {
		reqBody, err := json.Marshal(transaction)
		require.NoError(t, err)

		resp, err := http.Post(
			fmt.Sprintf("%s/transactions", ts.Server.URL),
			"application/json",
			bytes.NewBuffer(reqBody),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Transaction %d failed", i+1)

		var transactionResp CreateTransactionResponse
		err = json.NewDecoder(resp.Body).Decode(&transactionResp)
		require.NoError(t, err)
		assert.Equal(t, "COMPLETED", transactionResp.Status)
	}

	
	time.Sleep(100 * time.Millisecond)
	
	expectedBalances := map[int]string{
		baseID + 1001: "925.43209877",
		baseID + 1002: "575.67901233",
	}

	for _, accountID := range []int{baseID + 1001, baseID + 1002} {
		resp, err := http.Get(fmt.Sprintf("%s/accounts/%d", ts.Server.URL, accountID))
		require.NoError(t, err)
		defer resp.Body.Close()

		var accountResp GetAccountResponse
		err = json.NewDecoder(resp.Body).Decode(&accountResp)
		require.NoError(t, err)
		assert.Equal(t, accountID, accountResp.AccountID)
		assert.Equal(t, expectedBalances[accountID], accountResp.Balance, 
			"Account %d balance mismatch. Expected: %s, Got: %s", 
			accountID, expectedBalances[accountID], accountResp.Balance)
	}
}
