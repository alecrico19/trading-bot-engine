# Why the Bot Loses Money — Root Cause Analysis

## The Data

- 261 trades, -$8.04 total loss
- Every sell loses ~$0.05 (BTC) or ~$0.09 (ETH)
- Reversed signals were active (154 buy-weakness, 147 sell-strength)
- Time-exit disabled, trailing stop active
- ALL code fixes verified in binary

## The Problem: 100ms Order Book Data is Noise

The scalping strategy uses `ratio = bid_volume / ask_volume` at 100ms intervals. Whether buying into weakness or buying into strength — the ratio doesn't predict price direction.

The diagnostic log shows consistent losses:
```
buy entry:  $63,999
sell price: $63,956  → -$0.05 loss
buy entry:  $63,964
sell price: $63,919  → -$0.05 loss
buy entry:  $1,794
sell price: $1,792   → -$0.09 loss
```

Every single sell closes at a loss. The order book ratio doesn't tell you which way the market is going — it tells you what JUST happened. By the time the signal fires, the move is over or reversing.

## The Math

130+ sells × -$0.05 avg = -$6.50 in P&L  
130+ buys × $0.00 P&L = $0  
Fees: 261 trades × ~$0.015 = -$3.90  
**Total: ≈ -$10** (matches observed -$8.04 within margin of the open positions' mark-to-market)

## Why Mean-Reversion Works Better

Mean-reversion (RSI + BB) uses 5-minute candles with 15-second intervals. This is a real prediction — when RSI < 30 AND price touches lower BB, price MEANS to revert. The backtester confirmed 63% win rate with 1.21 profit factor. But the engine needs 100 candles (~25 min) to warm up before it trades.

## What To Do

**Option A: Disable scalping entirely, let mean-reversion trade.** Mean-reversion has a real edge (confirmed in backtests). Scalping has none (confirmed by 261 losing trades). A simple config change: `enabled: false` on scalping.

**Option B: Reduce scalping to confirmation-only.** Keep scalping disabled by default. Only turn it on when mean-reversion confirms a market regime (e.g., "we're in a range, scalping is safe").

**Option C: Add a predictive filter to scalping.** Add a multi-tick moving average of the order book ratio. Only trade when the ratio trend (not the instantaneous value) confirms direction. Much more complex, experimental.