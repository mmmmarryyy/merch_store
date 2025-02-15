package db

import (
	"merch_store/internal/models"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

var testDB *Database

func TestMain(m *testing.M) {
	SetupTestDB(&testDB)
	code := m.Run()

	ClearDatabase(testDB)
	_ = testDB.Close()
	os.Exit(code)
}

func TestGetUserByUsername(t *testing.T) {
	ClearDatabase(testDB)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	_, err := testDB.Pool.Exec(testDB.Ctx, "INSERT INTO users (username, password_hash, coins) VALUES ('testuser', $1, 100)", hashedPassword)
	assert.NoError(t, err)

	user, err := testDB.GetUserByUsername("testuser")
	assert.NoError(t, err)

	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, 100, user.Coins)
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	ClearDatabase(testDB)

	user, err := testDB.GetUserByUsername("nonexistent")
	assert.Error(t, err)
	assert.Nil(t, user)
}

func TestCreateUser(t *testing.T) {
	ClearDatabase(testDB)

	user := &models.User{Username: "newuser", PasswordHash: "hash", Coins: 500}
	err := testDB.CreateUser(user)
	assert.NoError(t, err)

	retrievedUser, err := testDB.GetUserByUsername("newuser")
	assert.NoError(t, err)
	assert.Equal(t, "newuser", retrievedUser.Username)
	assert.Equal(t, 1000, retrievedUser.Coins)
}

func TestTransferCoins(t *testing.T) {
	ClearDatabase(testDB)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	_, err := testDB.Pool.Exec(testDB.Ctx, "INSERT INTO users (username, password_hash, coins) VALUES ('user1', $1, 100)", hashedPassword)
	assert.NoError(t, err)
	_, err = testDB.Pool.Exec(testDB.Ctx, "INSERT INTO users (username, password_hash, coins) VALUES ('user2', $1, 50)", hashedPassword)
	assert.NoError(t, err)

	user1, _ := testDB.GetUserByUsername("user1")
	user2, _ := testDB.GetUserByUsername("user2")

	err = testDB.TransferCoins(user1.ID, user2.ID, 30)
	assert.NoError(t, err)

	updatedUser1, _ := testDB.GetUserByUsername("user1")
	updatedUser2, _ := testDB.GetUserByUsername("user2")

	assert.Equal(t, 70, updatedUser1.Coins)
	assert.Equal(t, 80, updatedUser2.Coins)
}

func TestGetMerchByName(t *testing.T) {
	ClearDatabase(testDB)
	_, err := testDB.Pool.Exec(testDB.Ctx, "INSERT INTO merch (name, price) VALUES ('special-item', 150)")
	assert.NoError(t, err)

	merch, err := testDB.GetMerchByName("special-item")

	assert.NoError(t, err)
	assert.Equal(t, "special-item", merch.Name)
	assert.Equal(t, 150, merch.Price)
}

func TestBuyMerch(t *testing.T) {
	ClearDatabase(testDB)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	_, err := testDB.Pool.Exec(testDB.Ctx, "INSERT INTO users (username, password_hash, coins) VALUES ('buyer', $1, 200)", hashedPassword)
	assert.NoError(t, err)
	_, err = testDB.Pool.Exec(testDB.Ctx, "INSERT INTO merch (name, price) VALUES ('fancy-item', 120)")
	assert.NoError(t, err)

	buyer, _ := testDB.GetUserByUsername("buyer")
	merch, _ := testDB.GetMerchByName("fancy-item")

	err = testDB.BuyMerch(buyer.ID, merch.ID, merch.Price)
	assert.NoError(t, err)

	inventory, _ := testDB.GetUserInventory(buyer.ID)
	assert.Len(t, inventory, 1)
	assert.Equal(t, "fancy-item", inventory[0].Type)
	assert.Equal(t, 1, inventory[0].Quantity)

	updatedBuyer, _ := testDB.GetUserByUsername("buyer")
	assert.Equal(t, 80, updatedBuyer.Coins)
}

func TestGetUserInventory(t *testing.T) {
	ClearDatabase(testDB)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	_, err := testDB.Pool.Exec(testDB.Ctx, "INSERT INTO users (username, password_hash, coins) VALUES ('inventory_user', $1, 100)", hashedPassword)
	assert.NoError(t, err)
	_, err = testDB.Pool.Exec(testDB.Ctx, "INSERT INTO merch (name, price) VALUES ('inventory_item', 50)")
	assert.NoError(t, err)

	user, _ := testDB.GetUserByUsername("inventory_user")
	merch, _ := testDB.GetMerchByName("inventory_item")

	_, err = testDB.Pool.Exec(testDB.Ctx, "INSERT INTO inventory (user_id, merch_id, quantity) VALUES ($1, $2, 3)", user.ID, merch.ID)
	assert.NoError(t, err)

	inventory, err := testDB.GetUserInventory(user.ID)
	assert.NoError(t, err)
	assert.Len(t, inventory, 1)
	assert.Equal(t, "inventory_item", inventory[0].Type)
	assert.Equal(t, 3, inventory[0].Quantity)
}

func TestGetUserTransactions(t *testing.T) {
	ClearDatabase(testDB)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	_, err := testDB.Pool.Exec(testDB.Ctx, "INSERT INTO users (username, password_hash, coins) VALUES ('sender', $1, 100)", hashedPassword)
	assert.NoError(t, err)
	_, err = testDB.Pool.Exec(testDB.Ctx, "INSERT INTO users (username, password_hash, coins) VALUES ('receiver', $1, 50)", hashedPassword)
	assert.NoError(t, err)

	sender, _ := testDB.GetUserByUsername("sender")
	receiver, _ := testDB.GetUserByUsername("receiver")

	err = testDB.TransferCoins(sender.ID, receiver.ID, 25)
	assert.NoError(t, err)

	history, err := testDB.GetUserTransactions(sender.ID)
	assert.NoError(t, err)
	assert.Len(t, history.Sent, 1)
	assert.Equal(t, "receiver", history.Sent[0].Username)
	assert.Equal(t, 25, history.Sent[0].Amount)

	history, err = testDB.GetUserTransactions(receiver.ID)
	assert.NoError(t, err)
	assert.Len(t, history.Received, 1)
	assert.Equal(t, "sender", history.Received[0].Username)
	assert.Equal(t, 25, history.Received[0].Amount)
}
