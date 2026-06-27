package app

import (
	"database/sql"
	"errors"
	"fmt"

	dbmigrations "github.com/Rebne/scrapy_project_v2/migrations"
	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func RunMigrations(databaseURL string) (err error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open migration database connection: %w", err)
	}
	defer func() {
		err = errors.Join(err, db.Close())
	}()

	driver, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{})
	if err != nil {
		return fmt.Errorf("create pgx migration driver: %w", err)
	}

	sourceDriver, err := iofs.New(dbmigrations.Files, ".")
	if err != nil {
		return fmt.Errorf("load embedded migrations: %w", err)
	}

	migrator, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer func() {
		sourceErr, dbErr := migrator.Close()
		err = errors.Join(err, sourceErr, dbErr)
	}()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}
