# Aggressive Scalping Tuning

## What Makes the Bot Miss Opportunities

| Parameter | Current | Aggressive | Effect |
|-----------|---------|------------|--------|
| `order_book_depth` | 20 | 10 | Faster reaction, catches micro-imbalances |
| Scalping buy threshold | bid/ask > 1.5 | > 1.2 | Fires on weaker signals |
| Scalping sell threshold | bid/ask < 0.67 | < 0.83 | Same |
| `max_concurrent_positions` | 3 | 5 | More trades open at once |
| `max_position_pct` | 0.10 | 0.08 | Slightly smaller per-trade = more trades |
| `max_order_size_usd` | 200 | 500 | Allows bigger notional if equity grows |
| Per-symbol position tracking | No (blocked if any BTC long) | Yes (multiple entries per symbol) | Multiple entries on same symbol |
| Re-entry cooldown | None (100ms) | 2 second min between entries | Prevents overtrading on noise |

## Quick Wins (config only, no code)

Changes to `config.yaml`:
```yaml
strategies:
  scalping:
    order_book_depth: 10
    position_size_usd: 100
    enabled: true
    symbols: ["BTCUSDT", "ETHUSDT"]

risk:
  max_concurrent_positions: 5
  max_order_size_usd: 500
  max_position_pct: 0.08
```

And lower the scalping thresholds in `scalping.go`:
- Buy: ratio > 1.2 (was 1.5, strong: 3.0 → 2.0)
- Sell: ratio < 0.83 (was 0.67, strong: 0.33 → 0.5)

## Medium Effort: Per-Symbol Multiple Entries

Currently the engine blocks new entries if ANY position exists on that symbol (single entryPrice per symbol). Change to:
- Track positions per strategy (scalping can open a 2nd long even if mean-reversion already has one)
- Or: allow scalping to have multiple concurrent positions on the same symbol (up to 2)

~20 lines in engine.go

## Expected Result

| Metric | Current | Aggressive |
|--------|---------|------------|
| Trades per hour | ~5-10 | ~20-40 |
| Signals caught | Strong imbalances only | Moderate + strong |
| Risk | Low | Moderate (more frequent small losses) |
| Profit per trade | ~0.1-0.3% | ~0.05-0.15% (more volume compensates) |

## Recommendation

**Do both.** The config changes are 5 lines and instant. The per-symbol multiple entries (~20 lines) is the real unlock — without it, the scalping strategy blocks itself on BTC while holding a position, missing the next signal. With both, every 100ms order book update has a chance to trade even with existing positions. This is what those Facebook ads on gold are doing — they don't wait for the first position to close before entering the next one.
