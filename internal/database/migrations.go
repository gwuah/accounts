package database

import (
	"database/sql"

	"github.com/lopezator/migrator"
)

var (
	postgresMigrations = migrator.Migrations(
		execsql(
			"create_users",
			`create table if not exists users (
				id SERIAL PRIMARY KEY,
				email VARCHAR(100) UNIQUE NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);`,
		),
		execsql(
			"create_accounts",
			`create table if not exists CREATE TABLE accounts (
				id SERIAL PRIMARY KEY,
				user_id INTEGER NOT NULL,
				account_number INTEGER UNIQUE NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);`,
		),
		execsql(
			"create_transactions",
			`create table if not exists CREATE TABLE transactions (
				id SERIAL PRIMARY KEY,
				reference VARCHAR(100) UNIQUE NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);`,
		),
		execsql(
			"create_transaction_lines",
			`create table if not exists CREATE TABLE transaction_lines (
				id SERIAL PRIMARY KEY,
				transaction_id INTEGER NOT NULL,
				reference VARCHAR(100) NOT NULL,
				account_id INTEGER NOT NULL,
				amount INTEGER NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
				FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE
			);`,
		),
		execsql(
			"create_unique_transaction_lines_index",
			"create unique index transaction_lines_unique_idx on transaction_lines(transaction_id, account_id);",
		),
	)
	sqliteMigrations = migrator.Migrations()
)

func execsql(name, raw string) *migrator.MigrationNoTx {
	return &migrator.MigrationNoTx{
		Name: name,
		Func: func(db *sql.DB) error {
			_, err := db.Exec(raw)
			return err
		},
	}
}