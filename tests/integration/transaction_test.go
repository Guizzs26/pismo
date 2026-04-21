package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type createTransactionRequest struct {
	AccountID       int64   `json:"account_id"`
	OperationTypeID int     `json:"operation_type_id"`
	Amount          float64 `json:"amount"`
}

func TestTransactionEndpoints(t *testing.T) {
	t.Run("POST /transactions - Should create transaction (201 Created)", func(t *testing.T) {
		clearDB(t)
		setupAccount(t, "12345678900")

		reqBody := createTransactionRequest{
			AccountID:       1,
			OperationTypeID: 1,
			Amount:          150.75,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", testServer.URL+"/transactions", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "1i02931imlkamdazxchiu123")

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("POST /transactions - Should return 200 for idempotent request with same key", func(t *testing.T) {
		clearDB(t)
		setupAccount(t, "11122233344")

		reqBody := createTransactionRequest{
			AccountID:       1,
			OperationTypeID: 4,
			Amount:          50.00,
		}
		bodyBytes, _ := json.Marshal(reqBody)
		idempotencyKey := "1i02931imlkamdazxchiu123"

		req1, _ := http.NewRequest("POST", testServer.URL+"/transactions", bytes.NewBuffer(bodyBytes))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("Idempotency-Key", idempotencyKey)

		res1, err := http.DefaultClient.Do(req1)
		assert.NoError(t, err)
		defer res1.Body.Close()
		assert.Equal(t, http.StatusCreated, res1.StatusCode)

		req2, _ := http.NewRequest("POST", testServer.URL+"/transactions", bytes.NewBuffer(bodyBytes))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Idempotency-Key", idempotencyKey)

		res2, err := http.DefaultClient.Do(req2)
		assert.NoError(t, err)
		defer res2.Body.Close()
		assert.Equal(t, http.StatusOK, res2.StatusCode)
	})

	t.Run("POST /transactions - Should return 409 Conflict if idempotency key is hijacked by another account", func(t *testing.T) {
		clearDB(t)
		setupAccount(t, "11111111111")
		setupAccount(t, "22222222222")

		idempotencyKey := "1i02931imlkamdazxchiu123"

		reqBody1 := createTransactionRequest{AccountID: 1, OperationTypeID: 1, Amount: 100.00}
		bodyBytes1, _ := json.Marshal(reqBody1)

		req1, _ := http.NewRequest("POST", testServer.URL+"/transactions", bytes.NewBuffer(bodyBytes1))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("Idempotency-Key", idempotencyKey)
		res1, err := http.DefaultClient.Do(req1)
		assert.NoError(t, err)
		defer res1.Body.Close()
		assert.Equal(t, http.StatusCreated, res1.StatusCode)

		reqBody2 := createTransactionRequest{AccountID: 2, OperationTypeID: 1, Amount: 100.00}
		bodyBytes2, _ := json.Marshal(reqBody2)

		req2, _ := http.NewRequest("POST", testServer.URL+"/transactions", bytes.NewBuffer(bodyBytes2))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Idempotency-Key", idempotencyKey)

		res2, err := http.DefaultClient.Do(req2)
		assert.NoError(t, err)
		defer res2.Body.Close()

		assert.Equal(t, http.StatusConflict, res2.StatusCode)
	})

	t.Run("POST /transactions - Should return 404 if account does not exist", func(t *testing.T) {
		clearDB(t)

		reqBody := createTransactionRequest{
			AccountID:       999,
			OperationTypeID: 1,
			Amount:          100.00,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", testServer.URL+"/transactions", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "1i02931imlkamdazxchiu123")

		res, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})
}

func setupAccount(t *testing.T, documentNumber string) {
	t.Helper()
	reqBody := map[string]string{"document_number": documentNumber}
	bodyBytes, _ := json.Marshal(reqBody)

	res, err := http.Post(testServer.URL+"/accounts", "application/json", bytes.NewBuffer(bodyBytes))
	assert.NoError(t, err)
	defer res.Body.Close()
	assert.Equal(t, http.StatusCreated, res.StatusCode, "Failed to setup account for transaction test")
}
