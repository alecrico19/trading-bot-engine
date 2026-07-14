# Scalping on Chart Data (Trade Stream) Instead of Order Book

## The Idea

Scalping currently uses `bid_volume / ask_volume` from order book snapshots (stale). Switch to the same real-time trade stream that tick-momentum already uses.

## How It Would Work

Instead of "order book ratio > 1.05 → buy", use:

```
Track last 3-5 trade prices (vs tick-momentum's 5-50):
- 3 consecutive UP ticks → scalping BUY
- 3 consecutive DOWN ticks → scalping SELL
- Exit on 2 consecutive opposite ticks
```

This is a FASTER version of tick-momentum — smaller window, quicker entries, more trades.

## Scalping vs Tick-Momentum

| | Scalping (proposed) | Tick-Momentum (current) |
|---|---|---|
| Trade window | 3-5 trades | 5-50 trades |
| Entry signal | 60%+ same direction | 60%+ same direction |
| Exit signal | 2 opposite ticks | Momentum reversal + cooldown |
| Hold time | 1-10 seconds | 10-60 seconds |
| Trade frequency | 20-40/hour | 2-5/hour |
| Data source | Trade stream (WebSocket) | Trade stream (WebSocket) |

## Implementation

Scalping already has a `PriceHistory` field (added for the 200 EMA trend filter, then removed). Reuse it for trade tracking:

```go
// In scalping.go Evaluate:
hist.Add(midPrice)  // already tracking mid-price from order book
// Change to: use trade prices from the trade stream
```

But the `Evaluate` method is called from the order book loop — scalping doesn't have access to the trade stream. Need to either:
1. Pipe trade data into the scalping strategy (like tick-momentum)
2. Subscribe to trades inside scalping
3. Convert scalping to evaluate from a shared trade history

**Simplest approach:** Give scalping access to the trade stream. Add to `MarketState` a `TradePrices []float64` field (last N trade prices). Populate it from the trade stream loop. Scalping reads from it.

**Even simpler:** Make the trade stream loop call `strat.Evaluate(state)` for scalping too (not just tick-momentum), passing the trade data in the state.

## Files
- `engine.go` — add scalping to trade stream evaluation loop (~3 lines)
- `scalping.go` — replace order book logic with trade-based logic (~30 lines)
- `types.go` — add trade history to MarketState (~3 lines)

~40 lines total

## Expected
Scalping fires 10-20x more frequently than tick-momentum. Both strategies use the same real-time trade data. Scalping catches micro-moves (3-5 trades), tick-momentum catches macro-moves (5-50 trades).