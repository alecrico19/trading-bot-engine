# Bot Performance — Current Session

## Summary
- **15 trades, -$0.53 P&L** (mean-reversion only)
- **Portfolio: $998.28** (started $1,000)
- **BTC/USDT: $62,283 | ETH/USDT: $1,773.83**

## Strategy Breakdown

| Strategy | P&L | Trades | Status |
|----------|-----|--------|--------|
| **Mean-reversion** | -$0.38 | 14 | Buying dips in falling market |
| **Stop-loss** | -$0.08 | 1 | Safety net fired |
| **Take-profit-50** | -$0.04 | 1 | ✅ Fix working — only ONE fire! |

## What's Working

✅ Take-profit-50 fires only once (fix confirmed)
✅ Mean-reversion entering on RSI oversold as designed
✅ Bot actively trading on BTC

## What's Not Working

❌ BTC in sustained downtrend — mean-reversion buys dips that keep dipping
❌ ETH barely trading (1 sell total) — not enough RSI extremes
❌ P&L still negative — -$0.53 on 15 trades

## Assessment

The issue isn't the code — it's the market. Mean-reversion buys at RSI oversold expecting a bounce. But BTC is in a FALLING market — $62,283 now vs $62,272 earlier entries. Every oversold signal gets followed by more selling. Mean-reversion needs a RANGING market to be profitable — not a trending one.

## What This Means

In a trending market, the only viable strategies are:
1. **Trend-following** (which we disabled — buys at upticks, but also loses on reversals)
2. **Stay out** (no trades in falling markets)
3. **Only trade the trend direction** (only sell/short in downtrends)

Currently: mean-reversion = buy dips in downtrend = guaranteed losses until market reverses.
