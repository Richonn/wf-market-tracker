package db

import (
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"

	"github.com/Richonn/wf-market-tracker/migrations"
)

type DB struct {
	*sql.DB
}

func New(dsn string) (*DB, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	conn.SetMaxOpenConns(5)
	conn.SetMaxIdleConns(2)

	if err := runMigrations(conn); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &DB{conn}, nil
}

func runMigrations(conn *sql.DB) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(conn, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}
