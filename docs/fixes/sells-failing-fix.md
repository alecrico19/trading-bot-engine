# Bot Can't Sell — Sells Rejected, Not "Selling Too Fast"

## The Real Problem

```
4 buys placed  → fills OK
2 sells placed → ALL rejected (insufficient balance)
0 sells filled → bot holds positions it can't exit
```

The bot generates sell signals correctly — but every sell fails at the paper trader. The paper trader's balance tracking and the order manager's position tracking disagree about what's held.

## What Actually Needs Fixing

The positions are tracked by the order manager (Go struct in memory). The balances are tracked by the paper trader (Go struct in memory). When a BUY fills:
- Order manager: adds to position ✓
- Paper trader: adjusts balance ✓

But the SELL checks `hasPosition` from the order manager, which says "yes, you have BTC." The paper trader says "no, you don't have enough."

**Root cause:** The paper trader's balance matches the order manager's position — both show BTC: 0.0013 (verified from /balance). The sells ARE for 0.0013 BTC. So why are they rejected?

The sell is probably being placed for the WRONG amount or with the WRONG base asset name. The strategy computes `holding` from the order manager's position, but the paper trader's `adjustBalance` might be looking for a different asset key.

## Fix: Debug the Exact Rejection

Add a log line in the paper trader showing WHY the sell was rejected:
```go
logger.Warn().Float64("available", b.Free).Float64("needed", amount).Str("base", base).Msg("sell rejected")
```

This tells us: "you tried to sell 0.0013 BTC, but only 0.0010 is free (0.0003 is locked)" or "you tried to sell BTC but the asset key is wrong."

## Files
`exchange/paper.go` — add diagnostic log on sell rejection (~2 lines)