# Time-Exit Selling Low — Root Cause & Fix

## The Data

All exits are time-exit at near-breakeven losses:
```
BTC: -0.02% (-$13)  → closed at 90s
ETH: -0.02% (-$3)   → closed at 90s
BTC: -0.004% (-$2)  → closed at 90s
```

Tick-momentum's sell signal (momentum reversal) never fires before the time-exit. The bot enters on upticks, price drifts slightly against it, and 90 seconds later the time-exit closes the entire position at a tiny loss.

## The Problem

Time-exit fires at ANY loss (pnlPct < 0), even -$2 (0.004%). These aren't "bad trades" — they're trades that need more than 90s to develop. Tick-momentum should be the one deciding when to exit, not a blind timer.

## Options

### A: Disable time-exit again (back to momentum-only exits)
Tick-momentum + momentum-based exits. Position stays open until momentum reverses. Pro: lets trades develop. Con: if momentum never reverses while position bleeds, you get the 10-minute bleed again.

### B: Increase time-exit to 300s (5 min) + only on significant losses
Fire only if pnlPct < -0.05% (meaningful loss, not noise). Gives trades time to breathe.

### C: Fix tick-momentum's sell signal sensitivity
Currently exits at upPct <= 0.40 (highly consistent downticks). Lower to <= 0.50 (any reversal). Exits happen faster, before the time-exit.

## Recommendation: Do A + C

Disable time-exit. Lower tick-momentum exit threshold from 0.40 to 0.50. Positions stay open until momentum truly reverses, not a blind 90s timer.

## Files
- `engine.go` — disable time-exit (change `false &&` back, or just remove the block)
- `tick_momentum.go` — `upPct <= 0.40` → `upPct <= 0.50` on sell conditions