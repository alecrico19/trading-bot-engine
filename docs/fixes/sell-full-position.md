# Sell Full Position, Not 8% Increments

## Problem
When tick-momentum detects a reversal, it sells 0.0012 BTC (8% of equity size), not the full holding. If the bot accumulated 0.0036 BTC from multiple buys, it takes 3 separate sells to close the position — each potentially at a lower price.

## Fix

### In tick-momentum Evaluate: set Amount to total holding
Instead of `Amount: 0` (which executeDecision sizes as 8% of equity), set `Amount` to the actual position quantity for that symbol:

```go
// Before sell decision, compute total holding
var holding float64
for _, p := range state.Positions {
    if p.Symbol == state.Symbol {
        holding += p.Amount
    }
}
if holding > 0 {
    return &types.Decision{
        ...
        Amount: holding,  // Sell EVERYTHING
    }
}
```

### Files
`tick_momentum.go` — ~5 lines for holding calculation

### Expected
After fix: SELL closes the ENTIRE position in one order. No more piece-by-piece selling at declining prices.