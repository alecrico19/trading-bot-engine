# Make Bot More Aggressive — Honest Assessment

## Current State
- Tick-momentum + mean-reversion (scalping disabled)
- 8% position size ($40 on $500)
- Tight stop 0.1%, mini-profit at 0.05%
- Time-exit disabled (just now)

## Aggression Levers

| Lever | Current | Aggressive | Impact |
|-------|---------|------------|--------|
| Position size | 8% ($40) | 20% ($100) | 2.5x bigger wins & losses |
| Max concurrent | 5 | 5 | Already maxed |
| Entry threshold | 60% upticks | 50% | More trades, more noise |
| Cooldown | 2s | 1s | Faster signals, more trades |
| Warmup | 50 trades | 10 trades | Faster start, worse entries |
| Short entries | No | Yes | Double opportunity, double risk |
| Anti-accumulation | Yes | Remove | More capital deployed, more risk |

## Honest Answer

Making the bot more aggressive without confirming it has an edge just means **losing money faster**. The tick-momentum strategy was briefly profitable (+$0.05) on its own before we started layering on fixes. We've been changing things so fast we don't know what actually works.

## My Recommendation

**Stop changing things. Pick ONE configuration and let it run for 1+ hours.** 

The setup that showed promise:
```
tick-momentum: enabled, 60% threshold, 2s cooldown, 50-trade warmup
mean-reversion: enabled
scalping: disabled
mini-profit: enabled (0.05%, fixed loop)
tight stop: 0.1%
time-exit: disabled
```

**After 1 hour of stable running, check P&L.** If profitable, THEN raise position size to 20%. One change at a time, measured, not guessed.

The bot doesn't need more aggression — it needs stability and data. We've pushed 20+ commits today. Let it run.
