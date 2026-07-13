# Why Sells Are Rejected — Rounding Mismatch

## The Data

```
Paper trader BTC balance: free=0.00127407
Sell order amount:        0.0012753477...

Difference: 0.0000013 BTC — tiny rounding error from fees
```

The strategy computes the sell amount from the **order manager's position** (full trade amount: 0.001279 BTC). The paper trader stores the **actual balance after fees** (0.001279 × 0.999 = 0.001274 BTC). The sell tries to sell the full trade amount, but only 99.9% is available. Rejected.

## The Fix: Sell From Actual Holdings

Instead of computing `holding` from the order manager's position, use the paper trader's actual balance (from `GetHoldings()`). The sell amount should be `min(strategy_computed_holding, paper_trader_balance)`.

Or simpler: in `checkStopLoss` and exit logic, use `baseHeld` (which already reads from the paper trader's actual balance) instead of the strategy's computed `holding`.

## Actually: The Fix Is Already There

The exit logic in `checkStopLoss` uses `baseHeld` (from holdings), not the strategy's computed holding. But the `executeDecision` path for scalping/tick-momentum sells DOES use the strategy's computed holding.

**Fix:** In `executeDecision`, when a SELL has `Amount` set by the strategy, cap it to the actual available balance from `GetHoldings()`.

## Files
`engine/engine.go` — cap sell amount at actual balance in executeDecision (~5 lines)
`engine/scalping.go` — already computing holding; just needs capping