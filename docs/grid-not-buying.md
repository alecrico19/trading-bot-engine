# Grid Not Buying — Diagnosis

## Status

Grid registered at 4:09 PM but 0 BTC trades. ETH had 3 trades (mean-reversion). Grid should have bought immediately on the first BTC trade event.

## Why Grid Might Not Fire

1. **Ticker fetch failed** — `FetchTicker(ctx, "BTCUSDT")` returns nil → grid returns nil
2. **Trade stream silence** — no BTC trade events since restart? Unlikely.
3. **Strategy filter** — `isStrategyForSymbol("grid", symbol)` might return false
4. **Grid bought but order was rejected** — risk manager or paper trader blocked it

## Fix: Add Diagnostic Log to Grid

Add a log line at entry to Evaluate showing why it returns:
```go
s.logger.Debug().Str("symbol", symbol).Dur("since_last", time.Since(s.lastBuy[symbol])).Msg("grid evaluate")
```

This tells us: is Evaluate being called? What's the lastBuy time? Why is it returning nil?

## Files

`engine/grid.go` — add diagnostic log (~3 lines)