# Option A vs Option C — Which Is Better?

## Honest Comparison

| | Option A (Separate Symbols) | Option C (Per-Strategy Tracking) |
|---|---|---|
| **Code** | 0 lines | ~30 lines |
| **Fighting** | Eliminated | Eliminated |
| **Coverage** | 1 symbol per strategy | Both on BTC |
| **Risk** | If ETH doesn't move, mean-rev idle | More code, more bugs |
| **What you learn** | Which strategy works on which symbol | How both perform on same symbol |
| **Time to deploy** | Right now | Need to build + test |

## The Real Question

Not "which is better?" but "**do we even know if EITHER strategy is profitable yet?**"

We don't. 21 trades, -$0.05 — essentially flat. We've never let a single strategy run cleanly for more than an hour without interference.

## Recommendation

**Option A first. Option C later.** 

Option A tells you immediately: "does grid make money on BTC?" and "does mean-reversion make money on ETH?" After 2-4 hours, you have data. If one works, great. If both work, even better. If neither works, you know the approach needs rethinking.

Only after you have data should you invest 30 lines in position tracking (Option C) to let both strategies run on BTC together.

## The Counter-Argument

"The more trades, the better." True — Option C generates more trades by running both on BTC. But more trades from an unproven strategy = more potential losses. Prove the edge first, then scale.

## Files (Option A)

`config/config.yaml` — 2 lines: grid on BTC, mean-rev on ETH