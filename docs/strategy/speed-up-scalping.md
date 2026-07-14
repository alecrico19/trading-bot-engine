# Speed Up Scalping — 10x More Trades

## Current State
~2-4 trades per minute. Scalping should be 20-40 per minute.

## Bottlenecks

| Limit | Current | Why It's Slow | Fix |
|-------|---------|---------------|-----|
| Buy threshold | ratio > 1.2 | Only fires on clear imbalance | > 1.05 (much more sensitive) |
| Sell threshold | ratio < 0.83 | Same | < 0.95 (more sensitive) |
| Anti-accumulation | max 3 buys before sell | Blocks entries | Remove or raise to 10 |
| Position check | no sells without position | Requires buy-before-sell | Allow shorts (sell first, buy back later) |
| One trade per signal | single BUY or SELL | Doesn't scale | Multiple position scaling (1/3, 2/3, full) |
| Spread filter | blocks if spread > 0.1% | BTC/ETH spreads are 0.01-0.02%, not blocking | Fine as-is |

## Quick Fixes (config only, ~5 lines changes)

1. Lower scalping thresholds: 1.2→1.05, 0.83→0.95, 2.0→1.5, 0.5→0.8
2. Remove anti-accumulation cap (raise to 10)
3. Allow sells even without position (open shorts)

## Medium Fix: Multi-Entry Scaling (~20 lines)

Instead of one BUY at 100% position size, split into 3 entries:
- First entry: 40% at ratio > 1.05
- Second entry: 30% at ratio > 1.10
- Third entry: 30% at ratio > 1.20

This catches more entries at different thresholds and increases trade frequency 3x.

## Recommendation
Do the quick fixes (config + threshold changes) immediately. The multi-entry scaling is a bigger change worth doing after verifying the quick fixes work. Expected result: 10-20 trades per minute instead of 2-4.