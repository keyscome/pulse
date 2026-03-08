# Pulse

A lightweight, cross-platform CLI tool for monitoring network service connectivity. Pulse performs TCP connection checks against a configurable list of service endpoints and produces structured reports with per-service success and failure details.

## Features

- **TCP connectivity checks** — verifies reachability of any TCP endpoint within a configurable timeout
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

Pulse reads a `config.yml` file from the **current working directory**. Each top-level key is a service name, and its value is a list of `host:port` addresses to check.

```yaml
# config.yml – example
web:
  - 10.0.31.131:30310
nacos:
  - 10.0.31.131:30848
redis:
  - 10.0.1.38:6379
  - 10.0.1.38:6380
kafka:
  - 10.0.1.30:9092
elasticsearch:
  - 10.0.1.24:9300
kibana:
  - 10.0.1.26:5601
minio:
  - 10.0.1.35:9000
zookeeper:
  - 10.0.1.27:3000
```

Add or remove services as needed — any service name is accepted.

## Usage

Run Pulse from the directory that contains `config.yml` and `report.tpl`:

```bash
./pulse
```

Pulse will:

1. Load `config.yml`.
2. Attempt a TCP connection to every address (3-second timeout per address).
3. Print a report to standard output.
4. Write detailed logs to `logs/<timestamp>/`.

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
├── checker/        # TCP connection checker package
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
