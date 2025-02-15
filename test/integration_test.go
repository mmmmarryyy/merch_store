package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"merch_store/internal/auth"
	"merch_store/internal/db"
	"merch_store/internal/handlers"
	"merch_store/internal/models"
)

var (
	testDB  *db.Database
	handler *handlers.Handler
	router  *mux.Router
)

func setupHandler() {
	handler = handlers.NewHandler(testDB)
	router = mux.NewRouter()

	router.HandleFunc("/api/auth", handler.AuthHandler)
	router.HandleFunc("/api/info", handler.InfoHandler)
	router.HandleFunc("/api/sendCoin", handler.SendCoinHandler)
	router.HandleFunc("/api/buy/{item}", handler.BuyHandler)
}

func executeRequest(req http.Request) httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, &req)
	return *rr
}

func generateAuthToken(username string) string {
	token, err := auth.GenerateToken(username)
	if err != nil {
		log.Fatalf("Failed to generate token: %v", err)
	}
	return token
}

func TestMain(m *testing.M) {
	db.SetupTestDB(&testDB)
	setupHandler()
	code := m.Run()

	os.Exit(code)
	db.ClearDatabase(testDB)
	testDB.Close()
}

func TestBuyMerchScenario(t *testing.T) {
	db.ClearDatabase(testDB)

	// 1. Authenticate user
	authReqBody := models.AuthRequest{Username: "testuser", Password: "testpassword"}
	authReqBytes, _ := json.Marshal(authReqBody)
	authReq, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(authReqBytes))
	authResp := executeRequest(*authReq)

	assert.Equal(t, http.StatusOK, authResp.Code)
	var authResponse models.AuthResponse
	err := json.Unmarshal(authResp.Body.Bytes(), &authResponse)
	assert.NoError(t, err)
	authToken := authResponse.Token

	// 2. Buy an item
	buyReq, _ := http.NewRequest("POST", "/api/buy/t-shirt", nil)
	buyReq.Header.Set("Authorization", authToken)
	buyResp := executeRequest(*buyReq)

	fmt.Println(buyResp.Body)
	assert.Equal(t, http.StatusOK, buyResp.Code)

	// 3. Check user's inventory
	infoReq, _ := http.NewRequest("GET", "/api/info", nil)
	infoReq.Header.Set("Authorization", authToken)
	infoResp := executeRequest(*infoReq)

	assert.Equal(t, http.StatusOK, infoResp.Code)

	var infoResponse models.InfoResponse
	err = json.Unmarshal(infoResp.Body.Bytes(), &infoResponse)
	assert.NoError(t, err)

	found := false
	for _, item := range infoResponse.Inventory {
		if item.Type == "t-shirt" {
			found = true
			assert.Equal(t, 1, item.Quantity)
			break
		}
	}
	assert.True(t, found, "T-shirt should be in inventory")

	// 4. Verify coins
	user, err := testDB.GetUserByUsername("testuser")
	assert.NoError(t, err)
	assert.Equal(t, 920, user.Coins, "User coins should be updated")
}

