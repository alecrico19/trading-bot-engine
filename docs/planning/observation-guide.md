# Observation Period — Let It Run

## Current Configuration (Stable)

| Setting | Value |
|---------|-------|
| Tick-momentum | Enabled (60% threshold, 2s cooldown, 50-trade warmup) |
| Mean-reversion | Enabled (RSI + BB, needs ~25 min warmup) |
| Scalping | Disabled |
| Position size | 8% ($40 on $500) |
| Max concurrent | 5 |
| Tight stop | 0.1% |
| Mini-profit | Disabled |
| Time-exit (losers) | Disabled |

## What To Check After 1 Hour

1. **P&L**: `curl -s http://localhost:8080/status | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['daily_pnl'])"`
2. **Strategy breakdown**: Which strategy is earning/losing?
3. **Trade count**: Expected 5-15 trades/hour
4. **Equity curve**: Is it trending up, down, or flat?

## If Profitable After 1 Hour
→ Scale to 20% position size, let it run another hour, compare

## If Still Losing After 1 Hour  
→ The momentum-following approach on retail latency may not have an edge
→ Consider switching to mean-reversion only (RSI+BB, proven 63% WR in backtests)
