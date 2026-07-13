# Stop-Loss Too Tight + Bogus Prices

## Problem 1: 0.1% stop killing mean-reversion entries

Mean-reversion buys at RSI oversold (price is LOW). The price often dips another 0.1% before recovering. The tight stop kills the trade at -0.1% before the recovery happens.

Data: BUY BTC $63,046 → 0.1% drop → stop-loss SELL $62,982 (-$0.08 loss)

**Fix:** Per-strategy stop-loss. Mean-reversion uses 1% (original). Tick-momentum uses 0.1% (tight). Add `stop_loss_pct_momentum: 0.001` and `stop_loss_pct_meanrev: 0.01` to config.

## Problem 2: Bogus ticker prices

BTC is ~$63,000 but the bot bought at $66,126 — a fake price from a corrupted Binance API response. The price anomaly check at 10% only guards against exits, not entries. The entry itself was at a bogus price.

**Fix:** Before entering, compare the ticker price to the last N trade prices from the trade stream. If it deviates >5%, reject the entry.

## Files

- `engine.go` — skip stop-loss for mean-reversion (or use per-strategy stops) ~5 lines
- `engine.go` — add entry price sanity check using trade stream data ~10 lines
- `config.yaml` — add per-strategy stop-loss values ~3 lines
- `config.go` — add per-strategy stop-loss fields ~3 lines