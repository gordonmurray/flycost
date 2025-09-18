# flycost

**Cost estimation tool for Fly.io applications**

flycost analyzes your `fly.toml` configuration files and estimates monthly costs for compute, storage, bandwidth, and IP addresses. Similar to [infracost](https://github.com/infracost/infracost), but specifically for Fly.io.

## Quick Start

### Install (Recommended)

**macOS/Linux:**
```bash
curl -sSLf https://github.com/gordonmurray/flycost/releases/latest/download/install.sh | bash
```

**Manual download:**
```bash
# Replace VERSION with latest release (e.g., v1.0.0)
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m | sed 's/x86_64/amd64/; s/aarch64/arm64/')
wget "https://github.com/gordonmurray/flycost/releases/download/VERSION/flycost_${OS}_${ARCH}.tar.gz"
tar -xzf flycost_*.tar.gz
sudo mv flycost /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/gordonmurray/flycost.git
cd flycost
go build -o flycost ./cmd/flycost
```

## Usage

### Basic Usage

```bash
# Scan current directory for fly.toml files
flycost

# Scan specific directory
flycost -root ./my-apps

# Output as JSON
flycost -format json
```

### Example Output

```
Kind        Monthly  App           Note                     File
----------- -------  ------------- ------------------------ -----------
compute     $  1.94  user-service  shared-cpu-1x 1024MB     express-microservice.toml
volume      $  1.50  user-service  10 GB (assumed)          express-microservice.toml
bandwidth   $  2.00  user-service  lhr 100 GB               express-microservice.toml
ip          $  2.00  user-service  1 IPv4 (assumed)         express-microservice.toml
compute     $ 82.12  ml-inference-api performance-1x 16384MB   ml-worker.toml
volume      $  1.50  ml-inference-api 10 GB (assumed)          ml-worker.toml
bandwidth   $  2.00  ml-inference-api iad 100 GB               ml-worker.toml
ip          $  2.00  ml-inference-api 1 IPv4 (assumed)         ml-worker.toml
compute     $  3.88  my-rails-api  shared-cpu-2x 2048MB     rails-api.toml
compute     $  1.94  my-rails-api  shared-cpu-1x 1024MB     rails-api.toml
volume      $  1.50  my-rails-api  10 GB (assumed)          rails-api.toml
bandwidth   $  2.00  my-rails-api  cdg 100 GB               rails-api.toml
ip          $  2.00  my-rails-api  1 IPv4 (assumed)         rails-api.toml
compute     $  1.94  my-react-app  shared-cpu-1x 256MB      react-spa.toml
bandwidth   $  2.00  my-react-app  iad 100 GB               react-spa.toml
ip          $  2.00  my-react-app  1 IPv4 (assumed)         react-spa.toml

TOTAL: $112.32 / mo
```

## Example Applications

The `examples/` folder contains realistic fly.toml configurations:

- **`react-spa.toml`** - Static React app ($5.94/mo)
- **`express-microservice.toml`** - Node.js API with file storage ($7.44/mo)
- **`rails-api.toml`** - Multi-instance Rails API with volumes ($13.32/mo)
- **`ml-worker.toml`** - High-performance ML inference service ($87.62/mo)

## Configuration

Create `.flycost.yml` in your project root to customize cost assumptions:

```yaml
root: "."
assumptions:
  egress_gb_per_app: 100        # Monthly bandwidth usage
  ipv4_per_app: 1               # Number of IPv4 addresses
  default_volume_gb: 10         # Default volume size when mounts exist
  postgres_monthly_usd: 0       # Add-on Postgres cost
bandwidth_rates:
  default: 0.02                 # $/GB for unknown regions
  iad: 0.02                     # $/GB for US East
  cdg: 0.02                     # $/GB for Europe
```

## Supported fly.toml Features

- **Compute**: VM sizes, CPU types (shared/performance), memory
- **Storage**: Volume mounts (estimates size if not specified)
- **Networking**: Bandwidth based on primary region
- **IP Addresses**: IPv4 allocation costs

## Sample fly.toml Formats

flycost supports both standard TOML array syntax and Fly.io's newer single-section syntax:

**Array syntax (`[[vm]]`, `[[mounts]]`):**
```toml
app = "my-web-app"
primary_region = "iad"

[[vm]]
cpus = 2
cpu_kind = "shared"
memory_mb = 2048

[[mounts]]
source = "data_volume"
destination = "/data"
```

**Single section syntax (`[mounts]`):**
```toml
app = "user-service"
primary_region = "lhr"

[[vm]]
cpus = 1
cpu_kind = "shared"
memory_mb = 1024

[mounts]
source = "uploads"
destination = "/app/uploads"
```

## Pricing

Built-in pricing is based on Fly.io's public pricing:

### Compute (Base VM costs with 256MB RAM)
- **Shared CPU**: $1.94/mo per vCPU
- **Performance CPU**: $31/mo per vCPU
- **Memory**: $4.26/mo per GB above base allocation

### Infrastructure
- **Volumes**: $0.15/mo per GB
- **IPv4**: $2/mo per address
- **Bandwidth**:
  - North America/Europe: $0.02/mo per GB
  - Asia Pacific: $0.04/mo per GB

### Updating Pricing

Pricing data is stored in `pkg/pricing/pricing.json`. To update:

1. Check current rates at [fly.io/docs/about/pricing](https://fly.io/docs/about/pricing)
2. Edit `pkg/pricing/pricing.json`:
   ```json
   {
     "ram_per_gb_month": 4.26,
     "volume_gb_month": 0.15,
     "egress_gb_by_region": {
       "default": 0.02,
       "iad": 0.02,     // North America
       "cdg": 0.02,     // Europe
       "nrt": 0.04,     // Asia Pacific
       "syd": 0.04      // Oceania
     },
     "vm_preset_month": {
       "shared-cpu-1x": 1.94,
       "performance-1x": 31.0
     }
   }
   ```
3. Rebuild: `go build -o flycost ./cmd/flycost`

*Prices are estimates. Always verify with [fly.io/docs/about/pricing](https://fly.io/docs/about/pricing) for official rates.*

## Command Line Options

```
Usage: flycost [options]

Options:
  -root string
        Directory to scan for fly.toml files (default: current directory)
  -format string
        Output format: table|json (default: "table")
  -config string
        Path to custom .flycost.yml config file
```

## Integration

### CI/CD Usage

```yaml
# GitHub Actions example
- name: Estimate Fly.io costs
  run: |
    curl -sSLf https://github.com/gordonmurray/flycost/releases/latest/download/install.sh | bash
    flycost -format json > costs.json
```

### Shell Integration

```bash
# Add to your deployment script
echo "💰 Estimating costs..."
flycost
echo "Deploying to Fly.io..."
fly deploy
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request