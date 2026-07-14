# Time-Exit Sells Not Recorded — Fix

## Root Cause
`checkStopLoss` calls `e.orderMgr.PlaceOrder()` directly. This fills the order but bypasses:
- `executeDecision` (P&L calculation, trade counting, strategy P&L)
- `db.RecordTrade` (no DB entry)
- `riskMgr.RecordTrade` (no daily P&L tracking)
- Diagnostic logging

The fills happen silently — no trace in the journal.

## Fix
Route time-exit and stop-loss sells through `executeDecision` instead of `PlaceOrder` directly. Create a `Decision` struct and call `e.executeDecision()`:

```go
// Instead of:
e.orderMgr.PlaceOrder(ctx, symbol, types.SideSell, types.TypeMarket, amount, 0, "time-exit")

// Do:
decision := &types.Decision{
    Action:   types.ActionSell,
    Symbol:   symbol,
    Side:     types.SideSell,
    Amount:   amount,
    Type:     types.TypeMarket,
    Reason:   "time-based exit (30s, in profit)",
    Strategy: "scalping",
}
e.executeDecision(ctx, decision)
```

This way the time-exit sell goes through the same path as scalping sells — P&L calculated, DB recorded, risk tracked, diagnostic logged.

## Files
`engine.go` — replace PlaceOrder calls in checkStopLoss with executeDecision calls (~10 lines changed)