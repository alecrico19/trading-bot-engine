# Lower Take-Profit Targets — Assessment

## Current State

Take-profit at 1-2% never fires. On BTC at $63,000, that's $630-1260. Our trades last seconds to minutes, not hours to days.

## What Actually Fires Now

| Exit | Target | BTC $ value | Fires? |
|------|--------|-------------|--------|
| **Peak-drop** | 0.02% from peak | $12.50 | ✅ Often |
| **Momentum reversal** | Tick-based | N/A | ✅ Often |
| **Stop-loss** | 0.5% loss | $315 | Rarely |
| **Take-profit** | 1-2% gain | $630-1260 | ❌ Never |

There's a gap between peak-drop (0.02%) and stop-loss (0.5%). Take-profit at 0.15-0.3% would fill that gap.

## Recommendation

Lower take-profit to match our timeframe:

| Exit | Current | Proposed | BTC $ |
|------|---------|----------|-------|
| **Partial (50%)** | 1% | 0.15% | ~$95 |
| **Full (100%)** | 2% | 0.3% | ~$189 |

This bridges the gap between peak-drop catching micro-moves and momentum-only exits. The 50% partial sell at 0.15% locks in a confirmed gain. The 100% sell at 0.3% takes the rest off the table.

## Files
`config.yaml` — 2 lines