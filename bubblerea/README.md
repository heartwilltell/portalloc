# bubblerea

Colorful TUI to find busy ports and the owning processes on macOS and Linux, built with Bubble Tea.

- Refresh every 2s automatically
- Press `r` to refresh immediately
- Press `q` to quit
- Pass `--once` to print a single snapshot and exit (good for scripts/CI)

Build:

```bash
cd bubblerea
go mod tidy
go build -o bin/bubblerea
```

Run (interactive TUI):

```bash
./bin/bubblerea
```

Run one-shot (non-interactive):

```bash
./bin/bubblerea --once
```

Dependencies:
- Linux: uses `ss` from `iproute2` (preferred) or falls back to `lsof`.
  - Install: `sudo apt-get install iproute2 lsof` (Debian/Ubuntu) or `sudo yum install iproute` (RHEL/CentOS)
- macOS: uses `lsof` (usually preinstalled). If missing: `brew install lsof`.

Notes:
- You may need elevated privileges to see processes owned by other users (e.g., run with `sudo`).