package db

import (
	"database/sql"
	"fmt"
	"time"


	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/kkr2/vessels/internal/errors"
	_ "github.com/lib/pq"

	_ "github.com/jackc/pgx/stdlib" // pgx driver
	"github.com/jmoiron/sqlx"
	"github.com/kkr2/vessels/internal/config"
)

const (
	maxOpenConns    = 60
	connMaxLifetime = 120
	maxIdleConns    = 30
	connMaxIdleTime = 20
)

// Return new Postgresql db instance
func NewPsqlDB(c *config.Config) (*sqlx.DB, error) {
	dataSourceName := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s",
		c.Postgres.PostgresqlHost,
		c.Postgres.PostgresqlPort,
		c.Postgres.PostgresqlUser,
		c.Postgres.PostgresqlDbname,
		c.Postgres.PostgresqlPassword,
	)
	operation := errors.Op("db.db_conn.NewPsqlDB")

	db, err := sqlx.Connect(c.Postgres.PgDriver, dataSourceName)
	if err != nil {
		return nil, errors.E(operation, errors.KindInternal, err)
	}
	err = runMigrations(db.DB)
	if err != nil {
		return nil, errors.E(operation, errors.KindInternal, err)
	}
	db.SetMaxOpenConns(maxOpenConns)
	db.SetConnMaxLifetime(connMaxLifetime * time.Second)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxIdleTime(connMaxIdleTime * time.Second)
	if err = db.Ping(); err != nil {
		return nil, errors.E(operation, errors.KindInternal, err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	operation := errors.Op("db.db_conn.NewPsqlDB.migration")
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return errors.E(operation, errors.KindInternal, err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver)
	if err != nil {
		return errors.E(operation, errors.KindInternal, err)
	}
	m.Up()
	
	return nil
}
