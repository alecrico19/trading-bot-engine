# Prevent Buying Into Downtrends

## Current Protections

| Strategy | Protection | Covers |
|----------|-----------|--------|
| **Scalping** | 20-trade trend check (just added) | ✅ Blocks buys when 20-trade window is down |
| **Tick-momentum** | 50-trade warmup only | ❌ After warmup, can buy into any micro-uptick |
| **Mean-reversion** | Buys at RSI oversold (low prices) | ✅ Inherently avoids tops |

## The Gap: Tick-Momentum Has No Trend Filter

Tick-momentum enters on 5-tick uptrend signals (>55% upticks in 5-trade window). Even in a massive downtrend, 5 random upticks happen. After the 50-trade warmup, the bot can buy into these micro-bounces within a larger downtrend.

Same bug we just fixed for scalping. Tick-momentum needs the same fix.

## Fix: Add Trend Check to Tick-Momentum

Add a 50-trade broader window check (tick-momentum's warmup window size doubles as the trend window). Before entering a buy:

```go
longVals := hist.Values()
longWindow := longVals[len(longVals)-50:]  // broader trend
if broader trend is >55% upticks → "up"
if broader trend is <45% upticks → "down"

// Gate buy entries:
if longTrend == "down": return nil  // don't buy into downtrend
// Gate sell entries (if holding):
if longTrend == "up": return nil    // don't sell into sustained uptrend
```

## Also: Startup Bias From First 50 Trades

After the 50-trade warmup, determine the initial trend bias. Store it. Use it until contradicted by a new longer-term signal. This prevents the bot from entering a downtrend immediately after warmup.

## Files

`engine/tick_momentum.go` — add longTrend check on entry conditions (~15 lines)
Config: no changes needed (uses existing warmup window size)

## Expected

After fix: neither scalping nor tick-momentum buys into a downtrend. Both verify the broader trade history before entering. Mean-reversion continues buying dips (which is correct in downtrends — buying at lower prices).