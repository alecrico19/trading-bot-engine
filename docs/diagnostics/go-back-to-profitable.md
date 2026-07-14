# Why It Was Profitable Before

## The Profitable Setup (5:25 AM)
- **Tick-momentum only** (scalping disabled)
- 9 trades, +$0.05 P&L
- ~$0.0056 profit per trade
- 2s cooldown between signals
- 50-trade warmup before first entry

## What Changed

We added scalping back (trade-based version):
- Scalping: 1s cooldown, 5-tick window, 60% threshold
- More frequent trades = more opportunities for small losses
- Both strategies compete for the same signals on the same symbol
- Scalping's losses overwhelm tick-momentum's small profits

## Fix: Go Back to What Worked

Disable scalping. Keep tick-momentum + mean-reversion + mini profit exit.

```yaml
scalping:
  enabled: false
```

1 config line.

## Why Tick-Momentum Worked (When Scalping Didn't)

| | Tick-Momentum | Scalping |
|---|---|---|
| Cooldown | 2s | 1s |
| Warmup | 50 trades | 10 trades |
| Window | 5-50 trades | 5 trades |
| Exit hold check | Yes (must have position) | Yes |
| **Net: fewer, higher quality entries** | **= profit** | More, lower quality entries = loss |

Tick-momentum's longer warmup and cooldown naturally filter out noise. Scalping's faster cycle catches every micro-move — including the ones that reverse immediately.