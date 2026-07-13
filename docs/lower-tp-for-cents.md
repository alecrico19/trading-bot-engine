# Why No Sell at +$60 BTC — Lower Take-Profit

## What Happened

Buy at $62,600 → price climbs to $62,660 (+$60, +0.096%). No sell fired because:

| Exit | Needs | Why Didn't Fire |
|------|-------|----------------|
| **Peak-drop** | Price peak THEN drop 0.02% | Price still climbing — no drop yet |
| **Momentum reversal** | 40%+ downticks | All ticks are UP — no reversal |
| **Take-profit** | 1% ($626) | $60 is only 0.096% — nowhere close |
| **Stop-loss** | Loss only | Price is UP — irrelevant |

The bot has NO exit for "I'm up $60 on a $40 trade, take the money." Every exit requires either a reversal (peak-drop, momentum) or a distant target (1-2%).

## Fix: Lower Take-Profit to Catch $60 Moves

You said you don't mind profiting cents. Lower take-profit to levels that match the moves you actually see:

| Exit | Current | Proposed | BTC $ at $62,600 |
|------|---------|----------|-------------------|
| **Partial (50%)** | 1% | 0.1% | ~$63 |
| **Full (100%)** | 2% | 0.3% | ~$188 |

At 0.1%: your $62,600 → $62,660 move triggers a 50% partial sell. You lock in $0.04 profit on a $40 position. The remaining 50% rides for more.

This fills the gap between "peak-drop catches the reversal" and "momentum catches the trend change." Now "price just went up enough" is also an exit.

## Files
`config/config.yaml` — 2 lines: `take_profit_1r_pct: 0.01 → 0.001`, `take_profit_target_pct: 0.02 → 0.003`
