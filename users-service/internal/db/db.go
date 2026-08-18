// Package db db.go provides db methods via
// DB interface implementation
package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"usrsrv/internal/utils"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type DB interface {
	Register(nickname, password, reqTrace string) error
	Login(nickname, reqTrace string) (string, string, error)
	Delete(uuid, reqTrace string) error
	Shutdown(ctx context.Context) error
}

type UDB struct {
	db *sqlx.DB
	bd sq.StatementBuilderType
}

func NewUDB(log *zap.Logger) (DB, error) {
	const op = "db.NewUDB"

	return &UDB{
		db: nil,
		bd: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil

	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		return nil, fmt.Errorf("%s: get db url: no db_url in env", op)
	}
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		return nil, fmt.Errorf("%s: get db driver: no db_driver in env", op)
	}

	db, err := sqlx.Connect(driver, connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: connect to db: %w", op, err)
	}

	connLifetime := time.Duration(utils.GetEnvInt("DB_MAX_CONN_LIFETIME", 10))
	connIdletime := time.Duration(utils.GetEnvInt("DB_MAX_CONN_IDLETIME", 5))

	db.SetMaxIdleConns(utils.GetEnvInt("DB_MAX_IDLE_CONNS", 10))
	db.SetMaxOpenConns(utils.GetEnvInt("DB_MAX_OPEN_CONNS", 10))
	db.SetConnMaxLifetime(connLifetime * time.Minute)
	db.SetConnMaxIdleTime(connIdletime * time.Minute)

	return &UDB{
		db: db,
		bd: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

func (d *UDB) Shutdown(ctx context.Context) error {
	const op = "db.Shutdown"
	if err := d.db.Close(); err != nil {
		return fmt.Errorf("%s: close db: %w", op, err)
	}
	return nil
}

func (d *UDB) Register(nickname, password, reqTrace string) error {
	return nil
}

func (d *UDB) Login(nickname, reqTrace string) (string, string, error) {
	return "", "", nil
}

func (d *UDB) Delete(uuid, reqTrace string) error {
	return nil
}
