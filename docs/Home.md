# Trading Bot

> Autonomous day trading & scalping bot — Go execution engine + Python research service

## Architecture

- [[trading-bot-plan]] — Full architecture plan, tech stack, decisions
- [[SETUP]] — Setup and testing guide
- [[code-audit]] — Full audit results (11 critical, 17 high, 12 medium bugs fixed)

## Active Plans & Fixes

- [[aggressive-tuning]] — Lower scalping thresholds, more positions
- [[equity-chart-plan]] — Live P&L equity curve
- [[improvements-plan]] — Stop-loss, trailing stop, strategy P&L, trade history
- [[research-integration]] — 200 EMA trend filter + take-profit from pro scalper research
- [[strategy-assessment]] — Two strategies: enough or add more?
- [[next-steps]] — Overnight run → review → live trading
- [[remote-access-plan]] — Tailscale for phone dashboard access

## Bug Fixes Applied

- [[pnl-fix-plan]] — P&L using entry price not ticker
- [[pnl-bug-fix]] — Paper trader phantom money
- [[positions-fix]] — Holdings invisible after buys
- [[zero-trade-fix]] — Rejected orders in trade history
- [[ws-reconnect-plan]] — WebSocket auto-reconnect
- [[systemd-fix-plan]] — Auto-start on boot
- [[backtesting-plan]] — Walk-forward backtesting engine
- [[gold-trading-plan]] — Gold (XAU/USD) via forex brokers

## Dashboard

- [[dashboard-fix-plan]] — Total portfolio equity + holdings-based positions

## Services

| Service | Language | Purpose |
|---------|----------|---------|
| **Execution Engine** | Go | Strategies, risk management, order execution, 4 exchange adapters |
| **Research Service** | Python | News/sentiment, technical analysis, signal generation |
| **Telegram Bot** | Python | Remote control and monitoring |
| **Redis** | — | Signal pipeline between research and execution |
| **SQLite** | — | Trade journal, daily P&L snapshots |

## Quick Links

- **Live Dashboard:** `http://localhost:8420`
- **API:** `http://localhost:8420/status`
- **Source:** `~/trading-bot/`
- **GitHub:** https://github.com/alecrico19/trading-bot-engine

## Running

```bash
# Launcher (starts everything)
~/Desktop/start-trading-bot.sh

# Or individual services
./scripts/start.sh run       # TUI dashboard
./scripts/start.sh headless  # Headless + API on :8420
./scripts/start.sh research  # Research service
./scripts/start.sh telegram  # Telegram bot
./scripts/start.sh backtest  # Backtest optimizer
```

## Exchange Support

- Binance (testnet + live, WebSocket with auto-reconnect)
- Coinbase Pro (REST)
- Kraken (REST)
- Alpaca (stocks, REST)
