# More Aggressive Capital Deployment

## Current State
- max_position_pct: 8% → $40 per trade on $500
- max_concurrent_positions: 5
- Max deployed at once: $200 (40% of $500)
- 60% of capital idle

## Options

### Quick: Just raise the numbers (config only)
```yaml
risk:
  max_position_pct: 0.25      # $125 per trade (was 8%)
  max_concurrent_positions: 3  # 3 × $125 = $375 deployed
  max_order_size_usd: 500
```
Deploys 75% of capital. 1 line change.

### Aggressive: Fixed dollar sizing
```yaml
risk:
  max_position_pct: 0.50      # $250 cap
  max_concurrent_positions: 3
  max_order_size_usd: 200     # per-trade cap
```
$150-200 per trade, 3 concurrent = $450-600 deployed (near 100%).

### Scale with profit: Kelly-based
Use the Kelly criterion: bet size = edge / odds. Automatically scales position size based on strategy win rate.
Requires code change (~10 lines).

## Recommendation

**Just raise the numbers.** 25% per trade, 3 concurrent = 75% deployed. No code changes needed. If scalping + tick-momentum prove profitable at this size, you can go higher. If they lose, the larger positions amplify losses — but you're on paper, so losses are educational.

## Files
`config/config.yaml` — 2 lines