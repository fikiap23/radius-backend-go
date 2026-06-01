//go:build ignore

package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"

	atlas "ariga.io/atlas/sql/migrate"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/lib/pq"

	"github.com/radius/radius-backend/ent/migrate"
)

const (
	migrationsDir      = "migrations"
	citextExtensionSQL = "CREATE EXTENSION IF NOT EXISTS citext"
	citextPreamble     = "-- Enable citext for case-insensitive email.\n" + citextExtensionSQL + ";\n\n"
)

func main() {
	ctx := context.Background()

	devURL := os.Getenv("ATLAS_DEV_URL")
	if devURL == "" {
		log.Fatal("ATLAS_DEV_URL is required")
	}

	if err := ensurePostgresExtensions(ctx, devURL); err != nil {
		log.Fatalf("failed ensuring postgres extensions: %v", err)
	}

	dir, err := atlas.NewLocalDir(migrationsDir)
	if err != nil {
		log.Fatalf("failed creating atlas migration directory: %v", err)
	}

	opts := []schema.MigrateOption{
		schema.WithDir(dir),
		schema.WithMigrationMode(schema.ModeReplay),
		schema.WithDialect(dialect.Postgres),
		schema.WithFormatter(atlas.DefaultFormatter),
		migrate.WithDropColumn(true),
		migrate.WithDropIndex(true),
	}

	if len(os.Args) != 2 {
		log.Fatalln("migration name is required. Use: 'go run -mod=mod ent/migrate/diff/main.go <name>'")
	}

	before, err := migrationSQLFiles(migrationsDir)
	if err != nil {
		log.Fatalf("failed listing migrations: %v", err)
	}

	if err := migrate.NamedDiff(ctx, devURL, os.Args[1], opts...); err != nil {
		log.Fatalf("failed generating migration file: %v", err)
	}

	after, err := migrationSQLFiles(migrationsDir)
	if err != nil {
		log.Fatalf("failed listing migrations: %v", err)
	}
	if len(after) == len(before) {
		log.Println("warning: no new migration file — Ent schema matches replayed migrations (no diff)")
	}

	if err := prependCitextToLatestMigration(migrationsDir); err != nil {
		log.Fatalf("failed patching migration for citext: %v", err)
	}
}

func ensurePostgresExtensions(ctx context.Context, devURL string) error {
	db, err := sql.Open("postgres", devURL)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.ExecContext(ctx, citextExtensionSQL)
	return err
}

// Ent does not emit CREATE EXTENSION; prepend it to the newest migration when missing.
func prependCitextToLatestMigration(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var latest string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		if latest == "" || e.Name() > latest {
			latest = e.Name()
		}
	}
	if latest == "" {
		return nil
	}

	path := filepath.Join(dir, latest)
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.Contains(string(content), citextExtensionSQL) {
		return nil
	}
	// Only the initial migration creates tables; later diffs must not get CREATE EXTENSION.
	if !strings.Contains(string(content), `CREATE TABLE "users"`) {
		return nil
	}

	return os.WriteFile(path, append([]byte(citextPreamble), content...), 0o644)
}

func migrationSQLFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	return files, nil
}
