# Tick-Momentum Stopped — Fix

## What Happened

Bot was making money (+$0.05 P&L!) then stopped entirely. The log shows:

1. **5:04 AM** — Last real trade (ETH sell)
2. **5:25-5:26 AM** — 15+ SELL signals fired in rapid succession with no BUYs
3. Prices dropping $63,996-$64,002 — tick momentum shows all downticks
4. Strategy fires SELL on every downtick — but nothing to sell (positions already closed)
5. Fill status check rejects empty sells — no new trades recorded

**Root cause:** After the first sell closes the position, the strategy keeps firing SELL signals because momentum is still down. The fills get rejected by the paper trader. No new BUYs fire because momentum never flips to "up" — price keeps dropping.

## Fix Plan

### Fix 1: Cooldown between signals on the same symbol
Don't fire another decision within 2 seconds of the last decision for the same symbol. This prevents the 15-sells-in-a-row spam.

```go
// In Evaluate, add:
if time.Since(s.lastDecisionTime[symbol]) < 2*time.Second {
    return nil
}
s.lastDecisionTime[symbol] = time.Now()
```

### Fix 2: Only sell if we actually hold something
The `hasPosition` check already exists but may not reflect the latest state. Add a safety check: only generate a SELL if the position has been held for at least one engine cycle.

### Fix 3: Limit max signals per minute
Global cap: max 10 decisions per minute. Prevents runaway signal loops.

## Files
- `tick_momentum.go` — Fix 1 (+5 lines), Fix 2 (+3 lines)
- `engine.go` — Fix 3 (+5 lines, decision counter)

## Expected Result
After fix: buys fire on uptick momentum, sells fire once on reversal, no more spam.