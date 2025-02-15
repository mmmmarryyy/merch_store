// TODO: split into several files
package db

import (
	"context"
	"fmt"
	"log"
	"merch_store/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
	ctx  context.Context
}

func NewDatabase(host string, port int, user string, password string, dbname string) (*Database, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, dbname)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Printf("Failed to parse config: %v", err)
		return nil, err
	}

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil, err
	}

	err = pool.Ping(ctx) // Ping the database to verify the connection
	if err != nil {
		log.Printf("Failed to ping database: %v", err)
		return nil, err
	}

	return &Database{pool: pool, ctx: ctx}, nil
}

func (db *Database) Close() error {
	db.pool.Close()
	return nil
}

func (db *Database) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := db.pool.QueryRow(db.ctx, "SELECT id, username, password_hash, coins FROM users WHERE username = $1", username).
		Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Coins)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *Database) CreateUser(user *models.User) error {
	_, err := db.pool.Exec(db.ctx, "INSERT INTO users (username,password_hash,coins) VALUES ($1,$2,1000);", user.Username, user.PasswordHash)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return err
	}
	return nil
}

func (db *Database) TransferCoins(fromUserID, toUserID, amount int) error {
	tx, err := db.pool.Begin(db.ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(db.ctx)
		} else {
			err = tx.Commit(db.ctx)
		}
	}()

	_, err = tx.Exec(db.ctx, "UPDATE users SET coins = coins - $1 WHERE id = $2", amount, fromUserID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(db.ctx, "UPDATE users SET coins = coins + $1 WHERE id = $2", amount, toUserID)
	if err != nil { // TODO: add transaction to return coins to `fromUserID`
		return err
	}

	_, err = tx.Exec(db.ctx, "INSERT INTO transactions (from_user_id, to_user_id, amount) VALUES ($1, $2, $3)",
		fromUserID, toUserID, amount)
	if err != nil { // TODO: add transaction to return coins to `fromUserID` and delete coins from `toUserID`
		return err
	}

	return nil
}

func (db *Database) GetMerchByName(name string) (*models.Merch, error) {
	var merch models.Merch
	err := db.pool.QueryRow(db.ctx, "SELECT id, name, price FROM merch WHERE name = $1", name).
		Scan(&merch.ID, &merch.Name, &merch.Price)
	if err != nil {
		return nil, err
	}
	return &merch, nil
}

func (db *Database) BuyMerch(userID, merchID, price int) error {
	tx, err := db.pool.Begin(db.ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback(db.ctx)
		} else {
			err = tx.Commit(db.ctx)
		}
	}()

	_, err = tx.Exec(db.ctx, "UPDATE users SET coins = coins - $1 WHERE id = $2", price, userID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(db.ctx, `
       INSERT INTO inventory (user_id, merch_id, quantity)
       VALUES ($1, $2, 1)
       ON CONFLICT (user_id, merch_id) DO UPDATE
       SET quantity = inventory.quantity + 1
   `, userID, merchID)
	if err != nil { // TODO: add transaction to return coins to `userID`
		return err
	}

	return nil
}

func (db *Database) GetUserInventory(userID int) ([]models.Inventory, error) {
	rows, err := db.pool.Query(db.ctx, `
       SELECT merch_id, quantity
       FROM inventory
       WHERE user_id = $1
   `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventory []models.Inventory
	for rows.Next() {
		var item models.Inventory
		err = rows.Scan(&item.MerchID, &item.Quantity)
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

func (db *Database) GetUserTransactions(userID int) (models.CoinHistory, error) {
	var history models.CoinHistory

	rows, err := db.pool.Query(db.ctx, `
       SELECT from_user_id, amount
       FROM transactions
       WHERE to_user_id = $1
   `, userID)
	if err != nil {
		return history, err
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.Transaction
		err = rows.Scan(&transaction.FromUserID, &transaction.Amount)
		if err != nil {
			return history, err
		}
		history.Received = append(history.Received, transaction)
	}

	if err := rows.Err(); err != nil {
		return history, err
	}

	rows, err = db.pool.Query(db.ctx, `
       SELECT to_user_id, amount
       FROM transactions
       WHERE from_user_id = $1
   `, userID)
	if err != nil {
		return history, err
	}
	defer rows.Close()

	for rows.Next() {
		var transaction models.Transaction
		err = rows.Scan(&transaction.ToUserID, &transaction.Amount)
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
