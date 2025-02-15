package db

import (
	"log"
	"strconv"
)

func ClearDatabase(db *Database) {
	_, err := db.Pool.Exec(db.Ctx, "DELETE FROM transactions")
	if err != nil {
		log.Fatalf("Failed to clear transactions: %v", err)
	}
	_, err = db.Pool.Exec(db.Ctx, "DELETE FROM inventory")
	if err != nil {
		log.Fatalf("Failed to clear inventory: %v", err)
	}
	_, err = db.Pool.Exec(db.Ctx, "DELETE FROM users")
	if err != nil {
		log.Fatalf("Failed to clear users: %v", err)
	}
	_, err = db.Pool.Exec(db.Ctx, "DELETE FROM merch")
	if err != nil {
		log.Fatalf("Failed to clear users: %v", err)
	}

	fillMerchTable(db)
}

func fillMerchTable(db *Database) {
	insertMerchSQL := `
		INSERT INTO merch (name, price) VALUES
			('t-shirt', 80),
			('cup', 20),
			('book', 50),
			('pen', 10),
			('powerbank', 200),
			('hoody', 300),
			('umbrella', 200),
			('socks', 10),
			('wallet', 50),
			('pink-hoody', 500)
		ON CONFLICT (name) DO NOTHING;`
	_, err := db.Pool.Exec(db.Ctx, insertMerchSQL)
	if err != nil {
		log.Fatalf("Failed to insert default merch: %v", err)
	}
}

func CreateTables(db *Database) {
	tableCreationSQL := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			coins INTEGER DEFAULT 1000 CHECK (coins >= 0)
	 	);`,
		`CREATE TABLE IF NOT EXISTS merch (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			price INTEGER NOT NULL CHECK (price > 0)
	 	);`,
		`CREATE TABLE IF NOT EXISTS inventory (
			user_id INTEGER REFERENCES users(id),
			merch_id INTEGER REFERENCES merch(id),
			quantity INTEGER DEFAULT 1 CHECK (quantity > 0),
			PRIMARY KEY (user_id, merch_id)
		);`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			from_user_id INTEGER REFERENCES users(id),
			to_user_id INTEGER REFERENCES users(id),
			amount INTEGER NOT NULL CHECK (amount > 0),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, sqlStmt := range tableCreationSQL {
		_, err := db.Pool.Exec(db.Ctx, sqlStmt)
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
	}
}

func SetupTestDB(testDB **Database) {
	port, err := strconv.Atoi("5433") // TODO: move consts to params or env
	if err != nil {
		log.Fatalf("Wrong port: %v", err)
	}

	*testDB, err = NewDatabase("localhost", port, "testuser", "testpassword", "testdb")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	CreateTables(*testDB)
	fillMerchTable(*testDB)
}
