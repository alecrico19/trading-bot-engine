# Sells at $0.00 — Fix

## Root Cause
Time-exit and take-profit place market SELL orders with `price=0`. The paper trader fills at the best bid from its cached order book. If the cache is empty/stale when the exit fires, `fillPrice` stays at 0. The order gets marked as `StatusFilled` with `AvgPrice=0` and `Price=0`.

This happens because:
1. `forwardOrderBook` populates the paper trader cache from Binance WS
2. Time-exit fires 30s after entry
3. If the cache wasn't populated (WS delay, reconnect, etc.), fillPrice=0
4. Order fills at $0.00 — recorded in DB with price=0, pnl=0

## Fix

### In `simulateFill`: fallback to entry price if order book empty
```go
if !ok || len(ob.Bids) == 0 {
    // No order book — use a fallback price
    // For paper mode, use the order's price if set
    if order.Price > 0 {
        fillPrice = order.Price
    }
    // Add ticker fallback later
}
```

### Better: add ticker fallback
Pass the entry price or current market price as a fallback when placing market orders:
```go
// In checkStopLoss, pass entry price as fallback for market orders
e.orderMgr.PlaceOrder(ctx, symbol, types.SideSell, types.TypeMarket, amount, entry*0.999, "time-exit")
```

Using `entry*0.999` as the limit price (even for market orders) gives the paper trader a fallback fill price.

## Files
`engine.go` — change PlaceOrder price from 0 to `entry` for exit trades (~4 line changes)

## Expected
After fix: SELL shows actual fill price (e.g., $64,350) and P&L shows actual gain/loss.