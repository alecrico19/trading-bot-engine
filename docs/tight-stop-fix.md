# Tight Stop Needed — 0.1% Not 5%

## What Happened

Buy at $64,242 → price drops for 10 minutes → sell at $64,178 (-$64 loss)

Stop-loss at 5% = $3,210 away. A $64 drop is invisible to it. The position bled for 10 minutes with no exit guard.

## Fix: Add Tight Stop at 0.1%

Replace the current stop-loss at 5% with two tiers:
- **Tight stop: -0.1%** ($64 on BTC) — fires within seconds of adverse move
- **Hard stop: -5%** (kept as safety net)

Or just lower the existing stop_loss_pct from 0.05 (5%) to 0.001 (0.1%).

## Also: Re-enable Time-Exit for Losers

If a position has been open >90s AND is at a loss → market sell. Gets out of losing trades fast. The disabled time-exit only fired for winners (pnlPct > 0) — reverse it for losers.

## Files
- `config.yaml` — `stop_loss_pct: 0.001` (1 line)
- `engine.go` — time-exit for losers (~5 lines)

## Expected
This trade would have been stopped at -$0.06 (0.1%) instead of -$0.04 on a partial fill at $64,178.