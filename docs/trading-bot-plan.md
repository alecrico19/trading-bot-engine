# Trading Bot Architecture Plan

## TL;DR

Go execution engine + Python research service, communicating via Redis pub/sub. Crypto-first (Binance), stock adapter later. Local TUI dashboard. All free tools, zero paid dependencies.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────┐
│                 TUI Dashboard                    │
│              (Go - Bubble Tea)                   │
│        monitor / start-strategy / kill-switch    │
└──────┬──────────────────────────────┬───────────┘
       │ HTTP REST                     │ HTTP REST
       ▼                               ▼
┌──────────────────┐         ┌─────────────────────┐
│  Research Service │ signals │  Execution Engine    │
│  (Python)         ├────────►│  (Go)                │
│                   │  Redis  │                      │
│  • News scraper   │ Pub/Sub │  • Strategy engine   │
│  • Sentiment      │         │  • Risk manager      │
│  • Tech analysis  │         │  • Order manager     │
│  • Signal gen     │         │  • Exchange adapters │
│                   │         │  • Paper trader      │
└──────┬────────────┘         └──────────┬──────────┘
       │                                 │
       │ market data                     │ order execution
       │ (ccxt REST)                     │ (ccxt + WebSocket)
       ▼                                 ▼
┌─────────────────────────────────────────────────┐
│              Exchange Layer                      │
│  Binance │ Coinbase │ Kraken │ ...              │
│  (ccxt unified API, Go + Python bindings)       │
└─────────────────────────────────────────────────┘
       │                       │
       ▼                       ▼
