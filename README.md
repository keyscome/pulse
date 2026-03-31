# Pulse

A lightweight, cross-platform CLI tool for monitoring network service connectivity. Pulse performs TCP connection checks against a configurable list of service endpoints and produces structured reports with per-service success and failure details.

## Features

- **TCP connectivity checks** — verifies reachability of any TCP endpoint within a configurable timeout
- **MinIO authenticated checks** — connects to MinIO with username/password credentials using the MinIO Go SDK
- **YAML-based configuration** — define services and their addresses in a simple `config.yml` file
- **Structured logging** — separate log files for successes, failures, and the full report, written to a timestamped directory under `logs/`
- **Templated reports** — customisable report format via `report.tpl` (Go `text/template`)
- **Cross-platform** — pre-built binaries for Linux (amd64/arm64), macOS (amd64/arm64), and Windows (amd64)

## Prerequisites

- Go **1.22** or later (only required when building from source)

## Installation

### Download a pre-built release

1. Go to the [Releases](../../releases) page and download the archive for your platform.
2. Extract the archive:

   ```bash
   # Linux / macOS
   tar -xzf pulse-<version>-<os>-<arch>.tar.gz -C /usr/local/pulse

   # Windows – extract the .zip file with your preferred tool
   ```

3. Add the binary to your `PATH` (Linux / macOS):

   ```bash
   export PULSE_HOME=/usr/local/pulse
   export PATH=$PATH:$PULSE_HOME
   ```

   Add these lines to `~/.bashrc` or `~/.zshrc` to make them permanent.

### Install via script (Linux / macOS)

```bash
bash scripts/install.sh
```

The script downloads the latest release, extracts it to `/usr/local/pulse`, and configures the necessary environment variables.

## Configuration

Pulse reads a `config.yml` file from the **current working directory**. The file has the following sections:

- **`services`** — a map of service names to lists of `host:port` addresses for TCP connectivity checks
- **`elasticsearch`** — Elasticsearch-specific configuration with credentials for authenticated HTTP health checks
- **`kibana`** — Kibana-specific configuration with credentials for authenticated HTTP status checks
- **`redis`** — Redis-specific configuration with optional password for protocol-level checks
- **`minio`** — MinIO-specific configuration with credentials and addresses for authenticated connection checks

```yaml
# config.yml – example
services:
  web:
    - 10.0.31.131:30310
  nacos:
    - 10.0.31.131:30848
  kafka:
    - 10.0.1.30:9092
  zookeeper:
    - 10.0.1.27:2181,10.0.1.28:2181,10.0.1.29:2181
  zk-ui:
    - 10.0.1.27:9090

elasticsearch:
  addresses:
    - 10.0.1.24:9200
    - 10.0.1.25:9200
    - 10.0.1.26:9200
  username: elastic
  password: changeme

kibana:
  addresses:
    - 10.0.1.26:5601
  username: elastic
  password: changeme

redis:
  password: ""
  addresses:
    - 10.0.1.38:6379
    - 10.0.1.38:6380

minio:
  username: minioadmin
  password: minioadmin
  addresses:
    - 10.0.1.35:9000
```

**Elasticsearch** connects via the HTTP API (`/_cluster/health`) and supports Basic Auth. `username` and `password` are optional — omit them for clusters with security disabled.

Add or remove services under `services` as needed — any service name is accepted. The `elasticsearch`, `kibana`, `redis`, and `minio` sections are optional; omit them if not needed.

## Usage

Run Pulse from the directory that contains `config.yml` and `report.tpl`:

```bash
./pulse
```

Pulse will:

1. Load `config.yml`.
2. Attempt a TCP connection to every address under `services` (3-second timeout per address).
3. Check the Elasticsearch cluster health via HTTP API using optional Basic Auth credentials.
4. Attempt an authenticated connection to every MinIO address using the provided `username` and `password`.
5. Print a report to standard output.
6. Write detailed logs to `logs/<timestamp>/`.

### Example output

```
===== 检测报告 =====
Network Connection Report - 2025-03-06 11:00:11

Service: redis
  Success:
    - 10.0.1.38:6379
  Failures:
    - 10.0.1.38:6380

Service: kafka
  Success:
    - 10.0.1.30:9092
  Failures:
    (无失败连接)
...
```

## Logs

Each run creates a timestamped directory under `logs/`:

```
logs/
└── 20250306110011/
    ├── success.log   # addresses that connected successfully
    ├── failure.log   # addresses that failed to connect
    └── report.log    # full formatted report
```

## Local Simulation Environment

The `docker/` directory contains a **Docker Compose** setup that starts all the
services Pulse monitors, so you can try the tool against a real local stack
without an external cluster.

### Services started

| Service | Image | Exposed port(s) |
|---------|-------|-----------------|
| nacos | `nacos/nacos-server:v2.3.2` | 8848 (HTTP), 9848 (gRPC) |
| redis | `redis:7-alpine` | 6379 |
| zookeeper | `bitnami/zookeeper:3.9` | 2181 |
| zk-ui | `elkozmon/zoonavigator:latest` | 9090 |
| kafka | `bitnami/kafka:3.7` | 9092 |
| elasticsearch | `elasticsearch:7.17.21` | 9200 (HTTP), 9300 (transport) |
| kibana | `kibana:7.17.21` | 5601 |
| minio | `minio/minio:RELEASE.2024-04-06T05-26-02Z` | 9000 (S3 API), 9001 (console) |

### Quick start

```bash
# 1. Start all services
cd docker
docker compose up -d

# 2. Wait for services to become healthy (typically 60-90 seconds)
docker compose ps

# 3. Run Pulse against the local stack
cd ..
cp docker/config.yml config.yml
./pulse
```

### Web UIs

| UI | URL | Default credentials |
|----|-----|---------------------|
| Nacos | <http://localhost:8848/nacos> | nacos / nacos |
| MinIO console | <http://localhost:9001> | minioadmin / minioadmin |
| Kibana | <http://localhost:5601> | elastic / changeme |
| ZooKeeper Navigator | <http://localhost:9090> | — |

### Tear down

```bash
cd docker
docker compose down
```

---

## Building from Source

Clone the repository and build with the standard Go toolchain:

```bash
go build -o pulse .
```

### Cross-platform builds

Use the provided build scripts in the `build/` directory, or set the environment variables manually:

| Platform        | Command |
|-----------------|---------|
| Linux amd64     | `GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -o pulse` |
| Linux arm64     | `GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -o pulse` |
| macOS amd64     | `GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -o pulse` |
| macOS arm64     | `GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -o pulse` |
| Windows amd64   | `GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o pulse.exe` |

## Project Structure

```
pulse/
├── build/          # Platform-specific build scripts
├── checker/        # TCP and Elasticsearch connection checker package
├── config/         # YAML configuration loader package
├── logger/         # Structured logger (success / failure / report)
├── scripts/        # Installation script
├── config.yml      # Example service configuration
├── go.mod          # Go module definition
├── main.go         # Main entry point
└── report.tpl      # Report template (Go text/template)
```

## License

This project does not currently include a license file. Please contact the repository owner for usage terms.
