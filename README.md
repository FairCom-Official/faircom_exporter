# FairCom Prometheus Exporter

Prometheus exporter for FairCom database that collects comprehensive performance metrics using the `ctstat` utility.

## Features

- **Comprehensive Metrics**: 58 metrics covering all aspects of FairCom server operation
- **ctstat Integration**: Uses native FairCom ctstat utility for reliable data collection
- **Zero CGO Dependencies**: Pure Go implementation with subprocess calls
- **Configurable Collectors**: Enable/disable specific metric groups
- **Structured Logging**: JSON or text format with automatic log rotation
- **Production Ready**: TLS support, basic auth, systemd integration
- **Real-time Data**: Direct access to live server statistics
- **Multiple Report Types**: Aggregates data from 6 different ctstat reports

## Metrics Collected

### Cache Statistics (4 metrics)
- Data cache hit/miss percentages
- Index cache hit/miss percentages

### I/O Statistics (2 metrics)
- Read operations per second
- Write operations per second

### Transaction Metrics (11 metrics)
- Transaction begins/commits/aborts
- Savepoints and restores
- Transaction log writes and bytes
- Active transactions and throughput
- Read/write transaction timing

### Lock Statistics (4 metrics)
- Locks currently held
- Lock hit/miss percentages
- Deadlock count

### Connection Metrics (2 metrics)
- Active users (current/max)

### File Metrics (9 metrics)
- Files open (current/max)
- File operations (opens/closes/creates/deletes/renames)
- Physical read/write operations

### ISAM Operations (8 metrics)
- Record adds/deletes/updates/reads
- First/last/next/previous operations

### SQL Operations (6 metrics)
- SELECT/INSERT/UPDATE/DELETE statements
- COMMIT/ROLLBACK operations

### User Statistics (10 metrics)
- Total memory usage across all users
- Active vs idle user counts
- Aggregated I/O operations and bytes
- Data and index buffer requests/hits

## Quick Start

### Building from Source

```bash
# Build for Linux ARM64
GOOS=linux GOARCH=arm64 go build -o faircom_exporter ./cmd/exporter

# Build for Linux AMD64
GOOS=linux GOARCH=amd64 go build -o faircom_exporter ./cmd/exporter

# Or use Makefile
make build-linux-arm64
make build-linux-amd64
```

### Basic Configuration

Create a `config.yaml` file:

```yaml
server:
  port: 9100
  path: /metrics

faircom:
  username: ADMIN
  password: ADMIN
```

### Running

```bash
# Direct execution
./faircom_exporter --config config.yaml

# Test metrics endpoint
curl http://localhost:9100/metrics | grep "^faircom_"
```

## Installation

### Using Package Managers

**Debian/Ubuntu (DEB):**
```bash
# Install package
sudo dpkg -i faircom_exporter-1.0.0-amd64.deb   # or arm64

# Create faircom user (if needed)
sudo useradd -r -s /bin/false faircom

# Start service
sudo systemctl start faircom_exporter
sudo systemctl enable faircom_exporter

# Check status
sudo systemctl status faircom_exporter
```

**RHEL/CentOS/Fedora (RPM):**
```bash
# Install package
sudo rpm -i faircom_exporter-1.0.0-amd64.rpm    # or aarch64

# Create faircom user (if needed)
sudo useradd -r -s /bin/false faircom

# Start service
sudo systemctl start faircom_exporter
sudo systemctl enable faircom_exporter

# Check status
sudo systemctl status faircom_exporter
```

**Manual Installation (tar.gz):**
```bash
# Extract
tar -xzf faircom_exporter-1.0.0-linux-amd64.tar.gz
cd faircom_exporter-1.0.0-linux-amd64

# Copy files
sudo cp faircom_exporter /usr/local/bin/
sudo mkdir -p /etc/faircom_exporter
sudo cp config.yaml /etc/faircom_exporter/

# Create log directory
sudo mkdir -p /var/log/faircom_exporter
sudo chown faircom:faircom /var/log/faircom_exporter

# Copy systemd service (optional)
sudo cp ../systemd/faircom_exporter.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable faircom_exporter
sudo systemctl start faircom_exporter
```

## Building Packages

```bash

# Build individually
make package-tar    # Creates tar.gz archives
make package-deb    # Creates Debian packages
make package-rpm    # Creates RPM packages
```

Packages are created in the `output/` directory:
- `output/tar/` - Compressed archives with binary, config, and README
- `output/deb/` - Debian packages with systemd integration
- `output/rpm/` - RPM packages with systemd integration

