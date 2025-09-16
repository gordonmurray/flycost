package types

type PriceTable struct {
	RAMPerGBMo float64            `json:"ram_per_gb_month"`
	VolGBMo    float64            `json:"volume_gb_month"`
	EgressGB   map[string]float64 `json:"egress_gb_by_region"`
	Presets    map[string]float64 `json:"vm_preset_month"`
}

type Config struct {
	Root        string `yaml:"root"`
	Assumptions struct {
		EgressGBPerApp     float64 `yaml:"egress_gb_per_app"`
		IPv4PerApp         int     `yaml:"ipv4_per_app"`
		DefaultVolumeGB    int     `yaml:"default_volume_gb"`
		PostgresMonthlyUSD float64 `yaml:"postgres_monthly_usd"`
	} `yaml:"assumptions"`
	BandwidthRates map[string]float64 `yaml:"bandwidth_rates"`
}

type FlyVM struct {
	CPUs     int    `toml:"cpus"`
	CPUKind  string `toml:"cpu_kind"`
	MemoryMB int    `toml:"memory_mb"`
}

type FlyMount struct {
	Source      string `toml:"source"`
	Destination string `toml:"destination"`
}

type FlyToml struct {
	App           string     `toml:"app"`
	PrimaryRegion string     `toml:"primary_region"`
	VMs           []FlyVM    `toml:"vm"`     // [[vm]]
	Mounts        []FlyMount `toml:"mounts"` // [[mounts]]
}

type LineItem struct {
	File    string  `json:"file"`
	App     string  `json:"app"`
	Kind    string  `json:"kind"`
	Note    string  `json:"note"`
	Monthly float64 `json:"monthly"`
}
