# Bot Not Trading in Range — Fix

## Diagnosis

BTC oscillating $63,990-$64,050. Tick-momentum sees mixed upticks/downticks (~50/50). The 60% threshold never crosses. Bot holds positions it never exits.

## Quick Fix: Lower thresholds

| Parameter | Current | New | Effect |
|-----------|---------|-----|--------|
| Entry (buy) | 60% upticks | 55% | More sensitive in mixed markets |
| Exit (sell) | ≤50% upticks | ≤45% | Sells on slight downtick bias |

## Better Fix: Range detection

If the market is in a range (price staying within 1% of entry for >60s):
- Buy at the bottom of the range
- Sell at the top
- Don't wait for momentum that never comes

This is mean-reversion on the tick level, not candle level. It directly catches the $63,990-$64,050 oscillations.

## Also Check: Mean-reversion warmup

Mean-reversion needs ~25 minutes of candle data. Has it been that long since the bot started? If so, it should be trading. Check for mean-reversion trades:

```bash
grep "mean.reversion\|mean-rev" /tmp/trading-bot-logs/engine.log
```

## Recommendation

**Quick fix first:** lower thresholds to 55%/45%. See if trades start flowing in the range. If not, implement range detection for tick-level mean-reversion.

## Files
- `tick_momentum.go` — 60% → 55%, 50% → 45% (~4 lines)