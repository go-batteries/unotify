package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
)

type PostgresDb struct {
	URL           string
	Dialect       string
	MigrationPath *string
}

type PostgresDbOpts func(*PostgresDb)

func WithMigrationPath(dirPath string) PostgresDbOpts {
	return func(pgdb *PostgresDb) {
		pgdb.MigrationPath = &dirPath
	}
}

func NewPostgresDb(dbURL string, opts ...PostgresDbOpts) *PostgresDb {
	pgdb := &PostgresDb{URL: dbURL, Dialect: "postrges"}

	for _, opt := range opts {
		opt(pgdb)
	}

	return pgdb
}

func (pgdb *PostgresDb) Connect(ctx context.Context) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", "user=foo dbname=bar sslmode=disable")
	if err != nil {
		logrus.WithError(err).Error("failed to connect to postgres db")
		return nil, err
	}

	err = db.PingContext(ctx)
	if err != nil {
		logrus.WithError(err).Error("database ping failed")
	}

	return db, err
}

var (
	ErrMirgationPathMissing = errors.New("migration_path_missing")
	ErrDialectNotSet        = errors.New("db_dialect_not_set")
	ErrMigrationFailed      = errors.New("db_migration_failed")
)

//go:embed migrations/postgres/*.sql
var embedMigrations embed.FS

func (pgdb *PostgresDb) Migrate(ctx context.Context, db *sql.DB, dir int) error {
	if pgdb.MigrationPath == nil {
		logrus.Error("migration path needed")
		return ErrMirgationPathMissing
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect(pgdb.Dialect); err != nil {
		logrus.WithError(err).Error("goose failed to load postrges dialiect")
		return ErrDialectNotSet
	}

	if dir > 0 {
		if err := goose.Up(db, *pgdb.MigrationPath); err != nil {
			logrus.WithError(err).Error("failed to migrate up")
			return ErrMigrationFailed
		}
	} else {
		if err := goose.Down(db, *pgdb.MigrationPath); err != nil {
			logrus.WithError(err).Error("failed to migrate down")
			return ErrMigrationFailed
		}
	}

	return nil
}
