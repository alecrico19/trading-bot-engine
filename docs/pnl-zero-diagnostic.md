# P&L Still $0 on Sells — Diagnostic Plan

## Symptom
Sell at $64303.62 after buy at $64367.04. P&L should be -$0.08. Shows $0.00000000.

## Diagnosis

The P&L calc in `executeDecision`:
```go
entry := e.entryPrice[decision.Symbol]
if entry > 0 {
    pnl = (order.AvgPrice - entry) * order.Filled
}
```

Either `entry == 0` or `order.AvgPrice == 0` or `order.Filled == 0`.

All three seem non-zero based on the trade history (qty and price are set).

## Fix: Add diagnostic logging

In `executeDecision`, add a log line BEFORE the P&L calc:
```go
e.logger.Info().
    Str("side", string(order.Side)).
    Str("symbol", decision.Symbol).
    Float64("entry", entry).
    Float64("avg_price", order.AvgPrice).
    Float64("filled", order.Filled).
    Float64("pnl", pnl).
    Msg("P&L diagnostic")
```

This will tell us exactly which value is 0.

## Also fix: possible race condition

The `runOrderBookLoop` goroutine and `evaluateMeanReversion` goroutine can both fire trades for the same symbol concurrently. Both modify `entryPrice` under `e.mu`. A sell from meanReversion could reset `entryPrice` to 0 before the scalping sell reads it.

**Fix:** Use a per-symbol mutex or serialize all trade execution through a single channel.

## Also fix: entryPrice overwritten on multiple buys

Two buys on the same symbol overwrite entryPrice. The second buy's price becomes the entry, but the first buy's position is "free" — its P&L is never tracked.

**Fix:** Track a blended average entry price: if buying more while already holding, compute weighted average of old entry and new price.

## Files
- `engine.go` — add diagnostic log (+5 lines), fix entry tracking (+10 lines)