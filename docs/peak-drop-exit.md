# Peak-Drop Exit: Sell When Price Drops From High

## What Happened

Buy BTC at $62,728 → peaks at $62,782 (+$54, +0.086%) → gently declines. No sell fires because:
- Mini-profit: disabled
- Take-profit: needs 1-2% ($627+)
- Scalping sell: needs 40%+ downticks (fading decline = mixed ticks)
- Tick-momentum sell: needs 45%+ downticks

The profit evaporated because no exit caught the transition from "going up" to "going down."

## Fix: Sell When Price Drops From Peak

Track `highWater` per symbol (already done for trailing stop). When price drops >0.02% from the high, sell 50% immediately.

```go
// In checkStopLoss, right after pnlPct calc:
if high > entry && (high - ticker.Last) / high > 0.0002 {
    // Price dropped 0.02% from recent high → sell half
    amount := baseHeld * 0.5
    if amount > 0.00001 {
        e.orderMgr.PlaceOrder(ctx, symbol, SideSell, TypeMarket, amount, ...)
        e.recordExit(...)
    }
}
```

Trusty: track the last high point. If price drops 0.02% from that high point, the move is exhausting. Sell half.

## Files
`engine.go` — add peak-drop check in checkStopLoss (~10 lines)

## Benefits
- Catches your $62,782 peak → $62,770 drop immediately
- Doesn't wait for momentum reversal (multiple downticks needed)
- Leaves 50% to ride in case of recovery
- Uses existing highWater tracking (zero new state)