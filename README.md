# cronwatch

A daemon that monitors cron job execution times and alerts on drift or missed runs via webhook.

---

## Installation

```bash
go install github.com/yourname/cronwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourname/cronwatch.git && cd cronwatch && go build -o cronwatch .
```

---

## Usage

Create a config file (`cronwatch.yaml`) defining the jobs to monitor:

```yaml
webhook: "https://hooks.example.com/alerts"

jobs:
  - name: "nightly-backup"
    schedule: "0 2 * * *"
    timeout: 30m
    drift_threshold: 5m

  - name: "hourly-sync"
    schedule: "0 * * * *"
    timeout: 5m
    drift_threshold: 1m
```

Start the daemon:

```bash
cronwatch --config cronwatch.yaml
```

cronwatch will watch for heartbeats from your cron jobs and fire a webhook alert if a job runs late, finishes beyond its timeout, or is missed entirely.

To send a heartbeat from your cron script, simply call:

```bash
curl -s "http://localhost:8080/heartbeat?job=nightly-backup"
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `cronwatch.yaml` | Path to config file |
| `--port` | `8080` | Port for the heartbeat HTTP listener |
| `--log-level` | `info` | Log verbosity (`debug`, `info`, `warn`, `error`) |

---

## License

MIT © yourname