# Options A + B + C Combined: Full Analysis

## What "All 3" Looks Like

| Option | What It Does | Lines | Risk |
|--------|-------------|-------|------|
| A | Disable scalping | 1 (config) | Zero |
| B | Scalp only when mean-rev says "range-bound" | ~30 | Medium — wrong regime detection = missed trades |
| C | Predictive SMA filter on order book ratio | ~25 | Medium — untested, might not fix edge |

Combined flow:
```
1. Mean-reversion checks market regime every 15s
2. If RSI between 30-70 AND BB not tight → market is ranging
3. Scalping enabled (only for this regime)
4. Scalping uses SMA(10) of order book ratio instead of instant value
5. SMA > 1.05 → buy, SMA < 0.95 → sell
6. If market not ranging → scalping disabled, mean-rev only
```

## Effort vs. Value

| Approach | Effort | Value | Risk |
|----------|--------|-------|------|
| Just A (disable scalping) | 1 line | Immediate stop to $-8/hr bleed | Zero |
| A + C (scalp with filter) | ~30 lines | Might fix edge, might not | Moderate |
| A + B + C (all three) | ~70 lines | Most sophisticated, most bugs | High |

## Honest Answer

We've spent all day trying to make scalping work. Reversed signals, disabled time-exit, lower thresholds, anti-accumulation — 261 trades later, still losing 100% on sells. **The 100ms order book ratio simply doesn't predict price movement.** Filtering it with an SMA won't change that — it just makes the noise smoother.

**My recommendation: Just A.** Disable scalping in config. Focus on making mean-reversion profitable. If mean-reversion proves it can make money over an overnight run, that's a win. If not, we know neither approach works at 5-minute candle resolution either, and we need entirely different data (1H candles, volume profile, order flow).

Options B and C are adding complexity to a broken signal — it's polishing a turd. Let's get one strategy working before we try to combine them.