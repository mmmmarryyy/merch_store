// Package db - package for working with the PostgreSQL database.
package db

// TODO: split package into several files

import (
	"context"
	"fmt"
	"log"
	"merch_store/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB interface...
type DB interface {
	GetUserByUsername(username string) (*models.User, error)
	CreateUser(user *models.User) error
	TransferCoins(fromUserID, toUserID, amount int) error
	GetMerchByName(name string) (*models.Merch, error)
	BuyMerch(userID, merchID, price int) error
	GetUserInventory(userID int) ([]models.InventoryInfo, error)
	GetUserTransactions(userID int) (models.CoinHistory, error)
	Close() error
}

// Database ...
type Database struct {
	Pool *pgxpool.Pool
	Ctx  context.Context
}

// NewDatabase connects to database...
func NewDatabase(host string, port int, user string, password string, dbname string) (*Database, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, dbname)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Printf("Failed to parse config: %v", err)
		return nil, err
	}

	Ctx := context.Background()
	Pool, err := pgxpool.NewWithConfig(Ctx, config)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil, err
	}

	err = Pool.Ping(Ctx) // Ping the database to verify the connection
	if err != nil {
		log.Printf("Failed to ping database: %v", err)
		return nil, err
	}

	return &Database{Pool: Pool, Ctx: Ctx}, nil
}

// Close function closes database...
func (db *Database) Close() error {
	db.Pool.Close()
	return nil
}

// GetUserByUsername finds user by name in database...
func (db *Database) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := db.Pool.QueryRow(db.Ctx, "SELECT id, username, password_hash, coins FROM users WHERE username = $1", username).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Coins)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates user in database...
func (db *Database) CreateUser(user *models.User) error {
	_, err := db.Pool.Exec(db.Ctx, "INSERT INTO users (username,password_hash,coins) VALUES ($1,$2,1000);", user.Username, user.PasswordHash)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return err
	}
	return nil
}

// TransferCoins implements logic for sending coins from one user to another in database...
func (db *Database) TransferCoins(fromUserID, toUserID, amount int) error {
	tx, err := db.Pool.Begin(db.Ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if rb := tx.Rollback(db.Ctx); rb != nil {
				log.Fatalf("query failed: %v, unable to abort: %v", err, rb)
			}
		} else {
			err = tx.Commit(db.Ctx)
		}
	}()

	_, err = tx.Exec(db.Ctx, "UPDATE users SET coins = coins - $1 WHERE id = $2", amount, fromUserID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(db.Ctx, "UPDATE users SET coins = coins + $1 WHERE id = $2", amount, toUserID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(db.Ctx, "INSERT INTO transactions (from_user_id, to_user_id, amount) VALUES ($1, $2, $3)",
		fromUserID, toUserID, amount)
	if err != nil {
		return err
	}

	return nil
}

// GetMerchByName finds merch by it's name in database...
func (db *Database) GetMerchByName(name string) (*models.Merch, error) {
	var merch models.Merch
	err := db.Pool.QueryRow(db.Ctx, "SELECT id, name, price FROM merch WHERE name = $1", name).
		Scan(&merch.ID, &merch.Name, &merch.Price)
	if err != nil {
		return nil, err
	}
	return &merch, nil
}

// BuyMerch implements buying merch logic in database...
func (db *Database) BuyMerch(userID, merchID, price int) error {
	tx, err := db.Pool.Begin(db.Ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rb := tx.Rollback(db.Ctx); rb != nil {
				log.Fatalf("query failed: %v, unable to abort: %v", err, rb)
			}
		} else {
			err = tx.Commit(db.Ctx)
		}
	}()

	_, err = tx.Exec(db.Ctx, "UPDATE users SET coins = coins - $1 WHERE id = $2", price, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(db.Ctx, `
       INSERT INTO inventory (user_id, merch_id, quantity)
       VALUES ($1, $2, 1)
       ON CONFLICT (user_id, merch_id) DO UPDATE
       SET quantity = inventory.quantity + 1
   `, userID, merchID)
	if err != nil {
		return err
	}

	return nil
}

// GetUserInventory gets user inventory from database...
func (db *Database) GetUserInventory(userID int) ([]models.InventoryInfo, error) {
	rows, err := db.Pool.Query(db.Ctx, `
        SELECT m.name, i.quantity
        FROM inventory i
        JOIN merch m ON i.merch_id = m.id
        WHERE i.user_id = $1
    `, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventory []models.InventoryInfo
	for rows.Next() {
		var item models.InventoryInfo
		err = rows.Scan(&item.Type, &item.Quantity)
		if err != nil {
			return nil, err
		}
		inventory = append(inventory, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return inventory, nil
}

// GetUserTransactions gets user transactions from database...
func (db *Database) GetUserTransactions(userID int) (models.CoinHistory, error) {
	var history models.CoinHistory

	rows, err := db.Pool.Query(db.Ctx, `
        SELECT u.username, t.amount
        FROM transactions t
        JOIN users u ON t.from_user_id = u.id
        WHERE t.to_user_id = $1
   `, userID)

	if err != nil {
		return history, err
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.TransactionInfo
		err = rows.Scan(&transaction.Username, &transaction.Amount)
		if err != nil {
			return history, err
		}
		history.Received = append(history.Received, transaction)
	}

	if err := rows.Err(); err != nil {
		return history, err
	}

	rows, err = db.Pool.Query(db.Ctx, `
        SELECT u.username, t.amount
        FROM transactions t
        JOIN users u ON t.to_user_id = u.id
        WHERE t.from_user_id = $1
    `, userID)

	if err != nil {
		return history, err
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.TransactionInfo
		err = rows.Scan(&transaction.Username, &transaction.Amount)
		if err != nil {
			return history, err
		}
		history.Sent = append(history.Sent, transaction)
	}

	if err := rows.Err(); err != nil {
		return history, err
	}

	return history, nil
}
