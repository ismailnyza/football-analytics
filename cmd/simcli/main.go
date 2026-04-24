package main

import (
	"fmt"
	"os"

	"github.com/ismailnyza/football-analytics/engine/evidence"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "status" {
		printStatus()
		return
	}

	fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
	os.Exit(1)
}

func printStatus() {
	factor := evidence.FactorRecord{
		Name:       "historical team strength baseline",
		Hypothesis: "stronger teams should improve outcome prediction over naive priors",
		Grade:      evidence.GradeD,
		TestMethod: "planned backtest",
	}

	fmt.Println("self-improving football simulation research repo")
	fmt.Println("status: bootstrap complete, empirical baseline pending")
	fmt.Printf("next factor: %s (%s)\n", factor.Name, factor.Grade)
}
