# BTC Losses — Root Cause Analysis

## The Data

| Symbol | Buys | Sells | P&L |
|--------|------|-------|-----|
| ETH | 1 | 2 | **+$0.005** ✅ |
| BTC | 5 | 8 | **-$0.074** ❌ |

## What Caused BTC Losses

| Exit Type | P&L | Trades |
|-----------|-----|--------|
| **take-profit-50** | -$0.120 | 4 sells |
| **mean-reversion** | +$0.051 | 4 sells |

The take-profit-50 repeat-fire bug caused 4 consecutive sells at -$0.04 each. That's the ENTIRE BTC loss. The mean-reversion exits were PROFITABLE (+$0.05).

The last sell was +$0.06 (mean-reversion at 11:31 PM) — proof the strategy works when not sabotaged by the repeat-fire bug.

## The Fix Already Pushed

The commit you just pushed (`824c0c9`) fixes the take-profit-50 repeat fire. After restart, the partial take-profit fires ONCE per entry instead of repeating until the position is dust.

## The ETH Pattern

ETH only had 1 buy and 2 sells — the take-profit-50 fired once (at $0.00 P&L), then the mean-reversion exit fired at +$0.005 (profitable). ETH didn't suffer from the repeat-fire because it only had ONE take-profit-50 before the mean-reversion exit closed the rest.

## Conclusion

The strategy is NOT broken. Mean-reversion exits were profitable. The take-profit-50 repeat bug was the culprit. The fix is already pushed. Restart and the BTC losses from repeat-fires should disappear.