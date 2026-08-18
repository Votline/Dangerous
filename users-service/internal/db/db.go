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
	Register(nickname, password, reqTrace string, ctx context.Context) error
	Get(nickname, reqTrace string, ctx context.Context) (string, error)
	Delete(nickname, reqTrace string, ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type UDB struct {
	tableName string
	log       *zap.Logger
	db        *sqlx.DB
	bd        sq.StatementBuilderType
}

func NewUDB(log *zap.Logger) (DB, error) {
	const op = "db.NewUDB"

	connStr := os.Getenv("DB_URL")
	if connStr == "" {
		return nil, fmt.Errorf("%s: get db url: no db_url in env", op)
	}
	driver := os.Getenv("DB_DRIVER")
	if driver == "" {
		return nil, fmt.Errorf("%s: get db driver: no db_driver in env", op)
	}
	tableName := os.Getenv("DB_TABLE_NAME")
	if tableName == "" {
		return nil, fmt.Errorf("%s: get table name: no table name", op)
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
		log:       log,
		db:        db,
		tableName: tableName,
		bd:        sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

func (d *UDB) Shutdown(ctx context.Context) error {
	const op = "db.Shutdown"
	if err := d.db.Close(); err != nil {
		return fmt.Errorf("%s: close db: %w", op, err)
	}
	return nil
}

func (d *UDB) Register(nickname, password, reqTrace string, ctx context.Context) error {
	const op = "db.Register"

	d.log.Info("Register request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	query, args, err := d.bd.Insert(d.tableName).
		Columns("nickname", "password").
		Values(nickname, password).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: insert to db: %w", op, err)
	}

	d.log.Info("Query created",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	if _, err := d.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("%s: insert to db: %w", op, err)
	}

	d.log.Info("Successfully registred user",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	return nil
}

func (d *UDB) Get(nickname, reqTrace string, ctx context.Context) (string, error) {
	const op = "db.Get"

	d.log.Info("Get request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	query, args, err := d.bd.Select("password").
		From(d.tableName).
		Where(sq.Eq{"nickname": nickname}).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("%s: insert to db: %w", op, err)
	}

	d.log.Info("Query created",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	var password string
	if err := d.db.GetContext(ctx, &password, query, args...); err != nil {
		return "", fmt.Errorf("%s: insert to db: %w", op, err)
	}

	d.log.Info("Successfully logged in",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	return password, nil
}

func (d *UDB) Delete(nickname, reqTrace string, ctx context.Context) error {
	const op = "db.Delete"

	d.log.Info("Delete request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	query, args, err := d.bd.Delete(d.tableName).
		Where(sq.Eq{"nickname": nickname}).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: insert to db: %w", op, err)
	}

	d.log.Info("Query created",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	if _, err := d.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("%s: insert to db: %w", op, err)
	}

	d.log.Info("Successfully deleted",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	return nil
}
