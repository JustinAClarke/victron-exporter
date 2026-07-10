# AGENTS

This repository implements a small Go-based exporter for Victron MQTT telemetry.
The exporter connects to a Victron MQTT broker, subscribes to device state topics,
translates those topics into Prometheus metrics, and exposes `/metrics` over HTTP.

## Project structure

- `main.go` - application entrypoint that starts the HTTP server, MQTT connections,
  and periodic polling.
- `env.go` - helpers for reading environment variables and falling back to defaults.
- `log.go` - helper for mapping a numeric logging configuration to `logrus` levels.
- `metrics.go` - Prometheus metric registration for exporter health and MQTT status.
- `mqtt.go` - MQTT connection management, TLS configuration, and subscription handling.
- `pem.go` - bundled Victron root certificate for secure MQTT connections.
- `topics.go` - mapping of Victron MQTT topic suffixes to Prometheus metric observers.
- `README.md` - user-facing documentation for installation, configuration, and usage.

## Building

This project uses Go modules. The repository expects a Go toolchain to be installed.

1. Ensure dependencies are available:
   - `go mod tidy`
2. Build the binary:
   - `go build -o victron-exporter .`

The resulting binary is `victron-exporter`.

## Running

The exporter is configured using command-line flags or environment variables.
Common environment variables include:

- `LISTEN_ADDR` - the HTTP listen address for Prometheus metrics (default `127.0.0.1:9226`).
- `MQTT_HOST` - the Victron MQTT broker host.
- `MQTT_PORT` - broker port, default `8883`.
- `MQTT_SECURE` - whether to use TLS, default `true`.
- `MQTT_USERNAME` / `MQTT_PASSWORD` - credentials for the MQTT broker.
- `MQTT_CLIENT_PREFIX` - prefix for the MQTT client IDs.
- `LOG_LEVEL` - numeric log level, `0=debug`, `1=info`, `2=warn`, `3=error`.
- `VICTRON_POLL_INTERVAL` - polling interval duration (e.g. `10s`).

## Testing

There are no Go unit tests in the repository at the moment. Validate the project by
ensuring the code compiles and formatting is correct:

- `gofmt -w *.go`
- `go test ./...`

## Code conventions

- Keep documentation comments concise and explanatory. Prefer describing why the
  code exists, not just what it does.
- Use `prometheus.MustRegister` only during metric construction, and avoid
  re-registering the same metric multiple times.
- Keep MQTT message handling resilient: ignore unsupported topics, log parse errors,
  and avoid panics from malformed payloads.
- Prefer explicit configuration handling in `env.go` helpers. Invalid configuration
  values should be fatal to prevent undefined behaviour.

## Contribution guidance for agents

When modifying this project, consider the following:

- If adding new metrics, add mappings in `topics.go` and register them via the
  existing observer helpers.
- If changing MQTT connection behaviour, update `mqtt.go` and ensure connection
  state metrics remain accurate.
- Keep `main.go` lightweight: orchestrate setup and let helper packages manage
  their respective concerns.
- Preserve existing behavior for Victron topic handling unless there is a clear
  backwards-compatible improvement.

If you need more context, review `README.md` for user-level assumptions and the
existing MQTT topic mapping in `topics.go`.
