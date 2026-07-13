# Take-Profit with Fees — Revised Targets

## I Missed Fees in My Earlier Calculation

Paper trader charges 0.1% fee per trade. The sell leg alone costs 0.1%:

| Take-Profit | Gross Gain | Minus 0.1% Sell Fee | Net Result |
|-------------|------------|---------------------|------------|
| 0.1% | +$63 | -$63 | **$0 (breakeven, no profit)** ❌ |
| 0.15% | +$94 | -$63 | **+$31 (+0.05% profit)** ✅ |
| 0.3% | +$188 | -$63 | **+$125 (+0.2% profit)** ✅ |

The 0.1% I suggested would result in zero profit — the fee eats the entire gain.

## Revised Targets (Fee-Aware)

| Exit | Old | Revised | BTC $ at $62,600 | Net Profit |
|------|-----|---------|-------------------|------------|
| **Partial 50%** | 1% | **0.15%** | ~$94 gross | +$31 net ✅ |
| **Full 100%** | 2% | **0.3%** | ~$188 gross | +$125 net ✅ |

The partial TP at 0.15% covers the 0.1% sell fee plus 0.05% actual profit. The full TP at 0.3% gives a wider margin.

## What About the Buy Fee?

The 0.1% buy fee was already paid when you entered. We can't recover that — it's sunk cost. The sell decision only needs to cover the sell fee (0.1%) plus profit.

## Files
`config/config.yaml` — 2 lines: `0.001 → 0.0015`, `0.003 → stays at 0.003`
