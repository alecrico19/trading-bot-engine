# Fix: Bot Buying High, Selling Low

## Root Cause

The scalping strategy reacts to order book imbalance AFTER the price has moved:
1. Heavy bid volume → price already rose → bot BUYs at the top
2. Heavy ask volume → price already fell → bot SELLs at the bottom

Also: the bot fires SELL signals with NO open position (selling short into nothing).

## Fix Plan

### Fix 1: Block sells when no position held (CRITICAL)
In `scalping.go`, before returning a sell decision, check if the bot actually holds the asset. If not, skip the sell.

```go
// Before sell decision: check if we hold the base asset
hasPosition := false
for _, p := range state.Positions {
    if p.Symbol == state.Symbol && abs(p.Amount) > 0.00001 {
        hasPosition = true
        break
    }
}
if !hasPosition {
    return nil // don't sell what we don't own
}
```

### Fix 2: Add momentum confirmation (HIGH)
Instead of buying the moment bid volume spikes, require price to have moved UP in the last 3-5 ticks. This confirms the momentum hasn't exhausted.

### Fix 3: Use limit orders below market for buys, above for sells (MEDIUM)
Currently: buy at bestAsk (top of book), sell at bestBid (top of book).
Change to: buy at bestBid (try to get filled at lower price), sell at bestAsk (try to sell at higher price). This flips the strategy from momentum-following to mean-reverting on the spread.

### Fix 4: Add cooldown between opposite signals (MEDIUM)
After a BUY, wait N seconds before allowing a SELL on the same symbol. Prevents the buy-high-then-immediately-sell-low pattern.

## Files
- `scalping.go` — Fix 1 (+10 lines), Fix 2 (+15 lines), Fix 4 (+10 lines)
- `scalping.go` — Fix 3: change `bestAsk` to `bestBid` on buys, `bestBid` to `bestAsk` on sells (~4 line changes)

## Recommendation
Do Fix 1 (block sells with no position) and Fix 3 (flip limit order prices) immediately. Fix 1 stops the bleeding. Fix 3 changes the strategy from "buy high sell low" to "buy low sell high" by placing limit orders on the favorable side of the spread instead of crossing it.