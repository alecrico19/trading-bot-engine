# Remaining Steps

## 1. Complete Tailscale Auth (5 min — you do this)
```bash
~/bin/tailscaled --tun=userspace-networking --socket=$HOME/.tailscale/tailscaled.sock &
~/bin/tailscale --socket=$HOME/.tailscale/tailscaled.sock up
~/bin/tailscale --socket=$HOME/.tailscale/tailscaled.sock ip
```
Visit the auth URL, log in. Once you have an IP (e.g., `100.x.x.x`), open `http://100.x.x.x:8420` on your phone. Also install Tailscale on your phone with the same account.

## 2. Restart Launcher (2 min)
The engine needs to pick up all audit fixes. Stop the current launcher (Ctrl+C), then:
```bash
~/Desktop/start-trading-bot.sh
```

## 3. Overnight Paper Run (8+ hours)
Let it run. Check periodically:
- **Dashboard** — equity curve, positions, P&L
- **Telegram** — `/status` from your phone
- **Trade journal** — `curl http://localhost:8420/trades | python3 -m json.tool`
- **Crash test** — does it survive 8 hours with no intervention?

Expected behavior:
- Scalping generates buy+sell pairs rapidly
- Mean reversion fires every few minutes
- Stop-loss triggers if a position drops 5%
- Equity stays near $1000 (minus fees)
- No crashes, no panics, no DB corruption

## 4. Review + Tune
After overnight run, check if:
- Scalping profit factor > 1.0 (more wins than losses)
- Mean reversion actually fires
- Stop-loss ever triggered (if so, check if thresholds are right)

Apply optimized params from backtest (`./scripts/start.sh backtest BTC/USDT 7`) to `config.yaml`.

## 5. Small Live Trading
When ready:
```yaml
live_trading: true
max_order_size_usd: 20
```
Supervise first hour. Scale up after a week of positive P&L.
