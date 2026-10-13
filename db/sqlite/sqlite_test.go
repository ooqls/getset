package sqlite

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tempDB(t *testing.T) string {
	return filepath.Join(t.TempDir(), "test.db")
}

func Test_GetSQLite(t *testing.T) {
	require.NoError(t, Init("get", tempDB(t), []string{"CREATE TABLE things (id INTEGER)"}))

	db, err := Get("get")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO things (id) VALUES (1)")
	require.NoError(t, err)

	var count int
	require.NoError(t, MustGet("get").QueryRow("SELECT COUNT(*) FROM things").Scan(&count))
	assert.Equal(t, 1, count)
}

func Test_GetReturnsTheSameDBEveryTime(t *testing.T) {
	require.NoError(t, Init("same", tempDB(t), nil))

	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			assert.Same(t, MustGet("same"), MustGet("same"))
		})
	}
	wg.Wait()
}

func Test_GetUnknownName(t *testing.T) {
	_, err := Get("never-initialized")
	assert.ErrorContains(t, err, "never-initialized")
	assert.Panics(t, func() { MustGet("never-initialized") })
}

func Test_InitAgainReplacesAndClosesTheOldDB(t *testing.T) {
	require.NoError(t, Init("replace", tempDB(t), nil))
	old := MustGet("replace")

	require.NoError(t, Init("replace", tempDB(t), []string{"CREATE TABLE fresh (id INTEGER)"}))

	assert.NotSame(t, old, MustGet("replace"))
	assert.Error(t, old.Ping(), "the replaced DB should be closed")
	_, err := MustGet("replace").Exec("INSERT INTO fresh (id) VALUES (1)")
	assert.NoError(t, err)
}

func Test_InitWithBadSchemaKeepsTheOldDB(t *testing.T) {
	require.NoError(t, Init("bad-schema", tempDB(t), nil))
	old := MustGet("bad-schema")

	err := Init("bad-schema", tempDB(t), []string{"NOT SQL"})

	assert.ErrorContains(t, err, "bad-schema")
	assert.Same(t, old, MustGet("bad-schema"))
	assert.NoError(t, old.Ping())
}
