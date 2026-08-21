// Package db db.go provides db methods via
// DB interface implementation
package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"mrksrv/internal/utils"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

type DB interface {
	New(nickname, comment, reqTrace string, lat, lng float64, ctx context.Context) error
	Get(lat, lng float64, reqTrace string, ctx context.Context) ([]AdditionalInfo, error)
	Delete(lat, lng float64, reqTrace string, ctx context.Context) error
	Shutdown(ctx context.Context) error
}

type MDB struct {
	tableName string
	log       *zap.Logger
	db        *sqlx.DB
	bd        sq.StatementBuilderType
}

type AdditionalInfo struct {
	Nickname string `db:"nickname"`
	Comment  string `db:"comment"`
}

func NewMDB(log *zap.Logger) (DB, error) {
	const op = "db.NewMDB"

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

	return &MDB{
		log:       log,
		db:        db,
		tableName: tableName,
		bd:        sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

func (d *MDB) Shutdown(ctx context.Context) error {
	const op = "db.Shutdown"
	if err := d.db.Close(); err != nil {
		return fmt.Errorf("%s: close db: %w", op, err)
	}
	return nil
}

func (d *MDB) New(nickname, comment, reqTrace string, lat, lng float64, ctx context.Context) error {
	const op = "db.New"

	d.log.Info("New request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	query, args, err := d.bd.Insert(d.tableName).
		Columns("nickname", "comment", "latitude", "longitude").
		Values(nickname, comment, lat, lng).
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

	d.log.Info("Successfully marked",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	return nil
}

func (d *MDB) Get(lat, lng float64, reqTrace string, ctx context.Context) ([]AdditionalInfo, error) {
	const op = "db.Get"

	d.log.Info("Get request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	query, args, err := d.bd.Select("nickname", "comment").
		From(d.tableName).
		Where(sq.Expr("ROUND(latitude::numeric, 3) = ?", lat)).
		Where(sq.Expr("ROUND(longitude::numeric, 3) = ?", lng)).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("%s: select from db: %w", op, err)
	}

	d.log.Info("Query created",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	var addinfo []AdditionalInfo

	if err := d.db.SelectContext(ctx, &addinfo, query, args...); err != nil {
		return nil, fmt.Errorf("%s: insert to db: %w", op, err)
	}

	d.log.Info("Successfully getted mark",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	return addinfo, nil
}

func (d *MDB) Delete(lat, lng float64, reqTrace string, ctx context.Context) error {
	const op = "db.Delete"

	d.log.Info("Delete request",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	query, args, err := d.bd.Delete(d.tableName).
		Where(sq.Eq{"latitude": lat}, sq.Eq{"longitude": lng}).
		ToSql()
	if err != nil {
		return fmt.Errorf("%s: insert to db: %w", op, err)
	}

	d.log.Info("Query created",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	if _, err := d.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("%s: delete from db: %w", op, err)
	}

	d.log.Info("Successfully deleted",
		zap.String("op", op),
		zap.String("reqTrace", reqTrace))

	return nil
}
