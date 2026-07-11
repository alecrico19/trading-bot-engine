# Why Bot Still Buys High, Sells Low (Diagnosis)

## The Pattern (Same As Before)

```
Price at $64,300 → starts rising
Price at $64,310 → 3 of last 5 ticks up → scalping: "BUY!"
Price at $64,320 → still going up
Price at $64,315 → 1 downtick → momentum mixed
Price at $64,305 → 3 of 5 down → scalping: "SELL!"

Result: Bought $64,310, sold $64,305 = -$5 loss
```

The bot REACTS to what already happened, never PREDICTS. By the time 60% of recent ticks are up, the move is mostly over.

## What Each Strategy Actually Does

| Strategy | Enters When | Exits When | Net Effect |
|----------|------------|-----------|------------|
| Scalping | 60% ticks UP | 60% ticks DOWN | Buy high, sell low |
| Tick-momentum | 60% ticks UP | 60% ticks DOWN | Buy high, sell low |
| Mean-reversion | RSI < 30 + BB lower | RSI > 70 + BB upper | Buy low, sell high (theoretically) |

Mean-reversion is the ONLY strategy designed to buy low and sell high. But it needs 25+ minutes of candle data to warm up and has 0 trades so far.

## What Would Actually Work

### Option A: Disable momentum strategies, only mean-reversion
Mean-reversion buys at RSI oversold (price is LOW) and sells at RSI overbought (price is HIGH). This is the mathematical opposite of momentum-following. Backtest showed 63% WR, 1.21 PF.

### Option B: Reverse the momentum logic (mean-revert the ticks)
Instead of: "ticks going up → buy" (momentum)
Do: "ticks going up → sell" (mean revert)
And: "ticks going down → buy" (mean revert)

This is what we tried before with the reversed scalping signals. It worked better than the original but still lost money because tick-level mean reversion is unreliable.

### Option C: Add predictive indicator
Don't just look at "are ticks going up?" — look at "is the RATE of upticks accelerating?" This requires tracking tick direction over multiple windows and detecting acceleration/deceleration.

## Recommendation

**Option A.** Scalping and tick-momentum are both momentum-following on short timeframes — they will ALWAYS buy high and sell low because the signal arrives AFTER the move. Mean-reversion is the only strategy that mathematically buys low and sells high. Disable momentum, let mean-reversion trade. If it also loses money after 2+ hours of warmup, we know the market is too efficient for any automated strategy at retail latency.