package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gwuah/accounts/internal/config"
	"github.com/lopezator/migrator"
)

const (
	POSTGRES  = "postgres"
	SQLITE    = "sqlite"
	MAX_CONNS = 10
)

type DB struct {
	_type    string
	instance *sql.DB
}

func (db *DB) Instance() *sql.DB {
	return db.instance
}

func New(ctx context.Context, config *config.Config, _type string) (*DB, error) {

	var db *sql.DB
	var err error

	switch _type {
	case POSTGRES:
		db, err = pqconn(ctx, config.DB_URL, postgresMigrations)
	case SQLITE:
		db, err = liteconn(ctx, config.DB_URL, sqliteMigrations)
	default:
		return nil, errors.New("unknown type")
	}
	return &DB{_type: _type, instance: db}, err
}

func pqconn(ctx context.Context, url string, opts ...migrator.Option) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(MAX_CONNS)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	m, err := migrator.New(opts...)
	if err != nil {
		return nil, err
	}
	if err := m.Migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func liteconn(ctx context.Context, url string, opts ...migrator.Option) (*sql.DB, error) {
	db, err := sql.Open("sqlite", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(MAX_CONNS)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	m, err := migrator.New(opts...)
	if err != nil {
		return nil, err
	}
	if err := m.Migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}
