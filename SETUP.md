# Trading Bot — Setup & Testing Plan

## Prerequisites

| Tool | Check | Install |
|------|-------|---------|
| Go 1.22+ | `go version` | Already installed at `/home/alecr/go` |
| Python 3.11+ | `python3 --version` | Already installed |
| Redis | `redis-cli ping` | `docker compose up -d` (Docker) or `apt install redis` |
| Docker | `docker ps` | Only needed for Redis |

## Step 1: Get Binance Testnet API Keys

1. Go to https://testnet.binance.vision
2. Log in with GitHub
3. Generate HMAC API key + secret
4. Set environment variables:

```bash
export BINANCE_API_KEY="your_testnet_api_key"
export BINANCE_API_SECRET="your_testnet_api_secret"
```

The config already points to `${BINANCE_API_KEY}` and `${BINANCE_API_SECRET}` via env var substitution.

## Step 2: Start Redis

```bash
# Option A: Docker
cd ~/trading-bot
docker compose up -d

# Option B: Local install
redis-server --daemonize yes
```

Verify: `redis-cli ping` → should return `PONG`

## Step 3: First Paper Run (Engine Only)

Start the engine in paper mode with TUI. No Redis needed for this step.

```bash
cd ~/trading-bot
./scripts/start.sh run
```

**Expected:** TUI dashboard opens showing:
- Status: `● PAPER`
- Equity: `$1,000.00`
- Strategies: `scalping ▶ RUNNING`, `mean-reversion ▶ RUNNING`
- Order book panel: `(WebSocket feed active)`

**Watch for:**
- Order book WebSocket connects (may take a few seconds for `BTCUSDT` and `ETHUSDT`)
- Positions panel: should remain `No active positions` since risk manager blocks orders without market data
- Any log errors about API connectivity

**Key bindings in the TUI:**
- `q` — quit
- `p` — pause all strategies
- `s` — resume all strategies
- `k` — kill switch (pauses + cancels open orders + trips breaker)

## Step 4: Add Research Service

Open second terminal:

```bash
cd ~/trading-bot
./scripts/start.sh research
```

**Expected:** Research service starts polling news and OHLCV data:
```
[INFO] research: rss: 12 articles
[INFO] research: reddit: 15 articles
[INFO] research: sentiment: combined=0.45 label=positive (from 27 articles)
[INFO] publisher: signal: {"id":"sig_abc123","direction":"long",...}
```

**If Redis is running:** Signals flow to the engine. Look for them in the TUI's "SIGNALS" panel.

**If Redis is NOT running:** Signals are logged to stderr only. The engine runs without them.

## Step 5: Check Signal Flow End-to-End

With both engine (TUI) and research running + Redis available:

1. Research publishes a signal → Redis `trading:signals` channel
2. Engine's signal consumer picks it up
3. Signal appears in TUI's "SIGNALS" panel with direction + confidence + reason
4. Strategies incorporate the signal into their decision logic

## Step 6: Enable Headless Mode with HTTP API

If you want the Telegram bot later, run the engine headless with the HTTP API:

```bash
cd ~/trading-bot
./scripts/start.sh headless
```

Test the API from a third terminal:

```bash
curl http://localhost:8080/status | python3 -m json.tool
curl http://localhost:8080/positions | python3 -m json.tool
curl -X POST http://localhost:8080/pause
curl -X POST http://localhost:8080/resume
```

## Step 7: Set Up Telegram Bot (Optional)

1. Message [@BotFather](https://t.me/botfather) on Telegram
2. Send `/newbot` and follow prompts (name it `MyTradingBot` or whatever)
3. Copy the token
4. Run:

```bash
export TELEGRAM_TOKEN="your_bot_token"
export API_URL="http://localhost:8080"
cd ~/trading-bot
./scripts/start.sh telegram
```

5. Send `/start` to your bot on Telegram

## Step 8: Full Stack (All Services)

```bash
cd ~/trading-bot
./scripts/start.sh all
```

Starts Redis → engine headless + API → research → Telegram bot all in one command.

## Step 9: Step-by-Step Testing Flow

Progressively verify each layer works before trusting anything with real money:

### 9a. Engine connectivity (5 min)
- [ ] Binance WebSocket connects without errors
- [ ] Order book data flows in (watch TUI debug logs with `--verbose`)
- [ ] `Ctrl+C` triggers graceful shutdown

### 9b. Paper trading decisions (30-60 min)
- [ ] Strategies generate decisions (watch log: `"scalping" "buy" "order book imbalance"`)
- [ ] Risk manager correctly blocks excess positions
- [ ] Check trade journal: `sqlite3 trading.db "SELECT * FROM trades ORDER BY created_at DESC LIMIT 5;"`
- [ ] No unrealistic fills (wrong prices, impossible amounts)

### 9c. Signal integration (1-2 hours)
- [ ] Research service runs without crash loops
- [ ] Signals appear in TUI signal panel
- [ ] Signals with TTL expire and get dropped
- [ ] Low-confidence signals don't override strategy decisions

### 9d. Kill switch (manual test)
- [ ] Press `k` in TUI → breaker trips, orders cancelled, strategies paused
- [ ] `curl -X POST http://localhost:8080/kill` → same effect from API
- [ ] `/kill` from Telegram → same effect

### 9e. Resilience (overnight)
- [ ] Let it run overnight in paper mode
- [ ] Check in morning: no crashes, no memory leaks, DB has trade entries
- [ ] Review trade log for reasonableness of decisions

## Step 10: Real Money Readiness Checklist

Before setting `live_trading: true`:

- [ ] Paper mode ran for at least 24 continuous hours with zero crashes
- [ ] At least 50 paper trades executed, all with reasonable P&L
- [ ] Risk checks verified: max daily loss, circuit breaker, position caps all trigger correctly
- [ ] Kill switch tested and working from TUI, API, and Telegram
- [ ] API keys on Binance MAINNET (not testnet) with IP whitelist + withdrawal disabled
- [ ] Start with `max_order_size_usd: 20` in config (tiny amounts)
- [ ] First live run is supervised (watch the TUI for the first hour)

## Common Issues

| Symptom | Likely Cause | Fix |
|---------|-------------|-----|
| `websocket: bad handshake` | Binance API unreachable | Check internet, wait, retry |
| `invalid API-key` | Wrong or expired testnet key | Regenerate testnet keys |
| `redis: connection refused` | Redis not running | `docker compose up -d` |
| `signal expired on arrival` | Signal TTL too short vs network latency | Increase `ttl_seconds` in research config |
| TUI flashing/blank | Terminal too small | Resize terminal to at least 80x24 |
| `position blocked by risk` spam | Strategy producing too many decisions | Tighten strategy thresholds in config |
