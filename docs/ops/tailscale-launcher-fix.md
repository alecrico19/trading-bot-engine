# Tailscale Launcher Fix

## Root Cause
The launcher starts tailscaled and checks `kill -0 $PID` after 2 seconds. But tailscaled takes ~3-4 seconds to initialize. The PID exists but the check is inconclusive at 2s.

## Fix: Wait for socket file instead of PID check

Replace:
```bash
$TAILSCALED_BIN ... &
TAILSCALED_PID=$!
sleep 2
if kill -0 $TAILSCALED_PID 2>/dev/null; then
```

With:
```bash
$TAILSCALED_BIN ... &
TAILSCALED_PID=$!
# Wait for socket file to appear (up to 10s)
for i in $(seq 1 20); do
    if [ -S "$TAILSCALE_SOCK" ]; then
        break
    fi
    sleep 0.5
done
if [ -S "$TAILSCALE_SOCK" ]; then
```

## File
`scripts/launcher.sh` — ~8 lines changed