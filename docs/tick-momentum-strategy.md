# Tick-Based Scalping — Capture Visible Price Moves

## The Missing Piece

Binance already sends us every trade (`SubscribeTrades`). We subscribe to it but never USE the data. Each trade is:

```
{trade: "BTC bought at $64,310", quantity: 0.5, timestamp: 10:30:00.123}
```

This is EXACTLY what you see on the 1s chart. We just need to turn trade data into trading decisions.

## The Strategy: Tick Momentum

Instead of "bid/ask ratio > 1.05," use:

```
Track last 5 trade prices:
$64,300 → $64,310 → $64,320 → $64,330 → $64,340

5 consecutive UP ticks → BUY (momentum confirmed)

5 consecutive DOWN ticks → SELL (momentum confirmed)

Exit after N ticks of adverse movement or 30s timeout
```

Or even simpler:

```
Track price over last 3 seconds:
If price went UP > 0.05% → BUY (momentum is real, ride it)
If price went DOWN > 0.05% → SELL

Exit when momentum reverses (3 consecutive opposite ticks)
```

## Why This Should Work

| What you see on Binance | What the bot would do |
|------------------------|----------------------|
| BTC drops $40 in 2 seconds | Detects 5+ consecutive down ticks → BUY the dip |
| BTC jumps $40 in 2 seconds | Detects 5+ consecutive up ticks → SELL to take profit |
| BTC oscillates $20-30 | Captures micro-swings through tick-based entries/exits |

The key difference: tick momentum uses ACTUAL trades (fills at real prices), not order book snapshots (which are stale by the time we see them).

## Implementation

- Data: `SubscribeTrades` — already receiving trade events, just unused
- Strategy: New strategy in `engine/tick_momentum.go`
- Process: Track last N trade prices, detect momentum, enter/exit
- Config: `min_consecutive_ticks: 5` (minimum same-direction ticks for entry)
- Exit: 3 opposite ticks or 60s timeout at market

## Files
`engine/tick_momentum.go` — new strategy (~80 lines)
`engine/engine.go` — register strategy (~5 lines)
`config.yaml` — config section (~5 lines)

## Expectation
Captures the $40 moves you can see. Much higher trade frequency than mean-reversion. Needs paper testing to confirm edge.