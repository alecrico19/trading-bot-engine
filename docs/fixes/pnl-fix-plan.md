# P&L Fix

## Bug
`executeDecision` calculates P&L against current ticker price, which is always ≈ fill price for paper trades. P&L always reads $0.

## Fix
Track entry price per position. Record P&L on the **exit** trade only.

In `executeDecision`:
```go
// Buy: record entry price, P&L = 0
if order.Side == types.SideBuy {
    e.entryPrice[order.Symbol] = order.AvgPrice
    pnl = 0
}
// Sell: P&L = (sell - entry) * qty
if order.Side == types.SideSell {
    entry := e.entryPrice[order.Symbol]
    if entry > 0 {
        pnl = (order.AvgPrice - entry) * order.Filled
    }
    e.entryPrice[order.Symbol] = 0
}
```

`entryPrice` map already exists (added for stop-loss). Just fix the P&L calc in executeDecision.

## File
`execution/internal/engine/engine.go` — ~10 lines changed in `executeDecision`
