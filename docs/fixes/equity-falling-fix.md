# Equity Falling — Root Cause & Fix

## The Data

```
Portfolio: $982.55 (started $1000)
Cash:      $904.00
P&L:       -$7.90
Trades:    108 (50+ round-trips)
Time-exit P&L: -$7.32  ← biggest loser
Scalping P&L:  -$0.58
Fees/round-trip: $0.15
```

## Root Cause

The scalping strategy at 100ms intervals is detecting noise, not edge:
1. Bid volume surges → bot buys → but price already peaked → position reverses
2. 30 seconds later → time-exit closes at small loss → repeat × 50
3. Each round-trip loses 0.2% in fees alone

## Critical Fix: Increase Time-Exit from 30s → 90s

The 30-second exit is the biggest loser (-$7.32). It closes positions before they have time to play out. Moving to 90 seconds gives trades room to recover from micro-reversals.

## Broader Fix: The Strategy Has No Edge

The order-book-imbalance strategy at 100ms doesn't predict price direction. Options:

### Option A: Turn off scalping, rely on mean-reversion only
Mean reversion (RSI + BB) has 0 trades so far because it needs 100 candles to warm up (~8 min). After warmup, it has a 63% win rate in backtests. More reliable than scalping.

### Option B: Fix the entry logic — buy weakness, sell strength
Current: BUY on bid surge (buyers are aggressive = price already up = bad entry)
Fix: BUY on ask surge (sellers are aggressive = price already down = good entry)

### Option C: Add minimum hold time + market direction filter
Don't sell in the first 90 seconds. Only exit on move > 0.1% in favorable direction. Use 1-minute candle direction as filter.

## Recommendation

**Do A (disable scalping) + increase time-exit to 90s.** Scalping is the problem. Mean reversion has a mathematical edge (confirmed in backtests). Let scalping hibernate while mean reversion trades, or use scalping only as an entry signal alongside a longer timeframe trend filter.

### Files
- `config.yaml` — disable scalping (`enabled: false`)
- `mean_reversion.go` — already built, just needs time to warm up
- `engine.go` — change 30s → 90s for time-exit

### Expected
Without scalping bleeding fees, portfolio should stabilize near $1000. Mean reversion takes fewer, higher-quality trades at lower frequency.