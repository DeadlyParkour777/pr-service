package handler

import (
	"context"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"path/filepath"

	"testing"
	"time"

	"github.com/DeadlyParkour777/pr-service/internal/service"
	"github.com/DeadlyParkour777/pr-service/internal/store"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testServerURL string
	testStore     *store.Store
)

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
	defer func() { _ = pgContainer.Terminate(ctx) }()

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

	appStore, err := store.NewStore(connStr)
	if err != nil {
		log.Fatalf("failed to create store: %s", err)
	}
	testStore = appStore

	deps := service.Dependencies{
		TeamRepo: appStore.Team(),
		UserRepo: appStore.User(),
		PRRepo:   appStore.PR(),
	}
	appService := service.NewService(deps)
	appHandler := NewHandler(appService)
	router := appHandler.InitRoutes()

	server := httptest.NewServer(router)
	defer server.Close()
	testServerURL = server.URL

	exitCode := m.Run()

	os.Exit(exitCode)
}

func truncateTables(ctx context.Context) {
	if err := testStore.TruncateAllTables(ctx); err != nil {
		log.Fatalf("failed to truncate tables: %v", err)
	}
}
