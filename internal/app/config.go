package app

// Config holds process-wide bootstrap configuration.
type Config struct {
	Name    string
	Version string
}

// DefaultConfig returns the baseline runtime metadata for local builds.
func DefaultConfig() Config {
	return Config{
		Name:    "Football Simulation Engine",
		Version: "0.1.0-bootstrap",
	}
}