```cmd.exe (Windows)
# Windows builds require both visual studio and gcc (mingw)
# 1. Install gcc from the instructions at:  https://www.mingw-w64.org/getting-started/msys2/
# 2. Check the FAIRCOMDB_DIR and UCRT64 values in Makefile-windows and update as needed.
# 3. From visual studio x64 command environment, run makewin.bat

## Configuration

### Basic Configuration

Minimal configuration for getting started:

```yaml
server:
  port: 9100
  path: /metrics

faircom:
  username: ADMIN
  password: ADMIN
```

### Full Configuration

Complete configuration with all available options:

```yaml
# HTTP server settings
server:
  port: 9100                    # Port to listen on
  path: /metrics                # Metrics endpoint path
  
  # TLS configuration (optional, recommended for production)
  # tls:
  #   enabled: true
  #   cert_file: /etc/faircom_exporter/tls/server.crt
  #   key_file: /etc/faircom_exporter/tls/server.key
  
  # Basic authentication for metrics endpoint (optional)
  # basic_auth:
  #   username: prometheus
  #   password: changeme
  
# FairCom database connection
faircom:
  # Basic authentication for faircom server
  basic_auth:
    username: ADMIN               # FairCom admin username
    password: ADMIN               # FairCom admin password
  # 
  # tls:
      enabled: true
      ca_file: /etc/faircom/ca.crt
  # Client certificate authentication for faircom server
  #   cert_file: /home/user/faircomuser.crt/
  #   key_file: /home/user/.keys/faircomuser.key
  


# Collector toggles (all enabled by default)
collectors:
  cache: true                   # Cache hit/miss metrics (4 metrics)
  transactions: true            # Transaction metrics (11 metrics)
  io: true                      # IO metrics
  locks: true                   # Lock contention metrics (4 metrics)
  files: true                   # File operation metrics (9 metrics)
  isam: true                    # ISAM operation metrics (8 metrics)
  users: true                   # User statistics (11 metrics)

# Logging configuration
log:
  level: info                   # Log level: debug, info, warn, error
  format: text                  # Format: text or json
  file: /var/log/faircom_exporter/faircom_exporter.log
  max_size: 100                 # Max size in MB before rotation (default: 100)
  max_backups: 3                # Number of old log files to keep (default: 3)
  max_age: 28                   # Days to retain old log files (default: 28)
  compress: true                # Compress rotated files (default: true)
```

### Production Configuration

For production deployments with TLS and authentication:

```yaml
server:
  port: 9100
  path: /metrics
  # Enable HTTPS
  tls:
    enabled: true
    cert_file: /etc/faircom_exporter/tls/server.crt
    key_file: /etc/faircom_exporter/tls/server.key
  # Protect metrics endpoint
  basic_auth:
    username: prometheus
    password: your-secure-password

faircom:
  # Enable TLS client authentication
  tls: 
    enabled: true
    cert_file: /home/user/user.crt
    key_file: /home/user/.keys/user.key
    ca_file: /etc/faircom/keys/ca.crt

log:
  level: info
  format: json                  # Use JSON for structured logging
  file: /var/log/faircom_exporter/faircom_exporter.log
  max_size: 100
  max_backups: 5                # Keep more backups in production
  max_age: 30
  compress: true
```

### Collector Configuration

You can disable specific collectors to reduce metrics volume:

```yaml
collectors:
  cache: true           # Keep cache metrics
  transactions: true    # Keep transaction metrics
  locks: true          # Keep lock metrics
  files: false         # Disable file metrics
  isam: false          # Disable ISAM metrics
  users: true          # Keep user statistics
```

See [config.example.yaml](config.example.yaml) for a complete annotated example.

## Requirements

- **FairCom Server**: FairCom database
- **libmtclient.so**: FairCom client library must be accessible 
- **Go**: 1.21+ for building from source
- **Faircom SDK** for building from source
- **MSYS2 UCRT64** gcc environment for building cgo from source on windows
- **Credentials**: FairCom admin username and password
- **Log Directory**: Write access to `/var/log/faircom_exporter/` (auto-created by systemd)

## Architecture

The exporter uses a cgo interface and wrapper to the C FairComDB SnapShot API call.


## Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'faircom'
    static_configs:
      - targets: ['localhost:9100']
    scrape_interval: 15s
```

## Systemd Service

