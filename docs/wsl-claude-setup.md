# Claude Code CLI in WSL — Step by Step

## Step 1: Open WSL Terminal

Open your Windows terminal and start WSL:

```bash
wsl
```

Or open the "Ubuntu" or "WSL" app from your Start menu.

## Step 2: Verify Node.js (needed for npm)

```bash
node --version   # Should show v18+ or v20+
npm --version    # Should show v9+ or v10+
```

If Node.js isn't installed:
```bash
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs
```

## Step 3: Install Claude Code CLI

```bash
npm install -g @anthropic-ai/claude-code
```

This installs the `claude` command globally in WSL.

## Step 4: Authenticate

```bash
claude login
```

This opens a browser window to link your Claude Pro subscription. Follow the prompts.

## Step 5: Navigate to Your Project

```bash
cd ~/trading-bot
```

## Step 6: Start Claude Code

```bash
claude
```

Claude Code opens in your terminal, pointing at the `~/trading-bot/` directory. It can read and edit all your project files.

## Step 7: Give Claude Context (First Message)

Paste this entire block as your first message:

```
You are continuing work on an automated crypto trading bot. Here's the project state:

PROJECT: ~/trading-bot/
- Go execution engine + Python research service + Telegram bot
- Binance testnet paper trading (port 8420)
- GitHub: https://github.com/alecrico19/trading-bot-engine

CURRENT STATE:
- Grid strategy: buys $25 BTC every 2h, stops at 3 positions
- Mean-reversion: trades ETH on RSI + BB
- Scalping + tick-momentum: disabled (lost money)
- Restart: ~/trading-bot/scripts/restart.sh
- Dashboard: http://localhost:8420
- Running now? Check: curl -s http://localhost:8420/status

KEY FILES:
- config/config.yaml (all settings)
- execution/internal/engine/engine.go (main logic)
- execution/internal/engine/grid.go (DCA grid — just built)
- execution/internal/engine/mean_reversion.go (RSI+BB)

TOP PROBLEM:
Grid buys 3 positions then stops. Exits rarely fire in range-bound BTC.
Plan to fix: docs/why-stops-at-3.md

The Obsidian vault is at:
/mnt/e/Users/AlecR/Documents/HermesVault/TradingBot/
with 72 markdown docs covering every bug, fix, and decision.

Read the docs/ folder for full context before making changes.
```

## Step 8: Common Commands While Working

| Task | Command |
|------|---------|
| Check if bot is running | `curl -s http://localhost:8420/status` |
| Restart bot | `~/trading-bot/scripts/restart.sh` |
| View engine logs | `tail -f /tmp/trading-bot-logs/engine.log` |
| Check P&L | `curl -s http://localhost:8420/status \| python3 -c "import sys,json; d=json.load(sys.stdin); print(d['daily_pnl'])"` |
| View trades | `curl -s http://localhost:8420/trades` |
| Go build | `cd ~/trading-bot/execution && go build ./cmd/main.go` |
| Git push | `cd ~/trading-bot && git push` |
| Start Redis | `~/redis/redis-server --daemonize yes --port 6379` |

## Step 9: Exit Claude Code

```
exit
```

Or press `Ctrl+C` twice.
