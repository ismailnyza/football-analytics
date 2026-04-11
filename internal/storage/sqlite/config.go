package sqlite

import "path/filepath"

const (
	// DefaultFileName is the baseline SQLite database name for local worlds.
	DefaultFileName = "football.db"
)

// DefaultPath returns the default SQLite file path rooted under a provided data directory.
func DefaultPath(dataDir string) string {
	return filepath.Join(dataDir, DefaultFileName)
}
