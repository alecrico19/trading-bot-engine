# Take-Profit-50 Selling Repeatedly at Low Prices — Fix

## The Bug

Take-profit-50 fires at +0.15%, sells 50%, but keeps `entryPrice` the same. Next cycle (3s later), pnlPct is STILL above 0.15% — fires again. And again. And again. Position gets whittled down:

```
First fire:  sell 50% of 0.045 ETH → 0.0225 sold
Next fire:   sell 50% of 0.0225 ETH → 0.0112 sold
Next fire:   sell 50% of 0.0056 ETH → 0.0028 sold
...until amount < min order size
```

From your log: amounts of 0.00002 and 0.00001 ETH — virtually dust. Each sell at essentially the same price but now with tiny amounts.

## The Fix

Same as what we did for mini-profit: after take-profit-50 fires, reset entryPrice to 0 so it doesn't re-fire. The remaining position gets handled by the take-profit-full (0.3%) or the mean-reversion exit.

```go
// After the 50% sell:
e.entryPrice[symbol] = 0  // prevent re-fire
```

Or: add a flag `tookProfit50` per symbol that blocks the second fire of the 50% take-profit.

## Files
`engine.go` — change `e.entryPrice[symbol] = entry` to `e.entryPrice[symbol] = 0` after take-profit-50 (~1 line)