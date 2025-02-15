package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"merch_store/internal/auth"
	"merch_store/internal/db"
	"merch_store/internal/models"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

var testDB *db.Database
var handler *Handler
var router *mux.Router

func setupHandler() {
	handler = NewHandler(testDB)
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

func TestAuthHandler_CreateUser(t *testing.T) {
	db.ClearDatabase(testDB)

	reqBody := models.AuthRequest{Username: "newuser", Password: "password"}
	reqBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(reqBytes))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	user, err := testDB.GetUserByUsername(reqBody.Username)
	assert.NoError(t, err)
	assert.Equal(t, reqBody.Username, user.Username)
}

func TestAuthHandler_Login(t *testing.T) {
	db.ClearDatabase(testDB)

	reqBody := models.AuthRequest{Username: "existinguser", Password: "password"}
	reqBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/auth", bytes.NewBuffer(reqBytes))
	w := httptest.NewRecorder()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reqBody.Password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	_, err = testDB.Pool.Exec(testDB.Ctx, "INSERT INTO users (username, password_hash, coins) VALUES ($1, $2, 1000)", reqBody.Username, hashedPassword)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInfoHandler(t *testing.T) {
	db.ClearDatabase(testDB)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	_, err := testDB.Pool.Exec(testDB.Ctx, "INSERT INTO users (username, password_hash, coins) VALUES ('testuser', $1, 200)", hashedPassword)
	assert.NoError(t, err)

	user, _ := testDB.GetUserByUsername("testuser")
	_, err = testDB.Pool.Exec(testDB.Ctx, "INSERT INTO merch (name, price) VALUES ('testitem_for_infohandler', 50)")
	assert.NoError(t, err)

	merch, _ := testDB.GetMerchByName("testitem_for_infohandler")
	_, err = testDB.Pool.Exec(testDB.Ctx, "INSERT INTO inventory (user_id, merch_id, quantity) VALUES ($1, $2, 5)", user.ID, merch.ID)
	assert.NoError(t, err)

	token := generateAuthToken("testuser") // Generate token
	req, _ := http.NewRequest("GET", "/api/info", nil)
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var infoResponse models.InfoResponse
	err = json.Unmarshal(w.Body.Bytes(), &infoResponse)
	assert.NoError(t, err)

	assert.Equal(t, 200, infoResponse.Coins)
	assert.Len(t, infoResponse.Inventory, 1)
	assert.Equal(t, "testitem_for_infohandler", infoResponse.Inventory[0].Type)
	assert.Equal(t, 5, infoResponse.Inventory[0].Quantity)
}

func TestSendCoinHandler(t *testing.T) {
	db.ClearDatabase(testDB)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	_, err := testDB.Pool.Exec(testDB.Ctx, "INSERT INTO users (username, password_hash, coins) VALUES ('sender', $1, 100)", hashedPassword)
	assert.NoError(t, err)
	_, err = testDB.Pool.Exec(testDB.Ctx, "INSERT INTO users (username, password_hash, coins) VALUES ('receiver', $1, 50)", hashedPassword)
	assert.NoError(t, err)

	reqBody := models.SendCoinRequest{ToUser: "receiver", Amount: 20}
	reqBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/api/sendCoin", bytes.NewBuffer(reqBytes))

	token := generateAuthToken("sender")
	req.Header.Set("Authorization", token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	updatedSender, _ := testDB.GetUserByUsername("sender")
	updatedReceiver, _ := testDB.GetUserByUsername("receiver")

	assert.Equal(t, 80, updatedSender.Coins)
	assert.Equal(t, 70, updatedReceiver.Coins)
}

func TestBuyHandler(t *testing.T) {
	db.ClearDatabase(testDB)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	_, err := testDB.Pool.Exec(testDB.Ctx, "INSERT INTO users (username, password_hash, coins) VALUES ('buyer', $1, 200)", hashedPassword)
	assert.NoError(t, err)
	_, err = testDB.Pool.Exec(testDB.Ctx, "INSERT INTO merch (name, price) VALUES ('testitem_for_buyhandler', 50)")
	assert.NoError(t, err)

	req, _ := http.NewRequest("POST", "/api/buy/testitem_for_buyhandler", nil)
	token := generateAuthToken("buyer")
	req.Header.Set("Authorization", token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	updatedBuyer, _ := testDB.GetUserByUsername("buyer")
	assert.Equal(t, 150, updatedBuyer.Coins)

	inventory, _ := testDB.GetUserInventory(updatedBuyer.ID)
	assert.Len(t, inventory, 1)
	assert.Equal(t, "testitem_for_buyhandler", inventory[0].Type)
	assert.Equal(t, 1, inventory[0].Quantity)
}
