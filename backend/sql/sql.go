package sql

import (
	"context"
	"database/sql"
	"fmt"
	g "github.com/Thooms/gatekeeper"
	"github.com/jmoiron/sqlx"
)

type SQLBackend struct {
	db        *sqlx.DB
	tableName string
}

func FromxDB(db *sqlx.DB, tableName string) *SQLBackend {
	return &SQLBackend{db: db, tableName: tableName}
}

type entry struct {
	ApiKey string `db:"api_key""`
	Limit  int64  `db:"api_limit"`
	Usage  int64  `db:"api_usage"`
}

func (b *SQLBackend) Allow(ctx context.Context, k g.Key) (bool, g.Stats, error) {
	tx, err := b.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, g.Stats{}, fmt.Errorf("unable to begin transaction: %v", err)
	}
	rollbackOnError := func(e error) error {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%v (unable to rollback: %v)", e, rbErr)
		}
		return e
	}
	res := &entry{}
	if err := tx.QueryRowxContext(ctx, b.rebindf(`SELECT * FROM %s WHERE api_key = ? LIMIT 1`, b.tableName), k).StructScan(res); err != nil {
		if err == sql.ErrNoRows {
			return false, g.Stats{}, rollbackOnError(g.ErrUnknownKey)
		}
		return false, g.Stats{}, rollbackOnError(fmt.Errorf("unable to get data for API key '%v': %v", k, err))
	}

	if res.Usage >= res.Limit {
		return false, g.Stats{Remaining: 0, Limit: res.Limit}, nil
	}

	if _, err := tx.ExecContext(ctx, b.rebindf(`UPDATE %v SET api_usage = api_usage + 1 WHERE api_key = ?`, b.tableName), k); err != nil {
		return false, g.Stats{}, rollbackOnError(fmt.Errorf("unable to update API usage for key '%v': %v", k, err))
	}
	if err := tx.Commit(); err != nil {
		return false, g.Stats{}, rollbackOnError(fmt.Errorf("unable to commit transaction: %v", err))
	}
	// -1 because at this point we successfully did `api_usage++`
	return true, g.Stats{Remaining: res.Limit - (res.Usage + 1), Limit: res.Limit}, nil
}

func (b *SQLBackend) Stats(ctx context.Context, k g.Key) (g.Stats, error) {
	res := &entry{}
	if err := b.db.QueryRowxContext(ctx, b.rebindf(`SELECT * FROM %s WHERE api_key = ? LIMIT 1`, b.tableName), k).StructScan(res); err != nil {
		if err == sql.ErrNoRows {
			return g.Stats{}, g.ErrUnknownKey
		}
		return g.Stats{}, fmt.Errorf("unable to get data for API key '%v': %v", k, err)
	}
	return g.Stats{Remaining: res.Limit - res.Usage, Limit: res.Limit}, nil
}

func (b *SQLBackend) createTableStmt() string {
	return fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
  api_key TEXT PRIMARY KEY,
  api_limit INT,
  api_usage INT
)
`, b.tableName)
}

// Ensures the database is accessible and have the right schema.
func (b *SQLBackend) Check(ctx context.Context) error {
	if err := b.db.PingContext(ctx); err != nil {
		return fmt.Errorf("unable to access database: %v", err)
	}
	if _, err := b.db.ExecContext(ctx, b.createTableStmt()); err != nil {
		return fmt.Errorf("unable to create table: %v", err)
	}
	return nil
}

func (b *SQLBackend) rebindf(format string, a ...interface{}) string {
	return b.db.Rebind(fmt.Sprintf(format, a...))
}
