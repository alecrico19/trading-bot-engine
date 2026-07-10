# Why Do We Keep Losing Money? — The Data

## Good news: P&L IS being calculated

The diagnostic logs confirm the engine IS producing non-zero P&L:

```
BTCUSDT sell: entry=64084.69  avg_price=64075.04  pnl=-0.012
exit trade P&L (time-exit): pnl=-0.079
exit trade P&L (time-exit on ETH): pnl=-0.159
```

The dashboard shows $0.00 because the DB has old pre-fix trades mixed in, and the values are very small (fractions of a cent).

## Bad news: Time-exit still bleeding money

| Component | P&L | 
|-----------|-----|
| Scalping (reversed signals) | -$0.04 |
| Time-exit (90s) | -$1.41 |
| **Total** | **-$1.45** |

The reversed signals WORKED — scalping went from -$0.58 to -$0.04 (near breakeven). But time-exit is still the problem. It closes positions with tiny 0.003-0.05% profits that get eaten by fees.

## Root cause: Exit timing doesn't match entry timing

The scalping entries target 0.1-0.3% moves. But the time-exit fires when the position shows even 0.003% profit (barely breakeven after fees). 90 seconds isn't always enough time for the move to develop. When it hasn't, a tiny profit minus fees = net loss.

## Are you impatient?

Partly, but the bigger issue is the time-exit is too aggressive. The reversed signals are the right fix — you're now buying low and selling high. But the time-exit is cutting winning trades short and converting small wins into small losses after fees. Removing the time-exit and relying only on take-profit (1-2%) + trailing stop would let trades run longer and capture larger moves.

## Fix: Disable time-exit, keep take-profit + trailing stop
- Remove the 90-second time-based exit
- Let take-profit (1-2%) capture real gains
- Let trailing stop lock in larger moves
- Let scalping signals close positions when the imbalance reverses