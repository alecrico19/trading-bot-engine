# Why So Few Signals? — Analysis

## Expected vs Actual

| Strategy | Expected Trades/Hour | Why |
|----------|---------------------|-----|
| Tick-momentum | 5-15 | 55% of 5-tick windows should fire frequently on BTC |
| Mean-reversion | 0-3 | Needs RSI extremes + BB extremes (rare) |

## Why Tick-Momentum Goes Silent

Once the bot buys BTC and holds a position, it can ONLY sell — the position check blocks new buys. If the sell threshold (45% downticks) isn't hit while the position is open, the bot sits idle. And if the position is small (already partially sold), further trades are blocked.

The cycle:
```
1. Upticks detected → BUY (holds position)
2. Now blocked from buying again
3. Waiting for downticks → might take minutes or never come
4. Position sits idle, no new entries possible
```

Tick-momentum was designed for trend-following: buy on uptrend, sell on reversal. But if the trend continues up without reversal, the bot can't buy more (one position max per symbol). If the trend is flat, it never sells. If the market oscillates between 45-55%, neither condition fires.

## Is This Normal?

**For mean-reversion: yes.** It's a low-frequency strategy by design.

**For tick-momentum: partially.** BTC has thousands of trades/min, so signals should fire every few minutes. But the one-position-per-symbol limit means once a position is open, no new entries until it closes.

## How To Increase Signal Frequency

1. **Allow multiple entries per symbol** — raise anti-accumulation cap, or allow scalping entries alongside tick-momentum entries
2. **Lower thresholds further** — 52% buy, 48% sell would fire more often
3. **Add scalping back** — faster cycle, catches micro-moves tick-momentum misses
4. **Reduce cooldown** — 2s → 1s for tick-momentum
5. **Allow both buy and sell signals independently** — buy signals even with a position, sell signals even without

## Recommendation

**Option 3: Bring back scalping on trade data.** It was disabled when it was losing money, but that was before we fixed the entry price sanity check, raised the stop to 0.5%, and removed the mini-profit/time-exit interference. Try it again in the cleaner environment.