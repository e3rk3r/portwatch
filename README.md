# portwatch

A lightweight CLI daemon that monitors port availability and triggers configurable webhooks or scripts on state changes.

---

## Installation

```bash
go install github.com/yourusername/portwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/portwatch.git && cd portwatch && go build -o portwatch .
```

---

## Usage

```bash
portwatch --port 8080 --interval 10s --webhook https://hooks.example.com/notify
```

Watch multiple ports using a config file:

```bash
portwatch --config portwatch.yaml
```

Example `portwatch.yaml`:

```yaml
interval: 15s
ports:
  - port: 8080
    on_up: "curl -X POST https://hooks.example.com/up"
    on_down: "/usr/local/bin/restart-service.sh"
  - port: 5432
    on_down: "curl -X POST https://hooks.example.com/db-down"
```

When a port transitions from open to closed (or vice versa), portwatch executes the configured webhook URL or shell script for that event.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | — | Single port to monitor |
| `--interval` | `10s` | Poll interval |
| `--webhook` | — | Webhook URL to call on state change |
| `--config` | — | Path to YAML config file |
| `--verbose` | `false` | Enable verbose logging |
| `--timeout` | `5s` | Dial timeout per port check |

---

## Environment Variables

All flags can also be set via environment variables using the `PORTWATCH_` prefix:

| Variable | Equivalent Flag |
|----------|-----------------|
| `PORTWATCH_CONFIG` | `--config` |
| `PORTWATCH_INTERVAL` | `--interval` |
| `PORTWATCH_WEBHOOK` | `--webhook` |
| `PORTWATCH_VERBOSE` | `--verbose` |

---

## License

MIT © 2024 yourusername
