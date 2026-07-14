# Integrate Professional Scalping Research Into Bot

## Source
Analysis of 15 YouTube scalping videos (25M+ views) from Zeus automation system.

## What to Integrate

### 1. Trend Filter (Critical — all 15 agree)
Add 200 EMA check. Scalping only buys above 200 EMA, only sells below. Prevents trading against the trend. ~10 lines in scalping.go.

**Before:** Order book imbalance → trade either direction.
**After:** Order book imbalance + price above 200 EMA → buy. Below → sell. Neutral → skip.

### 2. Take-Profit Targets (Replace Signal-Only Exits)
Currently exits only on reverse signal. Change to: exit 50% at 1% profit, move stop to breakeven, let rest run to next reverse signal. ~15 lines in engine.go.

### 3. MACD Zero-Line Strategy (New Strategy)
86% win rate strategy from #1 video on the list:
- MACD crosses above signal + both below zero → BUY
- MACD crosses below signal + both above zero → SELL
- Requires 200 EMA filter (price above 200 EMA for longs)
This is a NEW strategy, complementary to scalping.

### 4. Opening Range Breakout (Optional)
First 15-min candle of session defines range. Breakout with close back inside = mean reversion entry. Simple addition that the Data Trader video backtested at 72% win rate.

## Recommendation
**Build #1 (trend filter) and #2 (take-profit) now.** They're quick wins that fix the #1 criticism from professional scalpers — our bot trades against the trend. The MACD strategy (#3) can follow as a third strategy after verifying the trend filter improves results.

### Files
| Item | File | Lines |
|------|------|-------|
| Trend filter | `scalping.go` | ~10 |
| Take-profit targets | `engine.go` | ~15 |
| Config hooks | `config.yaml` | ~5 |

~30 lines total.
