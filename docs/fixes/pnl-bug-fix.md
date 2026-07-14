# Paper Trader Balance Bug Fix

## Bugs Found

### Bug 1: Phantom money from negative balance zeroing
`adjustBalance` has `if b.Free < 0 { b.Free = 0 }` — when a BUY costs more than available balance, the deficit is silently absorbed instead of rejecting. Over rapid scalping cycles, this accumulates phantom money.

**Before fix:** Start $1000 → after rapid trades → $2098 (impossible)
**After fix:** Start $1000 → after trades → $999.96 (correct, matches strategy P&L of -$0.04)

### Bug 2: Empty asset entries in holdings
`splitSymbol("BTCUSDT")` returns `("BTC", "USDT")` which works, but some order symbols might not have `/` in them (e.g., `"BTC"` alone), resulting in empty `""` asset entries.

## Fix

### Paper trader (`exchange/paper.go`)
- Remove `if b.Free < 0 { b.Free = 0 }` — instead, cap deduction at available balance and log a warning
- When buying, check if balance is sufficient: `if b.Free < cost { return }`
- Add empty-asset filter to `FetchBalance`

### Engine (`engine.go`)
- Record actual P&L per trade (already done correctly)
- Track cumulative realized P&L vs cumulative paper balance — log if they diverge

## Files
`exchange/paper.go` — ~10 lines changed

## Expected Result
Paper balance stays close to initial + sum of trade P&Ls. No phantom money.
