# Tick-Momentum Not Trading — Bug Fix

## Root Cause

`runTradeStream` receives every trade but throws it away:

```go
_ = trade  // ← BUG: trade data never reaches the strategy
decision := strat.Evaluate(state)
```

The `Evaluate` method reads from `prices` map, but nobody calls `FeedTrade()` to populate it. The strategy has no trade data to work with.

## Fix

In `runTradeStream`, call `FeedTrade` to populate the price history before evaluating:

```go
if tm, ok := strat.(*TickMomentumStrategy); ok {
    tm.FeedTrade(trade)
}
decision := strat.Evaluate(state)
```

## Also: Remove Evaluate from order book loop

The order book loop calls `Evaluate` for ALL strategies including tick-momentum. But tick-momentum doesn't use order book data — it only uses trades. Skip it in the OB loop to avoid wasting cycles:

```go
for _, strat := range e.strategies {
    if strat.Name() == "tick-momentum" {
        continue  // only evaluated from trade stream
    }
    ...
}
```

## Files
`engine.go` — fix trade data feeding (~3 lines), skip in OB loop (~3 lines)