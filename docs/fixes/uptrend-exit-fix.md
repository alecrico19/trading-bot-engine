# No Sell on $30 BTC Rise — Fix

## Root Cause

Both scalping and tick-momentum only exit on **momentum reversal** (downticks). If BTC goes straight up $30 with all upticks, the bot rides it up but never sells. Take-profit at 1-2% ($640-1280 on BTC) is way too far for a $30 move (0.05%).

The bot enters on upticks and holds forever if the uptrend never reverses.

## Fix: Profit-Based Exit for Uptrend Rides

Add a check in the exit logic: if we're in a long position AND price has moved up > 0.05% from entry → sell 50% to lock in the gain. The remaining 50% rides with a trailing stop.

```go
// In checkStopLoss, before the existing exits:
if pnlPct >= 0.0005 {  // 0.05% profit
    sell 50% of position at market
    keep entry for remaining 50%
}
```

This catches the $30 BTC move (0.05%) and locks it in without waiting for a reversal that might never come.

## Alternatively: Re-enable Time-Exit at Higher Threshold

Instead of the old 90s time-exit (closes EVERYTHING at tiny profit), re-enable it but only if profit > 0.05%:

```go
if time.Since(entryTime) > 60*time.Second && pnlPct > 0.0005 {
    sell 50%
}
```

## Files
- `engine.go` — add profit-based partial exit in checkStopLoss (~10 lines)
- `config.yaml` — add `scalp_exit_pct: 0.0005` (~1 line)

## Expected
BTC goes up $30 → bot sells 50% at +$15 profit → remaining 50% rides the trend upward. If trend reverses later, the trailing stop or momentum sell handles the rest.