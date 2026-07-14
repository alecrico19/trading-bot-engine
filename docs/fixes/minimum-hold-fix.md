# Bot Missed $16 Profit — Bought and Sold At Same Price

## The Data

Bot bought BTC at $62,526 and immediately sold at $62,526 — same second. Then price rose to $62,522 (+$16 from the bottom). Bot captured $0 of the move.

```
9:16 PM BUY  $62,526.01 (caught the dip) ← good entry!
9:16 PM SELL $62,526.00 (same price)      ← sold immediately
9:17 PM price at $62,522                   ← already out, missed it
```

## Root Cause

Scalping fires BUY on upticks (>60%), then ONE downtick reverses the signal and fires SELL. In a volatile market, the buy and sell happen on consecutive trades with zero price movement. The bot captures none of the recovery.

## Fix: Minimum Hold Time

Require the position to be held for at least 3 ticks before the first sell is allowed. This gives the price time to actually move:

```go
// In scalping Evaluate, track buyTickTime:
if justBought && ticksSinceBuy < 3:
    return nil  // don't sell yet
```

OR: block sell signals for 500ms after a buy. This lets the price develop before considering an exit.

## Or: Require Minimum Price Change

Only sell if the price has moved > 0.02% from the entry since the last decision:
```go
if time.Since(s.lastDecision[symbol]) < 500*time.Millisecond {
    return nil  // wait 500ms before next decision
}
```

## Files
`engine/scalping.go` — add minimum time/distance gate after buy (~5 lines)
`engine/tick_momentum.go` — same fix (~5 lines)

## Effect

After fix: buy at $62,526 → hold for 500ms (3-5 ticks) → price moves to $62,530 → sell at profit instead of same price.