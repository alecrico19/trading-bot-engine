# Mean-Reversion Sell Spam + Tick-Momentum Dead — Fix

## Symptom 1: Mean-Reversion Flooding Sells

Log shows 112 sell orders, ALL from mean-reversion. RSI condition triggered, but no positions exist. Every sell rejected by paper trader. Mean-reversion never bought — it went straight to selling.

**Fix:** Add `hasPosition` check to mean-reversion's Evaluate. Same guard as scalping and tick-momentum:
```go
if hasPosition && (rsi > overbought || price > upperBB) {
    return sell decision
}
if !hasPosition && rsi < oversold && price < lowerBB {
    return buy decision
}
```

## Symptom 2: Tick-Momentum Dead (0 tick checks)

Tick-momentum's FeedTrade debug log never fires. Trade streams are connected (2 starts, 0 disconnects). Something in the goroutine is blocking or the strategy filter is skipping it.

Possible causes:
- The decisionCh is full (mean-reversion spam fills the buffer)
- The trade stream channel is empty (WebSocket issue)
- The strategy filter (`isStrategyForSymbol`) returns false

**Fix:** Add diagnostic logging to the trade stream goroutine to confirm data flow. Check if scalping/TM skip is working correctly.

## Files
- `mean_reversion.go` — add position check (~5 lines)
- `engine.go` — trade stream diagnostic log (~3 lines)

## Priority
**Fix mean-reversion spam first.** 112 sells flooding the decision channel blocks everything else. Once the spam stops, tick-momentum should resume.
