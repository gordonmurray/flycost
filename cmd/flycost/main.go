package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/gordonmurray/flycost/internal/estimate"
	"github.com/gordonmurray/flycost/internal/loader"
	"github.com/gordonmurray/flycost/internal/types"
)

func main() {
	// minimal flags; config handles most things
	cfgPath := flag.String("config", "", "path to .flycost.yml (optional)")
	format := flag.String("format", "table", "table|json|both")
	rootArg := flag.String("root", "", "override scan root (optional)")
	flag.Parse()

	wd, _ := os.Getwd()
	cfg, _ := loader.LoadConfig(wd)
	if *cfgPath != "" {
		// simple override: if provided, load that file directly
		b, err := os.ReadFile(*cfgPath)
		if err == nil {
			_ = yaml.Unmarshal(b, &cfg)
		}
	}
	if *rootArg != "" {
		cfg.Root = *rootArg
	}

	prices, err := loader.LoadPrices()
	if err != nil {
		die(err)
	}

	var items []types.LineItem
	total := 0.0

	err = filepath.WalkDir(cfg.Root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := d.Name()
		if name == "fly.toml" || filepath.Ext(name) == ".toml" {
			ft, err := loader.ParseFlyToml(path)
			if err != nil {
				return nil
			}
			t, its := estimate.ForFile(path, ft, cfg, prices)
			total += t
			items = append(items, its...)
		}
		return nil
	})
	if err != nil {
		die(err)
	}

	switch *format {
	case "json":
		out := map[string]any{"total_monthly_usd": total, "items": items}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
	default:
		printTable(items, total)
	}
}

func printTable(items []types.LineItem, total float64) {
	fmt.Printf("Kind        Monthly  App           Note                     File\n")
	fmt.Printf("----------- -------  ------------- ------------------------ -----------\n")
	for _, it := range items {
		fmt.Printf("%-10s  $%6.2f  %-13s %-24s %s\n", it.Kind, it.Monthly, it.App, it.Note, it.File)
	}
	fmt.Printf("\nTOTAL: $%.2f / mo\n", total)
}

func die(err error) { fmt.Fprintln(os.Stderr, err); os.Exit(1) }
