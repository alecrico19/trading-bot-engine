# Mini-Profit Exit Loop — Fix

## The Bug

Mini-profit sell doesn't reset `entryPrice` after firing:

```go
if pnlPct >= 0.0005 {  // Same condition fires every 3s
    sell 50%            // Sells 50% of remaining
    // ← entryPrice stays the same!
}
// 3 seconds later → same condition, fires again
// And again. And again. Each sell at a tiny loss from spread.
```

Cycle: 0.0012 BTC → sell 0.0006 → keep 0.0006 → 3s later sell 0.0003 → keep 0.0003 → 3s later sell 0.00015 → etc. 20 trades of diminishing size, each losing $0.008 to the spread.

## Fix

Reset `entryPrice` and `entryTime` after the mini-profit sell, same as stop-loss and take-profit do:

```go
if pnlPct >= 0.0005 {
    amount := baseHeld * 0.5
    if amount > 0.00001 {
        ...
        e.recordExit(order, entry, symbol, "mini-profit")
        e.mu.Lock()
        e.entryPrice[symbol] = 0
        e.entryTime[symbol] = time.Time{}
        e.highWater[symbol] = 0
        e.mu.Unlock()
    }
}
```

## Expected
One mini-profit sell per position, then entry cleared. No more repeat-loop.

## Files
`engine.go` — +5 lines (add entry reset after mini-profit block)