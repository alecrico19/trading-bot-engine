# Is Removing Exits the Best Move? — Honest Answer

## No. It's a band-aid.

Removing exits helps the bot hold through noise and catch the $200 BTC move. But it doesn't fix why the bot enters at bad prices. The momentum-based entries (scalping, tick-momentum) buy AFTER upticks start — by the time the signal confirms, the move is already underway. The bot buys near the top and then the micro-pullback triggers a sell.

## What Actually Needs Fixing: Entry Quality

| Strategy | Entry Signal | Why It Loses |
|----------|-------------|-------------|
| Scalping | 5-tick uptick (60%+) | Buys AFTER price already went up |
| Tick-momentum | 5-tick uptick (55%+) | Same — momentum already happened |
| Mean-reversion | RSI < 30 + BB lower | Buys at LOWS (mathematically correct) |

Only mean-reversion buys at good prices. But it fired ONCE all day.

## The Real Best Move

**Make mean-reversion the primary entry. Use momentum strategies only for exit timing.**

```
Flow:
1. Mean-reversion detects oversold → BUY (at a GOOD price)
2. Scalping confirms the uptrend is actually happening → HOLD
3. Momentum reverses → SELL (exit on confirmed reversal, not micro-noise)
4. Take-profit at 0.3% catches the top
```

Current flow (wrong):
```
1. Momentum says buy → BUY (at a bad price)
2. Micro-reversal → SELL (at a loss)
3. Repeat × 19
```

## What This Means For Code

1. **Lower mean-reversion thresholds** — RSI < 35 instead of < 30, BB dev 1.5 instead of 2.0. Makes it fire more often.
2. **Use scalping/tick-momentum only for exit confirmation** — don't let them open new entries. Only let them close positions that mean-reversion opened.
3. **Keep take-profit at 0.15%/0.3%** — lock in gains when they come.
4. **Remove peak-drop** — too aggressive for entries that are actually well-timed.

## Expected Effect

Instead of 19 losing micro-trades, you get 2-3 well-timed entries at oversold levels, held through micro-reversals, exited at +0.3% profit. Fewer trades, more profit per trade, less fee drain.

## Files

- `config.yaml` — lower mean-reversion thresholds (RSI 30→35, BB 2.0→1.5)
- `config.yaml` — disable scalping/tick-momentum as entry sources, keep as exit-only
- `engine.go` — remove peak-drop (1 line)