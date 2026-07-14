# Transition to Claude Code — Setup Guide

## Step 1: Install Claude Code

```bash
# Install Claude Code CLI
npm install -g @anthropic-ai/claude-code

# Or with yarn
yarn global add @anthropic-ai/claude-code

# Verify install
claude --version
```

## Step 2: Point Claude Code at Your Project

```bash
cd ~/trading-bot
claude
```

Or open VS Code with the Claude extension. Either way, Claude Code has access to your entire project directory.

## Step 3: Give Claude Code Context

Paste this summary as your first message to Claude:

```
You are continuing work on an automated crypto trading bot. Here's the state:

PROJECT: ~/trading-bot/
- Go execution engine + Python research service + Telegram bot
- Binance testnet paper trading (port 8420, SearXNG was using 8080)
- GitHub: https://github.com/alecrico19/trading-bot-engine

CURRENT CONFIG:
- Grid strategy: buys $25 BTC every 2 hours, max 3 positions
- Mean-reversion: trades ETH on RSI + BB (enabled)
- Scalping + tick-momentum: disabled (losing money for hours)
- Restart: ~/trading-bot/scripts/restart.sh
- Dashboard: http://localhost:8420
- Redis + research service run alongside engine

KEY FILES TO READ FIRST:
- config/config.yaml (all strategy/risk settings)
- execution/internal/engine/engine.go (main engine, executeDecision, checkStopLoss)
- execution/internal/engine/grid.go (DCA grid strategy, just built)
- execution/internal/engine/mean_reversion.go (RSI + BB strategy)
- docs/ (70+ markdown files with full history of every decision)

TOP PROBLEMS RIGHT NOW:
1. Grid buys 3 positions then stops (maxPositions cap)
2. Exits rarely fire in range-bound BTC (take-profit 0.15% too far)
3. Bot needs stable runtime without constant changes
4. 15+ hours of effort, never had sustained profitable run >30 min

WHAT I WAS ABOUT TO DO:
- Reduce grid to 2 positions at $35 each
- Add 2-hour stale exit to rotate old positions
- See docs/why-stops-at-3.md for the full plan

The Obsidian vault is at /mnt/e/Users/AlecR/Documents/HermesVault/TradingBot/
with all 72 markdown docs synced. Check there for context on every bug, fix, and decision.
```

## Step 4: Key Claude Code Commands

```
/claude-code start          Start coding session
/claude-code stop           End session
Shift+Enter in terminal     Submit multi-line messages
```

## Step 5: Project State At Handoff

- **Running engine:** `http://localhost:8420` (port 8420)
- **Restart:** `~/trading-bot/scripts/restart.sh`
- **View logs:** `tail -f /tmp/trading-bot-logs/engine.log`
- **Check P&L:** `curl -s http://localhost:8420/status`
- **Tailscale:** http://100.111.179.14:8420 (phone dashboard)
- **Telegram bot:** running with AUTHORIZED_USERS=1473027968
- **Redis:** running at localhost:6379
- **Python research:** `research/main.py` (runs in background)
- **Obsidian vault:** `/mnt/e/Users/AlecR/Documents/HermesVault/TradingBot/`