┌──────────────┐     ┌──────────────────┐
│    Redis      │     │     SQLite       │
│  • Pub/Sub    │     │  • Trade journal │
│  • Cache      │     │  • Signal log    │
│               │     │  • Config store  │
└──────────────┘     └──────────────────┘
```

### Key Design Decisions

1. **Async signals, not sync queries** — The execution engine never blocks waiting for research. It consumes a stream of pre-computed signals. This prevents research latency from delaying trade execution.

2. **Scalping lives entirely in the execution engine** — Scalping decisions are made from order book depth, tick-by-tick price, and spread data (all via WebSocket). The research service provides *macro bias* only (e.g., "BTC bullish today, favor longs").

3. **Signal TTL** — Every signal has a TTL. A stale "buy" signal from 5 minutes ago is worse than no signal at all. The execution engine ignores expired signals.

---

## Technology Stack (All Free)

### Execution Engine (Go)
| Component | Tool | License | Why |
|-----------|------|---------|-----|
| Language | Go 1.22+ | BSD | Compiled, fast, goroutines for concurrent feeds, single binary deploy |
| Exchange API | ccxt (Go) | MIT | 100+ exchanges, unified API, WebSocket support (ccxt.pro) |
| IPC | go-redis | MIT | Pub/sub + caching in one dependency |
| Persistence | go-sqlite3 | MIT | Zero-config, embedded, single file |
| TUI | Bubble Tea + Lip Gloss | MIT | Elm-like TUI framework, looks great in terminal |
| Config | Viper | MIT | YAML/JSON/ENV config, hot reload |
| Logging | Zerolog | MIT | Zero-allocation structured logging |

### Research Service (Python)
| Component | Tool | License | Why |
|-----------|------|---------|-----|
| Language | Python 3.11+ | PSF | ML/NLP ecosystem, fast prototyping |
| Exchange data | ccxt (Python) | MIT | Same unified API, sync HTTP calls for research (don't need WebSocket here) |
| Tech indicators | pandas-ta + ta-lib | GPL/BSD | 200+ indicators, vectorized |
| Backtesting | vectorbt | GPL | Vectorized backtesting, fast, free |
| Sentiment/NLP | TextBlob / VADER | MIT | Free, no GPU needed, runs locally |
| News ingestion | feedparser + NewsAPI | MIT / free tier | RSS feeds (free) + NewsAPI (100 req/day free) |
| IPC | redis-py | MIT | Pub/sub to emit signals |

### Shared Infrastructure
| Component | Tool | License | Why |
|-----------|------|---------|-----|
| Message bus | Redis 7+ | BSD | Industry standard pub/sub, also used as signal cache |
| Time-series DB | SQLite (initial) | Public domain | Start simple; migrate to TimescaleDB or ClickHouse if scale demands |

### No Paid Tools Required
Nothing in this stack costs money. The only scenario where paid tools become necessary:
- **Colocated servers** — if you need <1ms latency to exchange matching engines (this is HFT territory, not our use case)
- **Premium news feeds** (Bloomberg, Reuters) — not needed; RSS + free APIs + exchange data is sufficient for scalping macro context
- **Historical tick data** — free from Binance API for past few days; paid vendors if you need years of tick history for backtesting

---

## Signal Protocol

All signals flow Redis channel `trading:signals`. JSON format:

```json
{
  "id": "sig_a1b2c3",
  "timestamp": "2026-07-09T14:30:00Z",
  "source": "research",
  "type": "bias",
  "symbol": "BTC/USDT",
  "direction": "long",
  "confidence": 0.72,
  "factors": {
    "rsi": 28.4,
    "sentiment_score": 0.65,
    "news_headline": "SEC approves Bitcoin ETF",
    "volume_spike": 2.3
  },
  "reason": "Oversold RSI + positive sentiment + volume spike",
  "ttl_seconds": 300
}
```

### Signal Types
- **`bias`** — Macro directional bias. "Favor longs on BTC today." TTL: 5-15 minutes.
- **`alert`** — Event-driven. "Fed announcement in 2 minutes — tighten stops." TTL: event window.
- **`sentiment`** — Aggregate sentiment score. TTL: 15-60 minutes.

### Signal Consumption Rules (Execution Engine)
1. Ignore signals with expired TTL
2. Bias signals below `confidence: 0.6` are advisory only — don't override strategy rules
3. Alert signals force risk reduction (tighten stops, reduce position size)
4. Multiple conflicting signals → weight by confidence, pick winner

---

## Component Details

### 1. Execution Engine (Go)

The critical path. Everything here must be non-blocking in the hot loop.

#### Strategy Engine
- Pluggable strategy interface: `type Strategy interface { Evaluate(MarketState) *Decision }`
- Built-in strategies: scalping (order book imbalance), mean reversion (Bollinger bands), momentum breakouts
- Each strategy loaded from YAML config, runs in its own goroutine
- Strategies are symbol-scoped: one strategy instance per symbol per timeframe

#### Risk Manager
- Position sizing: Kelly criterion, fixed fractional, or configurable max % of portfolio
- Max daily loss limit → hard stop, liquidate all positions, emit alert
- Max concurrent positions cap
- Correlation check: don't open both BTC and ETH longs if they're 0.9 correlated
- Drawdown circuit breaker at configurable %

#### Order Manager
- Limit orders by default (scalping needs precise entry/exit)
- Market orders as fallback with configurable slippage tolerance
- Order state machine: Pending → Placed → PartiallyFilled → Filled | Cancelled | Rejected
- Order book simulation for paper trading mode

#### Exchange Adapters
```go
type Exchange interface {
    Name() string
    FetchBalance() (*Balance, error)
    FetchTicker(symbol string) (*Ticker, error)
    FetchOrderBook(symbol string, depth int) (*OrderBook, error)
    CreateOrder(symbol, side, orderType string, amount, price float64) (*Order, error)
    CancelOrder(orderID string) error
    SubscribeOrderBook(symbol string) (<-chan *OrderBook, error)
    SubscribeTrades(symbol string) (<-chan *Trade, error)
    SubscribeAccount() (<-chan *AccountUpdate, error)
}
```
- Binance implementation first (via ccxt Go)
- Paper trading adapter (simulates fills against real order book)
- Coinbase, Kraken follow the same interface

### 2. Research Service (Python)

Runs on its own cadence. Never on the critical path.

#### News Ingestion
- RSS feeds: CoinDesk, CoinTelegraph, FXStreet, Bloomberg Crypto
- NewsAPI free tier: keyword queries for "Bitcoin", "Ethereum", "SEC crypto", "Fed rate"
- Reddit API (free): r/CryptoCurrency, r/BitcoinMarkets top posts
- Deduplication by URL hash, stored in SQLite with sentiment scores

#### Sentiment Analysis
- VADER (rule-based, no training data needed): headline sentiment -1 to +1
- TextBlob: longer article body sentiment
- Aggregate: weighted average, recency-weighted

#### Technical Analysis
- pandas-ta for 200+ indicators
- Multi-timeframe: 1m, 5m, 15m, 1h for scalping context
- Signal generation rules (configurable YAML), e.g.:
  ```yaml
  signal_rules:
    oversold_bounce:
      conditions:
        - rsi_14 < 30
        - volume_ratio > 1.5
        - price_above_vwap: false
      action: bias_long
      confidence: 0.7
  ```

#### Backtesting Engine
- vectorbt for portfolio-level backtesting
- Feed historical OHLCV from ccxt (free, last ~1000 candles per call)
- Walk-forward optimization: train on N days, test on N+1, repeat
- Output: Sharpe ratio, max drawdown, win rate, profit factor per strategy config

### 3. TUI Dashboard (Go)

Built with Bubble Tea. Runs in terminal. Panels:

```
┌─ Trading Bot v0.1 ────────────────────────────────┐
│ BTC/USDT 67,214.50 ▲ +2.3%  │  Signals (3 active)  │
│ ─────────────────────────── │  ├ bias:long  0.72   │
│ [████████░░░░] RSI 28 (OS)  │  ├ bias:short 0.45   │
│ Bid: 67,213  Ask: 67,215    │  └ alert:tighten     │
│ Vol: 1,234 BTC (24h)        │                      │
│                             │  Strategies          │
│ Positions                   │  ├ scalping   ▶ RUN  │
│ ├ ETH/USDT LONG 0.5 ▲+$42  │  ├ mean_rev   ▶ RUN  │
│ └ BTC/USDT SHORT 0.1 ▼-$15 │  └ momentum   ⏸ PAUSE│
│                             │                      │
│ P&L Today: +$27.42          │  [Q]uit [S]tart      │
│ P&L Week:  +$341.19         │  [P]ause [K]ill all  │
└────────────────────────────────────────────────────┘
```

### 4. Config & Persistence

#### SQLite Tables
```sql
CREATE TABLE trades (
    id TEXT PRIMARY KEY,
    exchange TEXT, symbol TEXT, side TEXT,
    amount REAL, price REAL, fee REAL,
    strategy TEXT, signal_id TEXT,
    created_at TIMESTAMP
);

