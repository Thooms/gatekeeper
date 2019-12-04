package sql_test

import (
	"context"
	g "github.com/Thooms/gatekeeper"
	"github.com/Thooms/gatekeeper/backend/sql"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheck(t *testing.T) {
	o, _ := sqlx.Open("sqlite3", ":memory:")
	b := sql.FromxDB(o, "keys")
	if err := b.Check(context.Background()); err != nil {
		t.Error(err)
	}
}

func TestAllow(t *testing.T) {
	o, _ := sqlx.Open("sqlite3", ":memory:")
	b := sql.FromxDB(o, "keys")
	b.Check(context.Background())

	// insert data

	o.MustExec(`INSERT INTO keys VALUES ("abcd", 2, 0)`)

	// simulate 3 calls

	ok, stats, err := b.Allow(context.Background(), "abcd")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, g.Stats{Remaining: 1, Limit: 2}, stats)

	ok, stats, err = b.Allow(context.Background(), "abcd")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, g.Stats{Remaining: 0, Limit: 2}, stats)

	ok, stats, err = b.Allow(context.Background(), "abcd")
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, g.Stats{Remaining: 0, Limit: 2}, stats)

	// unknown key

	ok, stats, err = b.Allow(context.Background(), "unknownKey")
	assert.Error(t, err, "")
	assert.False(t, ok)
	assert.Equal(t, g.Stats{}, stats)
}

func TestStats(t *testing.T) {
	o, _ := sqlx.Open("sqlite3", ":memory:")
	b := sql.FromxDB(o, "keys")
	b.Check(context.Background())

	// insert data

	o.MustExec(`INSERT INTO keys VALUES ("abcd", 2, 1)`)

	stats, err := b.Stats(context.Background(), "abcd")
	assert.NoError(t, err)
	assert.Equal(t, g.Stats{Remaining: 1, Limit: 2}, stats)

	stats, err = b.Stats(context.Background(), "unknownKey")
	assert.Error(t, err, "")
	assert.Equal(t, g.Stats{}, stats)
}
