# Combined Fix Plan — Sell Failures + Premature Exits

## Two Fixes, Both Needed

### Fix 1: Cap Sell Amount at Actual Balance (~5 lines)
Sells try to sell 0.001279 BTC but only 0.001274 is available (fees). Cap at actual holdings.

```go
// In executeDecision, after amount calc but before PlaceOrder:
if decision.Side == types.SideSell {
    cap := e.getBaseHeld(decision.Symbol)
    if cap > 0 && amount > cap {
        amount = cap
    }
}
```

### Fix 2: Minimum 500ms Hold After Buy (~5 lines)
Prevents selling on the very next tick before price can move.

```go
// In scalping + tick-momentum Evaluate, after buy decision:
if lastSide == "buy" && time.Since(lastDecision[symbol]) < 500*time.Millisecond {
    return nil
}
```

### Files

| File | Fix 1 | Fix 2 |
|------|-------|-------|
| `engine.go` | +8 lines | — |
| `scalping.go` | — | +3 lines |
| `tick_momentum.go` | — | +3 lines |

### Expected

After fixes:
- Sells actually FILL (Fix 1)
- Sells capture actual price movement, not same-second reversals (Fix 2)
- Your $62,526 → $62,522 move: bought, held 500ms, price moved up, sold at profit

**Total: ~14 lines across 3 files. Build both now.**