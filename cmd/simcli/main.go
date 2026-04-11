package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ismael/football-analytics/internal/app"
	"github.com/ismael/football-analytics/internal/ingestion"
)

func main() {
	cfg := app.DefaultConfig()
	if len(os.Args) >= 2 && os.Args[1] == "scrape" {
		fs := flag.NewFlagSet("scrape", flag.ExitOnError)
		stateDir := fs.String("state-dir", "", "state directory (default: user cache dir)")
		pageURL := fs.String("url", "", "FBref page URL (optional if "+ingestion.FbrefURLFile+" exists in state dir)")
		fs.Usage = func() {
			fmt.Fprintf(fs.Output(), "Usage: %s scrape [flags]\n\n", os.Args[0])
			fmt.Fprintln(fs.Output(), "Fetches one FBref HTML page, parses stats_table player rows, stages records, updates ingest ledger.")
			fmt.Fprintln(fs.Output(), "Respect site terms and rate limits; pass a single squad or stats page URL.")
			fmt.Fprintln(fs.Output())
			fs.PrintDefaults()
		}
		fs.Parse(os.Args[2:])
		dir := *stateDir
		if dir == "" {
			var err error
			dir, err = cfg.ResolveStateDir()
			if err != nil {
				fmt.Fprintf(os.Stderr, "state dir: %v\n", err)
				os.Exit(1)
			}
		}
		u := strings.TrimSpace(*pageURL)
		if u == "" {
			var err error
			u, err = ingestion.ReadFbrefURLFile(dir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "scrape: set --url or write %s in state dir: %v\n", ingestion.FbrefURLFile, err)
				os.Exit(1)
			}
		}
		if err := ingestion.RunFbrefScrape(context.Background(), dir, u, time.Now()); err != nil {
			fmt.Fprintf(os.Stderr, "scrape: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("state dir: %s\n", dir)
		fmt.Printf("ingest ledger: %s\n", filepath.Join(dir, ingestion.IngestLedgerFile))
		return
	}

	if len(os.Args) >= 2 && os.Args[1] == "fetch" {
		fs := flag.NewFlagSet("fetch", flag.ExitOnError)
		stateDir := fs.String("state-dir", "", "override state directory (default: user cache dir)")
		fs.Usage = func() {
			fmt.Fprintf(fs.Output(), "Usage: %s fetch [flags]\n\n", os.Args[0])
			fmt.Fprintln(fs.Output(), "Runs demo source adapters, applies per-source staging caps, updates ingest_state.json.")
			fs.PrintDefaults()
		}
		fs.Parse(os.Args[2:])
		dir := *stateDir
		if dir == "" {
			var err error
			dir, err = cfg.ResolveStateDir()
			if err != nil {
				fmt.Fprintf(os.Stderr, "state dir: %v\n", err)
				os.Exit(1)
			}
		}
		if err := ingestion.RunDemoFetch(context.Background(), dir, time.Now()); err != nil {
			fmt.Fprintf(os.Stderr, "fetch: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("state dir: %s\n", dir)
		fmt.Printf("ingest ledger: %s\n", filepath.Join(dir, ingestion.IngestLedgerFile))
		return
	}

	showVersion := flag.Bool("version", false, "print version information")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n\n", cfg.Name)
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n  %s [--version]\n  %s fetch [--state-dir DIR]\n  %s scrape [--url U] [--state-dir DIR]\n\n", os.Args[0], os.Args[0], os.Args[0])
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
