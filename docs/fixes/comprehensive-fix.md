# Bot Stops Trading — Comprehensive Diagnosis

## NOT just the EMA. Multiple bugs found:

## Bug 1: splitSymbol doesn't handle "BTCUSDT" format (CRITICAL)

`splitSymbol("BTCUSDT")` returns `("BTCUSDT", "USDT")` — base is the full symbol, not "BTC".

**Cascade of failures:**
- Crypto stored under asset name "BTCUSDT" instead of "BTC" ✓ confirmed in holdings
- `GetEquity()` fetches ticker for "BTCUSDT" + "USDT" = "BTCUSDTUSDT" → fails → counts as $0 ✓ Portfolio = Cash = $840
- `checkStopLoss` computes `base := h.Asset + "USDT"` = "BTCUSDTUSDT" → never matches symbol "BTCUSDT" → stop-loss NEVER fires
- Take-profit NEVER fires (same reason)
- Dashboard shows "BTCUSDT" as a position instead of "BTC"

**Fix:** Rewrite `splitSymbol` to handle both "BTC/USDT" and "BTCUSDT" formats:
```go
func splitSymbol(symbol string) (base, quote string) {
    if strings.Contains(symbol, "/") {
        parts := strings.Split(symbol, "/")
        return parts[0], parts[1]
    }
    // Try common quote currencies
    for _, q := range []string{"USDT", "USDC", "USD", "BTC", "ETH", "BNB"} {
        if strings.HasSuffix(symbol, q) {
            return symbol[:len(symbol)-len(q)], q
        }
    }
    return symbol, "USDT"
}
```

## Bug 2: 200 EMA on 100ms data is noise (HIGH)

200 EMA on 100ms updates = 20-second average. After 10s warmup, the filter blocks one direction permanently. Bot gets stuck.

**Fix:** Use 1-hour candle data from Binance (FetchKlines) for the 200 EMA, not 100ms ticks. ~30 lines.

Quick alternative: reduce EMA period to 50 and warmup to 200 points.

## Bug 3: P&L on sells is $0.000000 (HIGH)

Sell at $64409.99 after buy at $64410.02 → P&L should be -$0.0000343. Shows $0.000000.

**Cause:** `entryPrice` tracking overlaps with the trend filter blocking sells. When the sell finally happens, the entryPrice might be from a different buy cycle or was reset.

Also: the `executeDecision` P&L calc checks `order.Filled >= order.Amount*0.99` before resetting entryPrice. But `order.Filled` might be 0 if the sell was rejected (insufficient balance — because the base asset is stored as "BTCUSDT" not "BTC").

**Fix:** Fix splitSymbol (Bug 1) first — it enables proper balance tracking, which enables proper sells, which enables proper P&L.

## Bug 4: ETH never sold — accumulates with no exit (HIGH)

4 ETH buys, 0 sells. Trend filter blocks sells above 200 EMA. But also: the sell might fail because `adjustBalance("ETHUSDT", -order.Amount)` deducts from "ETHUSDT" balance, which exists (from Bug 1). But the sell itself is blocked by the trend filter.

**Fix:** Fix Bug 2 (EMA source data) — allows sells when appropriate.

## Bug 5: Stop-loss and take-profit never fire (HIGH)

`checkStopLoss` iterates holdings, computes `base := h.Asset + "USDT"`. With `h.Asset = "BTCUSDT"`, base = "BTCUSDTUSDT" ≠ "BTCUSDT" → never matches → goroutine finds nothing → stop-loss/take-profit dead code.

**Fix:** Fix Bug 1 (splitSymbol) — correct asset names → stop-loss/take-profit work.

## Fix Priority

| # | Bug | Fix | Lines |
|---|-----|-----|-------|
| 1 | splitSymbol | Rewrite to handle "BTCUSDT" | ~10 |
| 2 | 200 EMA noise | Use 1H klines instead of 100ms ticks | ~30 |
| 3 | P&L $0 | Fixed by Bug 1 fix (enables proper sells) | 0 |
| 4 | ETH never sold | Fixed by Bug 2 fix (allows sells) | 0 |
| 5 | Stop-loss dead | Fixed by Bug 1 fix (correct asset names) | 0 |

**Total: ~40 lines across 2 files.**

## What to fix
1. `exchange/paper.go` — rewrite `splitSymbol`
2. `engine/scalping.go` — fetch 1H klines for 200 EMA instead of using 100ms ticks
3. Restart launcher