# Bot Stops Trading After Few Trades — Diagnosis

## Root Cause: 200 EMA Trend Filter Too Strict

The 200 EMA on 100ms order book updates only spans 20 seconds of data — not a real higher-timeframe trend. After 100 ticks (~10s), if price drops below this short EMA, ALL buys are blocked. If price rises above, ALL sells are blocked. The bot gets stuck in one direction within seconds.

The pros use 200 EMA on **daily/hourly** charts. Our 100ms data frequency makes 200 EMA meaningless — it's just the average price over the last 20 seconds.

## Fix

### 1. Use a higher timeframe for EMA (correct approach)
Instead of computing EMA from 100ms order book ticks, fetch the 1-hour or 15-minute candle from Binance and compute a 200 EMA from those. This matches the pro scalper methodology.

~30 lines: add a FetchKlines call in the engine, pass EMA value to MarketState.

### 2. Or reduce EMA period and increase warmup (quick fix)
- Change EMA period from 200 to 50 (meaningful on 100ms data)
- Require 200 data points before activation (instead of 100)
- This gives ~20 seconds of warmup before the filter kicks in

~5 lines in scalping.go

### 3. Or make trend filter informational, not blocking (safest)
Don't block trades based on trend. Instead, reduce position size when trading against the trend (e.g., 50% smaller). This lets the bot keep trading but reduces risk on counter-trend entries.

~10 lines in scalping.go

## Other Issues Found

### P&L still shows $0.00 on sells
Portfolio $840, Cash $840 — the positions-fix (adjustBalance Asset field) might not be in the running engine. User needs to restart with latest code.

### Holdings not tracked
Portfolio equals Cash exactly, meaning no BTC/ETH holdings are being counted. The running engine likely predates the positions fix.

## Recommendation
**Fix #1 (correct approach) — fetch 1H klines for 200 EMA.** This is what the pros actually do. Quick fix #2 is a band-aid. The underlying issue: 200 EMA on 100ms data is noise, not trend.

But first: **restart with latest code.** The running engine predates several fixes (positions-fix, zero-trade-fix). Those changes need to be live before diagnosing further.