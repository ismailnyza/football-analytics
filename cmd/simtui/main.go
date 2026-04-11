package main

import (
	"fmt"

	"github.com/ismael/football-analytics/internal/app"
	"github.com/ismael/football-analytics/internal/tui/layout"
)

func main() {
	cfg := app.DefaultConfig()
	fmt.Print(layout.RenderBootstrapShell(cfg))
}