The DEB and RPM packages include a systemd service that:
- Runs as user `faircom` (you may need to create this user)
- Reads configuration from `/etc/faircom_exporter/config.yaml`
- Automatically creates log directory `/var/log/faircom_exporter/`
- Logs with automatic rotation (configurable max size, backups, age)
- Automatically restarts on failure
- Includes security hardening (NoNewPrivileges, ProtectSystem, ProtectHome)

**Service Management:**
```bash
# Start/stop
sudo systemctl start faircom_exporter
sudo systemctl stop faircom_exporter

# Enable/disable auto-start
sudo systemctl enable faircom_exporter
sudo systemctl disable faircom_exporter

# Check status
sudo systemctl status faircom_exporter

# View logs (from systemd journal)
sudo journalctl -u faircom_exporter -f

# View logs (from log file)
sudo tail -f /var/log/faircom_exporter/faircom_exporter.log

# View rotated logs
sudo ls -lh /var/log/faircom_exporter/
```

## File Locations

After package installation:
- **Binary**: `/usr/local/bin/faircom_exporter`
- **Configuration**: `/etc/faircom_exporter/config.yaml`
- **Log files**: `/var/log/faircom_exporter/faircom_exporter.log` (with rotation)
- **Systemd unit**: `/usr/lib/systemd/system/faircom_exporter.service` (RPM) or `/lib/systemd/system/faircom_exporter.service` (DEB)

## Docker Deployment

```bash
# Copy binary and config to container
docker cp faircom_exporter faircom-edge:/opt/faircom/
docker cp config.yaml faircom-edge:/opt/faircom/

# Create log directory
docker exec faircom-edge mkdir -p /var/log/faircom_exporter

# Start exporter in background
docker exec -d faircom-edge sh -c "cd /opt/faircom && ./faircom_exporter --config config.yaml"

# Verify metrics
curl http://localhost:9100/metrics | grep "^faircom_"

# View logs
docker exec faircom-edge tail -f /var/log/faircom_exporter/faircom_exporter.log
```

## Metric Examples

```prometheus
# Cache performance
faircom_cache_hit_percent{type="data"} 98
faircom_cache_miss_percent{type="data"} 2

# Transaction throughput
faircom_transactions_per_sec 150
faircom_transaction_commits_total 1234567

# Lock contention
faircom_lock_hit_percent 100
faircom_deadlocks_total 0

# SQL operations
faircom_sql_selects_total 45678
faircom_sql_inserts_total 12345

# User statistics
faircom_total_memory_kb 4096
faircom_total_active_users 5
faircom_total_read_bytes 1048576
```

## Development

### Project Structure
```
.
├── cmd/exporter/          # Application entry point
├── pkg/
│   ├── collector/         # Metrics collection logic
│   └── config/           # Configuration handling
├── systemd/              # Systemd service file
├── output/               # Build artifacts (gitignored)
│   ├── deb/
│   ├── rpm/
│   └── tar/
├── config.yaml           # Example configuration
├── config.example.yaml   # Complete annotated example
├── Makefile             # Build automation
└── README.md
```

### Building

```bash
# Clean build artifacts
make clean

# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Format code
make fmt

# Run all checks
make check
```

## Troubleshooting

**ctstat not found:**
- Ensure ctstat binary is available at the configured path (default: `/opt/faircom/ctstat`)
- Update `faircom.ctstat_path` in config.yaml if installed elsewhere
- Check executable permissions: `chmod +x /opt/faircom/ctstat`
- Verify FairCom SDK is properly installed

**Connection refused:**
- Check FairCom server is running
- Verify username/password in config.yaml
- Ensure ctstat can connect: `/usr/bin/faircom/ctstat -vas -u ADMIN -p ADMIN -i 1 1`

**No metrics appearing:**
- Check exporter logs: `tail -f /var/log/faircom_exporter/faircom_exporter.log`
- Or check systemd journal: `journalctl -u faircom_exporter -f`
- Verify port 9100 is not blocked by firewall
- Test endpoint: `curl http://localhost:9100/metrics`

**Permission errors:**
- Ensure exporter has execute permissions
- For systemd: create `faircom` user: `sudo useradd -r -s /bin/false faircom`
- Check log directory permissions: `ls -ld /var/log/faircom_exporter`

**Log rotation not working:**
- Check log file path in config.yaml
- Verify write permissions to log directory
- Check log settings: max_size, max_backups, max_age in config
- Old logs are compressed as `.gz` files by default

**Collector errors:**
- Enable debug logging: set `log.level: debug` in config.yaml
- Check which collectors are enabled in `collectors` section
- Disable problematic collectors individually

## License

MIT
