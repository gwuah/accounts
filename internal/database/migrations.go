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
			`create table if not exists accounts (
				id SERIAL PRIMARY KEY,
				user_id INTEGER NOT NULL,
				account_number VARCHAR(100) UNIQUE NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);`,
		),
		execsql(
			"create_transactions",
			`create table if not exists transactions (
				id SERIAL PRIMARY KEY,
				reference VARCHAR(100) UNIQUE NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);`,
		),
		execsql(
			"create_transaction_lines",
			`create table if not exists transaction_lines (
				id SERIAL PRIMARY KEY,
				transaction_id INTEGER NOT NULL,
				account_id INTEGER NOT NULL,
				amount INTEGER NOT NULL,
				purpose VARCHAR(50) NOT NULL,
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
		execsql(
			"disable_updates_on_transaction_lines", "CREATE RULE no_updates_on_transaction_lines AS ON UPDATE TO transaction_lines DO INSTEAD NOTHING;",
		),
		// execsql(
		// 	"disable_deletes_on_transaction_lines", "CREATE RULE no_deletes_on_transaction_lines AS ON DELETE TO transaction_lines DO INSTEAD NOTHING;",
		// ),
	)
	sqliteMigrations = migrator.Migrations(

		execsql(
			"create_users",
			`create table if not exists users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				email VARCHAR(100) UNIQUE NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);`,
		),

		execsql(
			"create_accounts",
			`create table if not exists accounts (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				account_number VARCHAR(100) UNIQUE NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			);`,
		),

		execsql(
			"create_transactions",
			`create table if not exists transactions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				reference VARCHAR(100) UNIQUE NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);`,
		),

		execsql(
			"create_transaction_lines",
			`create table if not exists transaction_lines (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				transaction_id INTEGER NOT NULL,
				account_id INTEGER NOT NULL,
				amount INTEGER NOT NULL,
				purpose VARCHAR(50) NOT NULL,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
				FOREIGN KEY (transaction_id) REFERENCES transactions(id) ON DELETE CASCADE
			);`,
		),

		execsql(
			"create_unique_transaction_lines_index",
			"create unique index transaction_lines_unique_idx on transaction_lines(transaction_id, account_id);",
		),

		execsql(
			"create_transaction_lines_update_trigger",
			`CREATE TRIGGER prevent_transaction_lines_update
				BEFORE UPDATE ON transaction_lines
				BEGIN
					SELECT RAISE(FAIL, 'Updates to transaction_lines are not allowed.');
				END;`,
		),
	)
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

func RunSeeds(db *sql.DB) error {
	// create 1 user for the bank
	// create 1 account for the banks user
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("insert into users (email) values ($1) on conflict do nothing;", "primary@accounts.com")
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("insert into accounts (user_id, account_number) values ($1,$2) on conflict do nothing;", 1, "000000000")
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
