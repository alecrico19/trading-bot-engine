# Trade Analysis — Still $0 P&L + Accumulation Pattern

## What the data shows

### Good news: ETH trades ARE profitable
```
BUY  $1797.58 → BUY $1798.03 → SELL $1798.33 (+$0.033 profit if P&L worked)
BUY  $1798.26 → SELL $1798.59 (+$0.014 profit if P&L worked)
```
The ETH scalping is buying low and selling higher. The strategy isn't broken — the P&L tracking is.

### Bad news: P&L still $0.00
The running engine predates the diagnostic logging and entry blending fixes (commit 076060f). Need to restart to pick them up.

### Bad news: Position accumulation
BTC: 5 buys (0.0060 BTC), 1 sell (0.0012). 0.0048 BTC stuck open.
The sell from take-profit or stop-loss should close this, but stop-loss/take-profit were also broken by the splitSymbol bug. With the splitSymbol fix, they should work now.

## What needs to happen

### 1. Restart the launcher (immediate)
Picks up: splitSymbol fix, entry blending fix, diagnostic logging, positions panel fix.

### 2. Then build anti-accumulation logic (build after restart)
After X buys on the same symbol without a sell, stop buying. This prevents the bot from sinking too much capital into one position.

### 3. Then build position-aware selling (build after that)
Only allow SELL if the bot currently holds the asset. This was flagged as audit #17 and the old engine log confirms it was selling without positions.

## Files needed
- `engine.go` — anti-accumulation (+5 lines in executeDecision)
- `scalping.go` — position-aware selling (+10 lines)