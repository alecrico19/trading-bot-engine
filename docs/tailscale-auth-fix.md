# Tailscale Auth Fix

## Status: Already Authenticated ✓

- IP: 100.112.254.73
- State file: ~/.tailscale/tailscaled.state (persistent)
- iPhone connected: 100.123.219.2

## Issue: Misleading launcher message

The launcher runs `tailscale ip` too early — the daemon hasn't finished connecting yet when the check runs. Fix: add a retry loop or wait longer.

## Fix (launcher.sh)

```bash
# Try up to 5 times with 2s delay
for i in $(seq 1 5); do
    TS_IP=$($TAILSCALE_BIN --socket="$TAILSCALE_SOCK" ip 2>/dev/null | head -1)
    [ -n "$TS_IP" ] && break
    sleep 2
done
```

This gives the daemon up to 10 seconds to connect before showing "(needs auth)".

## Device already connected: YES
Your phone (iPhone XS Max) is on the Tailscale network. Open `http://100.112.254.73:8080` in your phone browser to see the dashboard.