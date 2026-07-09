# Trading Bot

> Autonomous day trading & scalping bot — Go execution engine + Python research service

## Architecture

- [[Trading Bot — Architecture Plan]] — Full architecture plan, tech stack, decisions
- [[SETUP]] — Setup and testing guide

## Services

| Service | Language | Purpose |
|---------|----------|---------|
| **Execution Engine** | Go | Strategies, risk management, order execution, 4 exchange adapters |
| **Research Service** | Python | News/sentiment, technical analysis, signal generation |
| **Telegram Bot** | Python | Remote control and monitoring |
| **Redis** | — | Signal pipeline between research and execution |
| **SQLite** | — | Trade journal, daily P&L snapshots |

## Quick Links

- **Live Dashboard:** `http://localhost:8080`
- **API:** `http://localhost:8080/status`
- **Source:** `~/trading-bot/`
- **GitHub:** https://github.com/alecrico19/trading-bot-engine

## Running

```bash
# Launcher (starts everything)
~/Desktop/start-trading-bot.sh

# Or individual services
./scripts/start.sh run       # TUI dashboard
./scripts/start.sh headless  # Headless + API on :8080
./scripts/start.sh research  # Research service
./scripts/start.sh telegram  # Telegram bot
./scripts/start.sh all       # All services
```

## Exchange Support

- Binance (testnet + live, WebSocket)
- Coinbase Pro (REST)
- Kraken (REST)
- Alpaca (stocks, REST)
