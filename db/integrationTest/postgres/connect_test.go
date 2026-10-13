package integrationtest

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/ooqls/getset/db/containers"
	"github.com/ooqls/getset/db/pgx"
	"github.com/ooqls/getset/db/postgres"
	"github.com/ooqls/getset/db/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Initialize the Redis container
	redisContainer := containers.StartRedis(context.Background())
	defer func() {
		if err := redisContainer.Terminate(context.Background()); err != nil {
			log.Fatalf("failed to terminate redis container: %v", err)
		}
	}()

	// Initialize the Postgres container
	postgresContainer := containers.StartPostgres(context.Background(), containers.WithLogging())
	defer func() {
		if err := postgresContainer.Terminate(context.Background()); err != nil {
			log.Fatalf("failed to terminate postgres container: %v", err)
		}
	}()

	os.Exit(m.Run())
}

func TestConnectPostgres(t *testing.T) {
	// This is a placeholder test to ensure that the test suite runs correctly.
	// You can add your actual test logic here.
	t.Log("Running integration tests...")

	assert.Nilf(t, sqlx.InitDefault(), "Expected InitDefault to return nil, but got an error.")

	assert.Nilf(t, pgx.InitDefault(), "Expected InitDefault to return nil, but got an error.")
}

func TestInitKeepsTheConnectionItOpens(t *testing.T) {
	opts := postgres.GetRegistryOptions()

	pool, err := pgx.Init(context.Background(), opts)
	require.NoError(t, err)
	assert.Same(t, pool, pgx.GetPGX(), "GetPGX should return the pool Init opened")
	assert.NoError(t, pgx.GetPGX().Ping(context.Background()))

	db, err := sqlx.Init(opts)
	require.NoError(t, err)
	assert.Same(t, db, sqlx.GetSQLX(), "GetSQLX should return the db Init opened")
	assert.NoError(t, sqlx.GetSQLX().Ping())
}
