# Fix: Zero-amount trades in history after restart

## Bug
On fresh start, scalping strategy sends SELL signals. Paper trader rejects them (no holdings). Engine records the rejected trades anyway — showing $0.00, qty 0.0000 in history.

## Root Cause
`executeDecision` doesn't check `order.Status` after `PlaceOrder`. Rejected/cancelled orders get recorded in DB and risk manager as if they filled.

## Fix
In `engine.go:executeDecision`, after `PlaceOrder`:
```go
if order.Status != types.StatusFilled && order.Status != types.StatusPartiallyFilled {
    return  // don't record rejected/unfilled orders
}
```

## File
`execution/internal/engine/engine.go` — +3 lines in executeDecision
