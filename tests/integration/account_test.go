package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type createAccountRequest struct {
	DocumentNumber string `json:"document_number"`
}

type accountResponse struct {
	AccountID      int64  `json:"account_id"`
	DocumentNumber string `json:"document_number"`
}

func TestAccountEndpoints(t *testing.T) {
	t.Run("POST /accounts - Should create a new account", func(t *testing.T) {
		clearDB(t)

		reqBody := createAccountRequest{DocumentNumber: "12345678900"}
		bodyBytes, _ := json.Marshal(reqBody)

		res, err := http.Post(testServer.URL+"/accounts", "application/json", bytes.NewBuffer(bodyBytes))
		assert.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusCreated, res.StatusCode)

		var accRes accountResponse
		err = json.NewDecoder(res.Body).Decode(&accRes)
		assert.NoError(t, err)

		assert.Equal(t, int64(1), accRes.AccountID)
		assert.Equal(t, "12345678900", accRes.DocumentNumber)
	})

	t.Run("POST /accounts - Should return 409 for duplicate document", func(t *testing.T) {
		clearDB(t)

		reqBody := createAccountRequest{DocumentNumber: "99999999999"}
		bodyBytes, _ := json.Marshal(reqBody)

		http.Post(testServer.URL+"/accounts", "application/json", bytes.NewBuffer(bodyBytes))

		resp, err := http.Post(testServer.URL+"/accounts", "application/json", bytes.NewBuffer(bodyBytes))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})

	t.Run("GET /accounts/{id} - Should return existing account", func(t *testing.T) {
		clearDB(t)

		reqBody := createAccountRequest{DocumentNumber: "11122233344"}
		bodyBytes, _ := json.Marshal(reqBody)
		http.Post(testServer.URL+"/accounts", "application/json", bytes.NewBuffer(bodyBytes))

		resp, err := http.Get(testServer.URL + "/accounts/1")
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var accResp accountResponse
		json.NewDecoder(resp.Body).Decode(&accResp)
		assert.Equal(t, int64(1), accResp.AccountID)
		assert.Equal(t, "11122233344", accResp.DocumentNumber)
	})

	t.Run("GET /accounts - Should return 409 for duplicate document", func(t *testing.T) {
		clearDB(t)

		resp, err := http.Get(testServer.URL + "/accounts/999")
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
