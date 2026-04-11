package migrations

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

//go:embed *.sql
var files embed.FS

// Migration is a single ordered schema change.
type Migration struct {
	Name string
	SQL  string
}

// Load returns embedded SQL migrations sorted by filename.
func Load() ([]Migration, error) {
	entries, err := fs.ReadDir(files, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".sql" {
			continue
		}
		names = append(names, name)
	}

	slices.Sort(names)
	migrations := make([]Migration, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		if _, ok := seen[name]; ok {
			return nil, fmt.Errorf("duplicate migration name %q", name)
		}
		seen[name] = struct{}{}

		body, err := files.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", name, err)
		}

		sql := strings.TrimSpace(string(body))
		if sql == "" {
			return nil, fmt.Errorf("migration %q is empty", name)
		}

		migrations = append(migrations, Migration{
			Name: name,
			SQL:  sql,
		})
	}

	return migrations, nil
}
