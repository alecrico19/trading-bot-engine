# Add Price-Based Profit Exit

## Problem
Bot bought BTC at $64,440. Price rose to $64,456 (+$16, +0.025%). No exit triggered because:
- Fixed take-profit needs 1% ($644 move) — too far for scalping
- Order book imbalance exit needs selling pressure — balanced book = no sell signal
- Mean reversion needs RSI extremes — price is trending, not reverting

The bot has no "I'm in profit, lock it in" instinct for small gains.

## Fix

### Quick: Lower take-profit threshold for scalping
Add a new config option `scalp_profit_pct: 0.001` (0.1%). This is a tight profit target specifically for scalping entries. When a scalping trade reaches 0.1% profit, sell it. This catches $64 moves on BTC and $1.80 moves on ETH.

```yaml
risk:
  take_profit_1r_pct: 0.001  # Was 0.01 (1%), now 0.1%
```

### Better: Add scalping-specific exit in the stop-loss goroutine
In `checkStopLoss`, check if the position was opened by the scalping strategy. If so, use a tighter profit target (0.1% instead of 1%). Mean reversion positions keep the wider target (1-2%).

```go
if entryPrice is from scalping:
    target = 0.001  // 0.1%
elif entryPrice is from mean-reversion:
    target = 0.01   // 1%
```

### Files
- `config.yaml` — add scalp_profit_pct (1 line)
- `engine.go` — per-strategy profit targets (~5 lines)
- `config.go` — add config field (2 lines)