# Cooldown + Full Position Sell

## Two Fixes, One Combo

### Fix 1: Cooldown (2s between signals per symbol)
Prevents the 15-sells-in-2-seconds flood. After generating a decision for a symbol, skip all subsequent decisions for that symbol for 2 seconds.

```go
// In tick_momentum.go Evaluate, at the top:
if time.Since(s.lastDecision[symbol]) < 2*time.Second {
    return nil
}
```

### Fix 2: Sell full position
When exiting, sell ALL of the holding, not just 8% of equity.

```go
// Before SELL decision:
var holding float64
for _, p := range state.Positions {
    if p.Symbol == state.Symbol {
        holding += abs(p.Amount)
    }
}
if holding > 0 {
    return &types.Decision{..., Amount: holding}
}
```

### Files
`tick_momentum.go` — ~15 lines total (add `lastDecision` map, holding calc)

### Expected
After fix:
1. Upticks → BUY (one entry, 8% of equity)
2. Downticks → SELL (entire position, one order)
3. More downticks → skipped (cooldown + position closed)
4. Upticks return → BUY again (new entry)