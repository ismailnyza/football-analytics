package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ismael/football-analytics/internal/app"
)

func main() {
	cfg := app.DefaultConfig()
	showVersion := flag.Bool("version", false, "print version information")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n\n", cfg.Name)
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n  %s [--version]\n\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Bootstrap CLI for the football simulation engine.")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Printf("%s %s\n", cfg.Name, cfg.Version)
		return
	}

	flag.Usage()
}
