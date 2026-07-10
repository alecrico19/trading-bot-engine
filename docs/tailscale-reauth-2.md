# Tailscale Re-Auth — Quick Fix

## Step 1: Kill the old daemon
```bash
pkill tailscaled
```

## Step 2: Reset state
```bash
rm -f ~/.tailscale/tailscaled.state
```

## Step 3: Start fresh + get auth URL
Run these in one terminal (keep it open):
```bash
~/bin/tailscaled --tun=userspace-networking --socket=$HOME/.tailscale/tailscaled.sock --state=$HOME/.tailscale/tailscaled.state &
sleep 3
~/bin/tailscale --socket=$HOME/.tailscale/tailscaled.sock up --accept-routes
```

It will print a URL like `https://login.tailscale.com/a/abc123...`

## Step 4: Authenticate
Open that URL in a browser. Log in with Google/GitHub/Microsoft.

## Step 5: Verify
```bash
~/bin/tailscale --socket=$HOME/.tailscale/tailscaled.sock ip
```
Should show `100.x.x.x`.

## Step 6: Restart launcher
```bash
~/Desktop/start-trading-bot.sh
```

Tailscale now auto-starts with the launcher and stays authenticated.
