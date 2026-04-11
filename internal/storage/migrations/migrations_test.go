package migrations

import (
	"strings"
	"testing"
)

func TestLoadReturnsSortedEmbeddedMigrations(t *testing.T) {
	migrations, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(migrations) == 0 {
		t.Fatal("Load() returned no migrations")
	}

	if got, want := migrations[0].Name, "0001_initial_schema.sql"; got != want {
		t.Fatalf("first migration = %q, want %q", got, want)
	}

	for i := 1; i < len(migrations); i++ {
		if migrations[i-1].Name >= migrations[i].Name {
			t.Fatalf("migrations not sorted: %q before %q", migrations[i-1].Name, migrations[i].Name)
		}
	}
}

func TestInitialSchemaContainsCoreTables(t *testing.T) {
	migrations, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	schema := migrations[0].SQL
	required := []string{
		"CREATE TABLE IF NOT EXISTS schema_migrations",
		"CREATE TABLE IF NOT EXISTS worlds",
		"CREATE TABLE IF NOT EXISTS branches",
		"CREATE TABLE IF NOT EXISTS clubs",
		"CREATE TABLE IF NOT EXISTS players",
		"CREATE TABLE IF NOT EXISTS seasons",
		"CREATE TABLE IF NOT EXISTS fixtures",
		"CREATE TABLE IF NOT EXISTS matches",
		"CREATE TABLE IF NOT EXISTS lineup_entries",
		"CREATE TABLE IF NOT EXISTS match_events",
		"CREATE TABLE IF NOT EXISTS raw_payloads",
	}

	for _, fragment := range required {
		if !strings.Contains(schema, fragment) {
			t.Fatalf("schema missing fragment %q", fragment)
		}
	}
}
