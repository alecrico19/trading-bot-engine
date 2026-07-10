# P&L Still $0 + Bogus BTC Sell — Diagnosis

## Symptom 1: P&L $0.00 on all sells

The trade history you're seeing is from the PREVIOUS engine run (7:13-7:20 PM). That engine predates the diagnostic logging and entry blending fixes (committed at 076060f, dc5fa72).

The engine was restarted at 7:22 PM (I see the fresh start in the log). The new engine has all fixes. It just hasn't traded yet since restart.

## Symptom 2: BTC sell at $61,159.60

This is likely the paper trader's `FetchTicker` returning a stale/corrupted order book price. The engine calls `e.marketData.FetchTicker()` in `checkStopLoss` which can return wrong prices if:
- The Binance WS order book cache was replaced mid-read
- A Binance API rate limit caused a delayed response with old data
- The `parseFloat` function returned 0 on a broken string

## Fix Plan

### Fix 1: Verify P&L on next restart
The new engine has diagnostic logging. Wait for a sell to happen, then check:
```
grep "trade P&L" /tmp/trading-bot-logs/engine.log
```
This will show the actual `entry`, `avg_price`, `filled`, and `pnl` values.

### Fix 2: Sanity-check ticker prices
Add a guard in `checkStopLoss`: if the returned ticker price deviates more than 10% from the last known price, skip the trade and log a warning. This prevents bogus fills like $61,159 on BTC.

### Fix 3: The sell at $61,159 — was it actually a trade?
Check the trade journal to confirm it was a real fill, not a display artifact. If the DB shows qty=0.0012 @ $61,159, it was a real paper fill — likely triggered by a bad ticker.

## Recommendation
Do Fix 2 (sanity-check ticker prices) now — ~5 lines. It prevents the bogus fill scenario. The P&L issue should resolve on the next sell since the latest code is running.