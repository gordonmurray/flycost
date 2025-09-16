package estimate

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gordonmurray/flycost/internal/types"
)

func presetKey(vm types.FlyVM) string {
	if strings.HasPrefix(vm.CPUKind, "shared") {
		if vm.CPUs <= 1 {
			return "shared-cpu-1x"
		}
		return fmt.Sprintf("shared-cpu-%dx", vm.CPUs)
	}
	// very rough default; refine as you map Fly SKUs
	return "performance-1x"
}

func ForFile(path string, ft types.FlyToml, cfg types.Config, pr types.PriceTable) (float64, []types.LineItem) {
	var items []types.LineItem
	total := 0.0

	// Compute
	for _, vm := range ft.VMs {
		pk := presetKey(vm)
		base := pr.Presets[pk]
		// naive RAM uplift: assume base includes 1024 MB per vCPU
		extraMB := vm.MemoryMB - (1024 * vm.CPUs)
		if extraMB < 0 {
			extraMB = 0
		}
		ram := (float64(extraMB) / 1024.0) * pr.RAMPerGBMo
		cost := base + ram
		items = append(items, types.LineItem{
			File: filepath.Base(path), App: ft.App, Kind: "compute",
			Note: fmt.Sprintf("%s %dMB", pk, vm.MemoryMB), Monthly: cost,
		})
		total += cost
	}

	// Volumes (if mounts exist but no size known, use default)
	if len(ft.Mounts) > 0 && cfg.Assumptions.DefaultVolumeGB > 0 {
		gb := cfg.Assumptions.DefaultVolumeGB
		cost := float64(gb) * pr.VolGBMo
		items = append(items, types.LineItem{
			File: filepath.Base(path), App: ft.App, Kind: "volume",
			Note: fmt.Sprintf("%d GB (assumed)", gb), Monthly: cost,
		})
		total += cost
	}

	// Bandwidth
	if cfg.Assumptions.EgressGBPerApp > 0 {
		r := cfg.BandwidthRates[strings.ToLower(ft.PrimaryRegion)]
		if r == 0 {
			r = pr.EgressGB["default"]
		}
		if r == 0 {
			r = 0.02
		}
		cost := r * cfg.Assumptions.EgressGBPerApp
		items = append(items, types.LineItem{
			File: filepath.Base(path), App: ft.App, Kind: "bandwidth",
			Note:    fmt.Sprintf("%s %.0f GB", ft.PrimaryRegion, cfg.Assumptions.EgressGBPerApp),
			Monthly: cost,
		})
		total += cost
	}

	// IPv4 (optional future: price table)
	if cfg.Assumptions.IPv4PerApp > 0 {
		// placeholder at $2.0 / mo per IPv4 (adjust when you map exact)
		cost := float64(cfg.Assumptions.IPv4PerApp) * 2.0
		items = append(items, types.LineItem{
			File: filepath.Base(path), App: ft.App, Kind: "ip",
			Note: fmt.Sprintf("%d IPv4 (assumed)", cfg.Assumptions.IPv4PerApp), Monthly: cost,
		})
		total += cost
	}

	// Postgres add-on
	if cfg.Assumptions.PostgresMonthlyUSD > 0 {
		items = append(items, types.LineItem{
			File: filepath.Base(path), App: ft.App, Kind: "postgres",
			Note: "assumed plan", Monthly: cfg.Assumptions.PostgresMonthlyUSD,
		})
		total += cfg.Assumptions.PostgresMonthlyUSD
	}

	return total, items
}