func TestSendCoinScenario(t *testing.T) {
	db.ClearDatabase(testDB)

	// 1. Authenticate sender
	authReqBodySender := models.AuthRequest{Username: "testuser", Password: "testpassword"}
	authReqBytesSender, _ := json.Marshal(authReqBodySender)
	authReqSender, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(authReqBytesSender))
	authRespSender := executeRequest(*authReqSender)

	assert.Equal(t, http.StatusOK, authRespSender.Code)
	var authResponseSender models.AuthResponse
	err := json.Unmarshal(authRespSender.Body.Bytes(), &authResponseSender)
	assert.NoError(t, err)
	authTokenSender := authResponseSender.Token

	// 2. Authenticate recipient (optional, but good practice)
	authReqBodyRecipient := models.AuthRequest{Username: "recipient", Password: "testpassword"}
	authReqBytesRecipient, _ := json.Marshal(authReqBodyRecipient)
	authReqRecipient, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(authReqBytesRecipient))
	authRespRecipient := executeRequest(*authReqRecipient)

	assert.Equal(t, http.StatusOK, authRespRecipient.Code)
	var authResponseRecipient models.AuthResponse
	err = json.Unmarshal(authRespRecipient.Body.Bytes(), &authResponseRecipient)
	assert.NoError(t, err)
	authTokenRecipient := authResponseRecipient.Token

	// 3. Send coins
	sendCoinReqBody := models.SendCoinRequest{ToUser: "recipient", Amount: 50}
	sendCoinReqBytes, _ := json.Marshal(sendCoinReqBody)
	sendCoinReq, _ := http.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(sendCoinReqBytes))
	sendCoinReq.Header.Set("Authorization", authTokenSender)
	sendCoinResp := executeRequest(*sendCoinReq)

	assert.Equal(t, http.StatusOK, sendCoinResp.Code)

	// 4. Verify sender's coins
	infoReqSender, _ := http.NewRequest("GET", "/api/info", nil)
	infoReqSender.Header.Set("Authorization", authTokenSender)
	infoRespSender := executeRequest(*infoReqSender)

	assert.Equal(t, http.StatusOK, infoRespSender.Code)
	var infoResponseSender models.InfoResponse
	err = json.Unmarshal(infoRespSender.Body.Bytes(), &infoResponseSender)
	assert.NoError(t, err)
	assert.Equal(t, 950, infoResponseSender.Coins, "Sender coins should be updated")

	// 5. Verify recipient's coins
	infoReqRecipient, _ := http.NewRequest("GET", "/api/info", nil)
	infoReqRecipient.Header.Set("Authorization", authTokenRecipient)
	infoRespRecipient := executeRequest(*infoReqRecipient)

	assert.Equal(t, http.StatusOK, infoRespRecipient.Code)
	var infoResponseRecipient models.InfoResponse
	err = json.Unmarshal(infoRespRecipient.Body.Bytes(), &infoResponseRecipient)
	assert.NoError(t, err)
	assert.Equal(t, 1050, infoResponseRecipient.Coins, "Recipient coins should be updated")
}

func TestBuyMerch_InsufficientFunds(t *testing.T) {
	db.ClearDatabase(testDB)

	// 1. Authenticate user
	authReqBody := models.AuthRequest{Username: "testuser", Password: "testpassword"}
	authReqBytes, _ := json.Marshal(authReqBody)
	authReq, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(authReqBytes))
	authResp := executeRequest(*authReq)

	assert.Equal(t, http.StatusOK, authResp.Code)
	var authResponse models.AuthResponse
	err := json.Unmarshal(authResp.Body.Bytes(), &authResponse)
	assert.NoError(t, err)
	authToken := authResponse.Token

	// 2. Attempt to buy the expensive item multiple times to exhaust funds
	item := "pink-hoody"
	for i := 0; i < 2; i++ { //Buy 2 pink-hoodies
		buyReq, _ := http.NewRequest("POST", "/api/buy/"+item, nil)
		buyReq.Header.Set("Authorization", authToken)
		buyResp := executeRequest(*buyReq)

		assert.Equal(t, http.StatusOK, buyResp.Code)
	}

	// Attempt the final buy that will exhaust the funds
	buyReq, _ := http.NewRequest("POST", "/api/buy/"+item, nil)
	buyReq.Header.Set("Authorization", authToken)
	buyResp := executeRequest(*buyReq)
	assert.Equal(t, http.StatusBadRequest, buyResp.Code)
	assert.Contains(t, buyResp.Body.String(), "Insufficient coins")
}

func TestBuyMerch_NonExistentItem(t *testing.T) {
	db.ClearDatabase(testDB)

	// 1. Authenticate user
	authReqBody := models.AuthRequest{Username: "testuser", Password: "testpassword"}
	authReqBytes, _ := json.Marshal(authReqBody)
	authReq, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(authReqBytes))
	authResp := executeRequest(*authReq)

	assert.Equal(t, http.StatusOK, authResp.Code)
	var authResponse models.AuthResponse
	err := json.Unmarshal(authResp.Body.Bytes(), &authResponse)
	assert.NoError(t, err)
	authToken := authResponse.Token

	// 2. Attempt to buy a nonexistent item
	buyReq, _ := http.NewRequest("POST", "/api/buy/nonexistent-item", nil)
	buyReq.Header.Set("Authorization", authToken)
	buyResp := executeRequest(*buyReq)

	assert.Equal(t, http.StatusNotFound, buyResp.Code)
	assert.Contains(t, buyResp.Body.String(), "Item not found")
}

