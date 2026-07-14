# Is Shorting the Best Move? — Honest Assessment

## The Situation

- BTC in downtrend ($62,283, falling from $62,800+ earlier)
- Mean-reversion buys dips → loses in falling market
- 15 trades, -$0.53 P&L

## Options

### Option A: Short in downtrend
Flip mean-reversion: sell at RSI overbought, cover at RSI oversold. Profitable in downtrends, but loses in uptrends. Same problem, opposite direction. Requires knowing when to flip.

### Option B: Add market regime detection
Use 1H candle data to determine: uptrend, downtrend, or range.
- Uptrend → buy dips (current mean-reversion)
- Downtrend → sell rips (flipped mean-reversion)
- Range → buy dips + sell rips (bidirectional)

This is the RIGHT solution but takes ~50 lines to implement.

### Option C: Do nothing
Let current setup run. BTC might reverse — mean-reversion buys at the eventual bottom and profits on the recovery. The strategy works in the right market.

## Honest Answer

**Option C for now. Option B is the correct long-term fix.**

BTC could reverse at any moment. If we flip to short now and BTC reverses, we've just swapped one losing direction for another. The regime detector (Option B) is the real solution — adapt to whatever the market is doing.

## If You Want Action Now

Lower the mean-reversion exit thresholds so it sells FASTER on any recovery, limiting losses in downtrends:
- `rsi_overbought: 65 → 55` (exit sooner)
- `bb_stddev: 1.5 → 1.2` (tighter bands, faster exits)

This doesn't flip the strategy — it just makes it quicker to exit when wrong. Less time in the market = less exposure to the downtrend.

## Files for Quick Fix
`config/config.yaml` — 2 lines: rsi_overbought 65→55, bb_stddev 1.5→1.2
