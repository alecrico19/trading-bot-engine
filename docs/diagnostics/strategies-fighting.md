# Bot Idle For 1 Hour — Strategies Fighting

## The Problem

Grid bought BTC at 10:36 AM. Mean-reversion immediately sold it at the exact same second. They're canceling each other out:

```
10:36 AM: Grid BUY $62,438
10:36 AM: Mean-reversion SELL $62,438  ← kills grid's position
```

Now BTC is at $62,486 (+$48 from grid entry) but the bot has 0 BTC. The profit opportunity is gone. Both strategies held zero after fighting.

## Why No Trades Since

- BTC at $62,486 — up from last entry at $62,438
- Mean-reversion: RSI not oversold, no buy signal
- Grid: next buy at 12:36 PM (2 hours since last)
- Result: bot sits idle with $998 in cash for over an hour

## The Fix: Grid Evaluates on Timer, Not Trade Events

Currently the grid evaluates through the trade stream loop (every trade event). But it should evaluate on a timer (every minute, just checking if 2 hours have passed). This way:

1. Grid doesn't receive unintended sell signals from mean-reversion
2. Grid only triggers on its own timer
3. Mean-reversion trades independently

OR: remove grid from the trade stream loop. Instead, add a timer-based goroutine that evaluates grid periodically.

## Files

`engine.go` — remove grid from trade stream eval loop (~1 line)
`engine.go` — add grid evaluation to trade loop (which already has a 15s timer) or create a new goroutine (~5 lines)

## Alternatively

Just run grid on a different goroutine with its own ticker:
```go
go func() {
    ticker := time.NewTicker(30 * time.Second)
    for {
        select {
        case <-ticker.C:
            for _, symbol := range symbols {
                if gridStrat.LastBuy[symbol] was >2h ago {
                    buy
                }
            }
        }
    }
}()
```