func TestSendCoin_NonExistentRecipient(t *testing.T) {
	db.ClearDatabase(testDB)

	// 1. Authenticate sender
	authReqBodySender := models.AuthRequest{Username: "testuser", Password: "testpassword"}
	authReqBytesSender, _ := json.Marshal(authReqBodySender)
	authReqSender, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(authReqBytesSender))
	authRespSender := executeRequest(*authReqSender)

	assert.Equal(t, http.StatusOK, authRespSender.Code)
	var authResponseSender models.AuthResponse
	err := json.Unmarshal(authRespSender.Body.Bytes(), &authResponseSender)
	assert.NoError(t, err)
	authTokenSender := authResponseSender.Token

	// 2. Attempt to send coins to a nonexistent recipient
	sendCoinReqBody := models.SendCoinRequest{ToUser: "nonexistent", Amount: 50}
	sendCoinReqBytes, _ := json.Marshal(sendCoinReqBody)
	sendCoinReq, _ := http.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(sendCoinReqBytes))
	sendCoinReq.Header.Set("Authorization", authTokenSender)
	sendCoinResp := executeRequest(*sendCoinReq)

	assert.Equal(t, http.StatusNotFound, sendCoinResp.Code)
	assert.Contains(t, sendCoinResp.Body.String(), "Recipient not found")
}

func TestSendCoin_InsufficientFunds(t *testing.T) {
	db.ClearDatabase(testDB)

	// 1. Authenticate sender
	authReqBodySender := models.AuthRequest{Username: "testuser", Password: "testpassword"}
	authReqBytesSender, _ := json.Marshal(authReqBodySender)
	authReqSender, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(authReqBytesSender))
	authRespSender := executeRequest(*authReqSender)

	assert.Equal(t, http.StatusOK, authRespSender.Code)
	var authResponseSender models.AuthResponse
	err := json.Unmarshal(authRespSender.Body.Bytes(), &authResponseSender)
	assert.NoError(t, err)
	authTokenSender := authResponseSender.Token

	// 2. Authenticate receiver
	authReqBodySender = models.AuthRequest{Username: "recipient", Password: "recipientpassword"}
	authReqBytesSender, _ = json.Marshal(authReqBodySender)
	authReqSender, _ = http.NewRequest("POST", "/api/auth", bytes.NewBuffer(authReqBytesSender))
	authRespSender = executeRequest(*authReqSender)

	assert.Equal(t, http.StatusOK, authRespSender.Code)

	// 3. Attempt to send more coins than available
	sendCoinReqBody := models.SendCoinRequest{ToUser: "recipient", Amount: 2000}
	sendCoinReqBytes, _ := json.Marshal(sendCoinReqBody)
	sendCoinReq, _ := http.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(sendCoinReqBytes))
	sendCoinReq.Header.Set("Authorization", authTokenSender)
	sendCoinResp := executeRequest(*sendCoinReq)

	assert.Equal(t, http.StatusBadRequest, sendCoinResp.Code)
	assert.Contains(t, sendCoinResp.Body.String(), "Insufficient coins")
}

func TestInfoHandler_Unauthorized(t *testing.T) {
	db.ClearDatabase(testDB)

	// 1. Make a request to /api/info without a valid token
	infoReq, _ := http.NewRequest("GET", "/api/info", nil)
	infoResp := executeRequest(*infoReq)

	assert.Equal(t, http.StatusUnauthorized, infoResp.Code)
	assert.Contains(t, infoResp.Body.String(), "Unauthorized")

	// 2. Make a request to /api/info with an invalid token format
	infoReqInvalid, _ := http.NewRequest("GET", "/api/info", nil)
	infoReqInvalid.Header.Set("Authorization", "invalid-token-format") // Missing "Bearer"
	infoRespInvalid := executeRequest(*infoReqInvalid)

	assert.Equal(t, http.StatusUnauthorized, infoRespInvalid.Code)
	assert.Contains(t, infoRespInvalid.Body.String(), "Unauthorized")

	// 3. Make a request to /api/info with an expired/invalid token
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwIjoxNjYyMjgxNjAwfQ.1234567890abcdef1234567890abcdef" // Example expired token
	infoReqExpired, _ := http.NewRequest("GET", "/api/info", nil)
	infoReqExpired.Header.Set("Authorization", expiredToken)
	infoRespExpired := executeRequest(*infoReqExpired)

	assert.Equal(t, http.StatusUnauthorized, infoRespExpired.Code)
	assert.Contains(t, infoRespExpired.Body.String(), "Unauthorized")
}
