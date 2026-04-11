package app

import (
	"os"
	"path/filepath"
)

func (c Config) ResolveStateDir() (string, error) {
	if c.StateDir != "" {
		return c.StateDir, nil
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "football-analytics"), nil
}
