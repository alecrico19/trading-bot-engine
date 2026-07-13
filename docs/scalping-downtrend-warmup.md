# Scalping Entering on Downtrend — Fix

## Why
Scalping warmup is only 10 trades. In a downtrend, 5 random upticks happen and the bot buys into a falling market. Tick-momentum has the same issue (50-trade warmup helps but doesn't guarantee trend awareness).

## Fix: Require broader trend confirmation

Before scalping enters, check a longer window (20 trades). The longer window shows the overall trend. Only enter if the broader trend agrees (also up).

```go
// In scalping Evaluate, after checking the 5-trade window:
longVals := hist.Values()
longWindow := longVals[len(longVals)-20:] // broader 20-trade window
if longTrend is down (less than 50% upticks):
    return nil // don't buy into downtrend
```

Trick: confirm the SHORT window signal (5 trades) with the LONG window trend (20 trades). Only enter when both agree.

Already partially implemented in tick-momentum with 50-trade warmup. Same concept for scalping, just a smaller window.

## Files
`engine/scalping.go` — add broader trend check before entry (~10 lines)