package loader

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"

	"github.com/gordonmurray/flycost/internal/types"
	"github.com/gordonmurray/flycost/pkg/pricing"
)

func LoadPrices() (types.PriceTable, error) {
	var pt types.PriceTable
	if err := json.Unmarshal(pricing.Raw, &pt); err != nil {
		return pt, err
	}
	return pt, nil
}

func LoadConfig(start string) (types.Config, error) {
	// walk up to find .flycost.yml
	cur := start
	for {
		cfg := filepath.Join(cur, ".flycost.yml")
		if _, err := os.Stat(cfg); err == nil {
			var c types.Config
			b, err := os.ReadFile(cfg)
			if err != nil {
				return c, err
			}
			if err := yaml.Unmarshal(b, &c); err != nil {
				return c, err
			}
			if c.Root == "" {
				c.Root = cur
			}
			return c, nil
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	// no config found → sensible defaults
	return types.Config{
		Root: ".",
	}, nil
}

func ParseFlyToml(path string) (types.FlyToml, error) {
	var ft types.FlyToml
	b, err := os.ReadFile(path)
	if err != nil {
		return ft, err
	}

	// Parse into a generic structure first to handle different mount formats
	var raw map[string]interface{}
	if err := toml.Unmarshal(b, &raw); err != nil {
		return ft, err
	}

	// Extract basic fields
	if app, ok := raw["app"].(string); ok {
		ft.App = app
	}
	if region, ok := raw["primary_region"].(string); ok {
		ft.PrimaryRegion = region
	}

	// Handle VMs
	if vmsData, ok := raw["vm"]; ok {
		if vmSlice, isSlice := vmsData.([]interface{}); isSlice {
			for _, vmData := range vmSlice {
				if vmMap, isMap := vmData.(map[string]interface{}); isMap {
					vm := types.FlyVM{}
					if cpus, ok := vmMap["cpus"].(int64); ok {
						vm.CPUs = int(cpus)
					}
					if kind, ok := vmMap["cpu_kind"].(string); ok {
						vm.CPUKind = kind
					}
					if mem, ok := vmMap["memory_mb"].(int64); ok {
						vm.MemoryMB = int(mem)
					}
					ft.VMs = append(ft.VMs, vm)
				}
			}
		}
	}

	// Handle mounts - both [mounts] and [[mounts]]
	if mountsData, ok := raw["mounts"]; ok {
		if mountMap, isMap := mountsData.(map[string]interface{}); isMap {
			// Single [mounts] section
			if source, hasSource := mountMap["source"].(string); hasSource {
				if dest, hasDest := mountMap["destination"].(string); hasDest {
					ft.Mounts = []types.FlyMount{{
						Source:      source,
						Destination: dest,
					}}
				}
			}
		} else if mountSlice, isSlice := mountsData.([]interface{}); isSlice {
			// Multiple [[mounts]] sections
			for _, mountData := range mountSlice {
				if mountMap, isMap := mountData.(map[string]interface{}); isMap {
					if source, hasSource := mountMap["source"].(string); hasSource {
						if dest, hasDest := mountMap["destination"].(string); hasDest {
							ft.Mounts = append(ft.Mounts, types.FlyMount{
								Source:      source,
								Destination: dest,
							})
						}
					}
				}
			}
		}
	}

	if strings.TrimSpace(ft.App) == "" {
		return ft, errors.New("missing app name")
	}
	return ft, nil
}
