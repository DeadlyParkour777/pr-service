package store

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testStore *Store

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("test-db"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %s", err)
	}
	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate postgres container: %s", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %s", err)
	}

	migrationsPath, _ := filepath.Abs("../../migrations")
	migrator, err := migrate.New(fmt.Sprintf("file://%s", migrationsPath), connStr)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %s", err)
	}
	if err := migrator.Up(); err != nil {
		log.Fatalf("failed to run migrations: %s", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to create connection pool: %s", err)
	}
	testStore = &Store{conn: pool}

	exitCode := m.Run()

	os.Exit(exitCode)
}

func truncateTables(ctx context.Context) {
	_, err := testStore.conn.Exec(ctx, `TRUNCATE teams, users, pull_requests, pull_request_reviewers RESTART IDENTITY CASCADE;`)
	if err != nil {
		log.Fatalf("failed to truncate tables: %v", err)
	}
}
