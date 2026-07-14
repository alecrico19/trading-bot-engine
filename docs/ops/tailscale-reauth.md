# Tailscale Re-Authentication

## Current State
- Daemon has IP: 100.112.254.73 (authenticated)
- iPhone listed: 100.123.219.2 (connected)
- State file exists at ~/.tailscale/tailscaled.state

## If You Need to Re-Auth the PC Daemon

```bash
# Stop the daemon
pkill tailscaled

# Remove old state
rm ~/.tailscale/tailscaled.state

# Start fresh
~/bin/tailscaled --tun=userspace-networking --socket=$HOME/.tailscale/tailscaled.sock --state=$HOME/.tailscale/tailscaled.state &
sleep 3

# Auth (will print URL to visit)
~/bin/tailscale --socket=$HOME/.tailscale/tailscaled.sock up --accept-rules

# Visit the URL in your browser, log in
```

## If iPhone Shows Offline

1. Open Tailscale app on iPhone
2. Sign in with same account (alecricohermoso@gmail.com)
3. Toggle the connection switch on
4. Both devices should show "connected" within 10 seconds

## Verify
```bash
~/bin/tailscale --socket=$HOME/.tailscale/tailscaled.sock status
```

Both devices should show `active` or `idle`, not `offline`.
