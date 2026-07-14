# Don't Sell All Positions — Per-Strategy Position Tracking

## The Problem

Mean-reversion's exit sells `p.Amount` — the FULL position for that symbol. If grid bought 0.0004 BTC and mean-reversion also bought 0.0013 BTC, the exit sells 0.0017 BTC total — closing BOTH strategies' positions.

```
Grid buys 0.0004 BTC
Mean-rev buys 0.0013 BTC
Total: 0.0017 BTC held

Mean-rev exit fires → sells 0.0017 BTC (closes everything, including grid's)
```

Grid's position was wiped before it could hit take-profit. BTC then went up $48 — profit missed.

## The Fix

Either:
1. **Separate per-strategy position tracking** — mean-reversion only sells its own entries. Grid positions managed by take-profit/stop-loss.
2. **Only one strategy per symbol** — simpler: grid trades BTC, mean-reversion trades ETH (no overlap).
3. **Partial position tracking** — positions tagged with strategy name. Exit only closes the tagged portion.

## Simplest Fix: Option 2

```yaml
grid:
  symbols: ["BTCUSDT"]       # Grid only buys BTC
mean-reversion:
  symbols: ["ETHUSDT"]       # Mean-rev only trades ETH
```

Zero code changes. No more fighting. Grid buys BTC every 2 hours, mean-rev trades ETH on its own. Combined with separating grid from the trade stream loop.

## Files

`config.yaml` — 2 symbol changes (BTC for grid, ETH for mean-rev)