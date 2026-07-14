# Mini-Profit Losing Money — Fix

## Root Cause

Mini-profit sells 50% at +0.05% gain. But real costs eat that:
- Bid-ask spread: 0.01-0.02%
- Trading fee: 0.1% (paper trader simulateFee)
- **Net: 0.05% gain - 0.12% cost = -0.07% loss**

The ticker says "you're up 0.05%" but after fees and spread, it's a loss.

## Fix Options

### A: Disable mini-profit (simplest)
Remove it entirely. Tick-momentum handles exits via momentum reversal.
Zero lines changed (already has the `false &&` pattern or just delete the block).

### B: Raise threshold to 0.15% (covers costs)
```yaml
mini_profit_pct: 0.0015  # 0.15%, covers 0.12% costs
```
Still marginal. On a $500 account, 0.03% net = $0.15 per trade.

### C: Make it net of fees
Calculate: `pnlPct >= 0.0012` (cover 0.12% costs). Whatever remains is profit.

## Recommendation: A — Disable it

Three exit mechanisms is too many: mini-profit, stop-loss, momentum reversal. Mini-profit was meant to catch uptrend rides, but it can't overcome basic trading costs at 0.05%. Tick-momentum's reversal exit + tight stop at 0.1% are sufficient.

## Files
`engine.go` — remove or disable mini-profit block (~10 lines)