# Bot Stopped — Sold Everything, Can't Re-Enter

## The Data

```
Last trade: 9:33 PM (15+ min ago)
BTC: 0.00000128 (basically zero)
ETH: 0.00000000 (zero)
USDT: $997.85 (almost all capital idle)
19 trades, -$0.79 P&L
```

## What Happened

The bot sold everything. Now it has no positions, and it can't find a buy signal to re-enter. The trend filter blocks buys when the broader trend is down, and the market may not be producing clear uptick signals.

This is the same "bot stops after depleting positions" pattern we've seen throughout the day.

## Good News

Take-profit IS firing now — ETH hit +0.3% and the full take-profit sold. The exit system is working. But:
1. The entries haven't been profitable enough to overcome the losses
2. When positions close, the bot can't find new entries in a flat/down market

## Why The Bot Can't Re-Enter

Possible causes (diagnose one of these):
1. Trend filter blocking: broader window shows "down" → buys blocked
2. No uptick signals: market is flat → no 55%+ uptick windows
3. Risk manager: P&L at -$0.79 might be near daily loss limit

## Fix: Check What's Blocking Entries

Add a diagnostic log: every time a buy is blocked, log the reason:
```
"scalp buy blocked: trend=down"
"scalp buy blocked: upPct=0.48 (threshold=0.55)"
"tick-momentum buy blocked: trend=down"
"mean-rev buy blocked: RSI=42 (need <30)"
```

This tells us which of the 3 strategies is closest to firing and what's holding it back.

## Files
- `scalping.go` — add blocked-buy debug logs (~5 lines)
- `tick_momentum.go` — add blocked-buy debug logs (~5 lines)
