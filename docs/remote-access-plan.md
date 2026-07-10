# Remote Dashboard Access (Phone)

## Approach: Tailscale (Recommended)

Free, secure VPN. Both devices get a private IP — phone accesses `http://100.x.x.x:8080` directly.

### Setup
1. Install Tailscale on your PC: `curl -fsSL https://tailscale.com/install.sh | sh`
2. Install Tailscale on your phone (App Store / Play Store)
3. Both devices auto-join the same Tailscale network
4. Bookmark `http://[PC-tailscale-IP]:8080` on your phone

### Code change needed
HTTP API currently binds to `127.0.0.1` only (local). Needs `--public` flag to bind `0.0.0.0:8080` for Tailscale access.

- `cmd/main.go`: add `--public` flag, change bind address
- ~3 lines

## Alternative: ngrok

One command, instant. No install on phone.

```bash
ngrok http 8080
```

Gives you a URL like `https://abc123.ngrok.io` you can open on your phone browser. Free tier limitations: random URL on restart, bandwidth cap.

## Security Note
Either way, the HTTP API has no auth (we already flagged this). For remote access, enabling basic auth or the API token would be wise. Can add `X-API-Token` header check in ~10 lines.
