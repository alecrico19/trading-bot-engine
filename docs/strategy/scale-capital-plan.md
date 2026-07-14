# Deploy More Capital — Plan

## Current State
- max_position_pct: 8% → $40/trade on $500
- 5 max concurrent = $200 deployed (40%)
- 60% of capital idle

## The Real Question

Not "should we use more money?" — but "**does the bot make money at current size?**"

If the bot loses $4/hour at $40/trade, it loses $20/hour at $200/trade. Scaling a losing strategy makes it lose faster.

## The Plan

### Step 1: Let it run (1 hour, no changes)
Current config: tick-momentum + mean-reversion, mini-profit disabled, tight stop 0.1%, time-exit disabled.

### Step 2: Check P&L after 1 hour
```bash
curl -s http://localhost:8080/status | python3 -c "import sys,json; d=json.load(sys.stdin); print(f'P&L: \${d[\"daily_pnl\"]:.4f} | Trades: {d[\"trades\"]}')"
```

### Step 3: If profitable → scale
```yaml
risk:
  max_position_pct: 0.20  # $100/trade (was $40)
```
1 config line. 2.5x bigger positions. $500 deployed (100% of capital) with 5 concurrent.

### Step 4: If not profitable → don't scale
Fix the strategy first. More money won't fix a losing edge.

## Don't Scale Until Proven

We've made 20+ changes today. Nothing has run for more than 30 minutes without a change. Scaling a mystery box is gambling, not trading.