CREATE TABLE signal_log (
    id TEXT PRIMARY KEY,
    source TEXT, type TEXT, symbol TEXT,
    direction TEXT, confidence REAL,
    action_taken TEXT,
    created_at TIMESTAMP
);

CREATE TABLE daily_pnl (
    date DATE PRIMARY KEY,
    realized_pnl REAL,
    unrealized_pnl REAL,
    trade_count INTEGER,
    win_rate REAL
);
```

#### YAML Configuration
```yaml
exchanges:
  binance:
    api_key: "${BINANCE_API_KEY}"
    api_secret: "${BINANCE_API_SECRET}"
    testnet: true

risk:
  max_position_pct: 0.05
  max_daily_loss_pct: 0.02
  max_concurrent_positions: 5
  circuit_breaker_drawdown_pct: 0.10

strategies:
  scalping:
    enabled: true
    symbols: ["BTC/USDT", "ETH/USDT"]
    order_book_depth: 20
    min_spread_pct: 0.02
    position_size_usd: 100

research:
  update_interval_seconds: 60
  sentiment_sources: ["rss", "newsapi", "reddit"]
  signal_min_confidence: 0.6
```

---

## Development Phases

### Phase 1: Core Execution Engine (Weeks 1-2)
- Go project scaffold, config loader, logging
- Exchange adapter interface + Binance implementation (ccxt Go)
- Order manager with state machine
- Risk manager (position sizing, daily loss limit)
- **Paper trading mode only** — simulates fills against live order book
- CLI command: `trading-bot run --paper`

**Deliverable**: Bot can place and manage paper trades on Binance testnet

### Phase 2: TUI Dashboard (Week 3)
- Bubble Tea dashboard showing positions, P&L, active orders
- Hotkeys: start/pause strategies, kill switch, switch symbols
- Live order book visualization
- Trade journal view

**Deliverable**: Monitor and control the bot from terminal

### Phase 3: Research Service (Weeks 4-5)
- Python project scaffold, Redis pub/sub integration
- News ingestion pipeline (RSS + NewsAPI)
- VADER sentiment scoring
- Technical analysis with pandas-ta, signal generation
- Signal emitter: publishes `trading:signals` to Redis

**Deliverable**: Execution engine receives and logs research signals (no action yet)

### Phase 4: Strategy Engine (Weeks 6-7)
- Plug-and-play strategy interface
- Implement scalping strategy: order book imbalance detection
- Implement mean reversion: RSI + Bollinger bands
- Wire signals into strategy decisions
- Walk-forward backtesting with vectorbt

**Deliverable**: Bot makes autonomous trading decisions in paper mode

### Phase 5: Live Trading & Hardening (Weeks 8-9)
- Real money mode with confirmation prompts
- Circuit breakers, sanity checks (max order size, price deviation)
- Comprehensive logging and trade journal
- Graceful shutdown (cancel all orders, close monitor positions)
- Alerting: console + optional Telegram/Discord webhook

**Deliverable**: Production-ready live trading

### Phase 6: Additional Exchanges & Markets (Week 10+)
- Coinbase adapter
- Kraken adapter
- Stock adapter (Alpaca) — proves the adapter pattern works
- Cross-exchange arbitrage scanner (optional)

### Phase 7: Remote Access (Week 12+)
- **Telegram bot**: consume the same REST API the TUI uses. Commands:
  - `/status` — P&L, positions, active strategies
  - `/start <strategy>` / `/pause <strategy>` — remote control
  - `/alert` — push notifications for risk events (circuit breaker, daily loss limit hit)
- **Optional iOS widget**: same API, read-only dashboard (SwiftUI widget, free to build)
- REST API already built in Phase 2, so this is purely a frontend layer — no engine changes

---

## Directory Structure

```
trading-bot/
├── config/
│   └── config.yaml
├── execution/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── engine/
│   │   ├── exchange/
│   │   ├── order/
│   │   ├── risk/
│   │   └── signal/
│   ├── tui/
│   │   ├── dashboard.go
│   │   └── components/
│   ├── db/
│   └── go.mod
├── research/
│   ├── main.py
│   ├── news/
│   │   ├── rss.py
│   │   ├── newsapi.py
│   │   └── reddit.py
│   ├── sentiment/
│   │   └── vader.py
│   ├── technical/
│   │   ├── indicators.py
│   │   └── signals.py
│   ├── backtest/
│   │   └── vectorbt_runner.py
│   ├── publisher.py
│   └── requirements.txt
├── scripts/
│   ├── start.sh
│   └── backtest.sh
├── docker-compose.yml
└── README.md
```

---

## What We're NOT Building (Yet)

- **ML-based price prediction** — Overkill for v1. Rule-based strategies work well enough for scalping. ML can be added to the research service later without changing anything else.
- **GUI/Web dashboard** — TUI first. Web GUI is a separate project that consumes the same REST API.
- **Multi-instance/clustering** — Single binary, local machine. Scale horizontally later if needed.
- **Real-time P&L accounting** (FIFO/LIFO tax lot tracking) — Track realized P&L only. Tax reporting can be done post-hoc from the trade journal.
- **Social trading / copy trading** — Out of scope.

---

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Bug causes real money loss | Start paper-only. Paper → testnet → small real amounts. Kill switch in TUI. |
| ccxt Go bindings immature | Can fall back to Go native Binance client (`github.com/adshao/go-binance`). ccxt Go is the abstraction layer; swap the implementation if needed. |
| Redis as SPOF | Local Redis is reliable for single-machine. If Redis is down, engine runs on last-known signals (graceful degradation). |
| Overfitting backtests | Walk-forward validation, out-of-sample testing, paper trading before live. |
| Exchange rate limits | ccxt handles rate limiting. Strategies include cooldown periods to avoid rapid-fire orders. |

---

## Decisions Made

1. **Symbols**: BTC/USDT + ETH/USDT to start
2. **Position sizing**: Fixed $100 USD per trade (10-20% of $500-1000 portfolio). Kelly is premature without known edge; %-based makes positions too small for fees.
3. **Strategies**: Built-in order book imbalance scalping + RSI/Bollinger mean reversion. Add more as data comes in.
4. **CLI mode**: Yes. Included for headless operation, scriptable status checks, and faster backtesting.
