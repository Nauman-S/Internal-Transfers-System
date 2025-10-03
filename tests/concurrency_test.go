package tests

import (
	"bytes"
	"strconv"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConcurrentAccountCreation(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	baseID := int(time.Now().Unix()) % 100000
	accountID := baseID + 10000
	numGoroutines := 1000

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex


	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

		req := CreateAccountRequest{
			AccountID:      accountID,
			InitialBalance: "100.00",
		}

			reqBody, err := json.Marshal(req)
			require.NoError(t, err)

			resp, err := http.Post(
				fmt.Sprintf("%s/accounts", ts.Server.URL),
				"application/json",
				bytes.NewBuffer(reqBody),
			)
			require.NoError(t, err)
			defer resp.Body.Close()

			mu.Lock()
			if resp.StatusCode == http.StatusOK {
				successCount++
			}
			mu.Unlock()

			assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusConflict,
				"Expected 200 or 409, got %d", resp.StatusCode)
		}()
	}

	wg.Wait()

	assert.Equal(t, 1, successCount, "Expected exactly 1 successful account creation, got %d", successCount)

	resp, err := http.Get(fmt.Sprintf("%s/accounts/%d", ts.Server.URL, accountID))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var accountResp GetAccountResponse
	err = json.NewDecoder(resp.Body).Decode(&accountResp)
	require.NoError(t, err)
	assert.Equal(t, accountID, accountResp.AccountID)
	assert.Equal(t, "100", accountResp.Balance)
}


func TestConcurrentTransfers(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	baseID := int(time.Now().UnixNano()) % 100000

	accounts := []CreateAccountRequest{
		{AccountID: baseID + 20000, InitialBalance: "10000000.00"},
		{AccountID: baseID + 20001, InitialBalance: "10000000.00"},
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
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Logf("Account creation failed with status %d for account %d", resp.StatusCode, account.AccountID)
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	numTransfers := 1000
	transferAmount := "10.00"
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	for i := 0; i < numTransfers; i++ {
		wg.Add(1)
		go func(transferID int) {
			defer wg.Done()

			// Alternate between account1->account2 and account2->account1
			var sourceID, destID int
			if transferID%2 == 0 {
				sourceID = baseID + 20000
				destID = baseID + 20001
			} else {
				sourceID = baseID + 20001
				destID = baseID + 20000
			}

			req := CreateTransactionRequest{
				SourceAccountID:      sourceID,
				DestinationAccountID: destID,
				Amount:              transferAmount,
			}

			reqBody, err := json.Marshal(req)
			require.NoError(t, err)

			resp, err := http.Post(
				fmt.Sprintf("%s/transactions", ts.Server.URL),
				"application/json",
				bytes.NewBuffer(reqBody),
			)
			require.NoError(t, err)
			defer resp.Body.Close()

			mu.Lock()
			if resp.StatusCode == http.StatusOK {
				successCount++
			}
			mu.Unlock()

			assert.True(t, resp.StatusCode == http.StatusOK,
				"Expected 200, got %d", resp.StatusCode)
		}(i)
	}

	wg.Wait()

	time.Sleep(100 * time.Millisecond)

	for _, accountID := range []int{baseID + 20000, baseID + 20001} {
		resp, err := http.Get(fmt.Sprintf("%s/accounts/%d", ts.Server.URL, accountID))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var accountResp GetAccountResponse
		err = json.NewDecoder(resp.Body).Decode(&accountResp)
		require.NoError(t, err)
		assert.Equal(t, accountID, accountResp.AccountID)
		balance, err := strconv.ParseFloat(accountResp.Balance, 64)
		require.NoError(t, err)
		
		expectedBalance := 10000000.0
		assert.Equal(t, expectedBalance, balance, "Account %d should have balance %f, got %f", accountID, expectedBalance, balance)
	}
}

func TestConcurrentReadsAndWrites(t *testing.T) {
	ts := SetupTestServer(t)
	defer ts.Cleanup()

	baseID := int(time.Now().Unix()) % 100000

	account := CreateAccountRequest{
		AccountID:      baseID + 40000,
		InitialBalance: "500.00",
	}

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

	numOperations := 1000
	var wg sync.WaitGroup

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			resp, err := http.Get(fmt.Sprintf("%s/accounts/%d", ts.Server.URL, baseID+40000))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var accountResp GetAccountResponse
			err = json.NewDecoder(resp.Body).Decode(&accountResp)
			require.NoError(t, err)
			assert.Equal(t, baseID+40000, accountResp.AccountID)

			balance, err := strconv.ParseFloat(accountResp.Balance, 64)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, balance, 0.0, "Account has negative balance: %s", accountResp.Balance)
		}()
	}

	destAccount := CreateAccountRequest{
		AccountID:      baseID + 40001,
		InitialBalance: "0.00",
	}

	reqBody, err = json.Marshal(destAccount)
	require.NoError(t, err)

	resp, err = http.Post(
		fmt.Sprintf("%s/accounts", ts.Server.URL),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := CreateTransactionRequest{
				SourceAccountID:      baseID + 40000,
				DestinationAccountID: baseID + 40001,
				Amount:              "0.500",
			}

			reqBody, err := json.Marshal(req)
			require.NoError(t, err)

			resp, err := http.Post(
				fmt.Sprintf("%s/transactions", ts.Server.URL),
				"application/json",
				bytes.NewBuffer(reqBody),
			)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.True(t, resp.StatusCode == http.StatusOK,
				"Expected 200, got %d", resp.StatusCode)
		}()
	}

	wg.Wait()

	time.Sleep(100 * time.Millisecond)

	for _, accountID := range []int{baseID + 40000, baseID + 40001} {
		resp, err := http.Get(fmt.Sprintf("%s/accounts/%d", ts.Server.URL, accountID))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var accountResp GetAccountResponse
		err = json.NewDecoder(resp.Body).Decode(&accountResp)
		require.NoError(t, err)
		assert.Equal(t, accountID, accountResp.AccountID)

		balance, err := strconv.ParseFloat(accountResp.Balance, 64)
		require.NoError(t, err)
		if accountID == baseID + 40000 {
			assert.Equal(t, balance, 0.0, "Account %d should have balance 0.0, got %f", accountID, balance)
		} else {
			assert.Equal(t, balance, 500.0, "Account %d should have balance 500.0, got %f", accountID, balance)
		}
	}
}
