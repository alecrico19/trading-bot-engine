# Fix: Positions showing 0 despite active buys

## Bug
USDT is deducted on buys (visible), but BTC/ETH holdings show 0.

## Root Cause
`splitSymbol("BTCUSDT")` returns `("BTCUSDT", "USDT")` — base is the full symbol, not "BTC".
When `adjustBalance("BTCUSDT", amount)` creates a new balance entry, `Asset` field is empty (zero-value).
`FetchBalance` filters out empty Asset strings — hides the holding.

## Fix
In `adjustBalance`, set Asset when creating a new balance entry:
```go
if b.Asset == "" {
    b.Asset = asset
}
```

## File
`exchange/paper.go` — 3 lines
