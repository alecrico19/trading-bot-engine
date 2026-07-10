# Re-Enable Scalping

## Change
`config.yaml`: `scalping.enabled: false` → `enabled: true`

## What Happens
All three strategies run together:
- **Tick-momentum** — trades real tick data (just started working)
- **Mean-reversion** — RSI + BB on 5-min candles
- **Scalping** — order book imbalance at 100ms (was losing money)

## Risk
Scalping was losing -$8/hour earlier today. Re-enabling it alongside tick-momentum might drain profits from the other two strategies.

## Recommendation
Enable scalping but with position cap of 1 (max 1 scalping entry at a time) and smaller position size (5% vs 8%). This limits exposure while you evaluate whether the three strategies together perform better than tick-momentum alone.

## Files
`config/config.yaml` — 1 line