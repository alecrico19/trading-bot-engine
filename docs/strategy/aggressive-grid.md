# Aggressive Grid for $500 Capital

## The Problem

$10 every 4 hours = too slow to compound with $500. Need tighter, faster cycles.

## Adjusted Grid: Fast + Small Profits

| Parameter | Conservative | Aggressive |
|-----------|-------------|------------|
| Buy interval | 4 hours | **2 hours** |
| Buy amount | $10 | **$25** |
| Profit target | 2% | **0.5%** |
| Position cap | Unlimited | **3 concurrent** |
| Max deployed | $500 over 80h | **$75 active, $25 cycling** |

## How It Works

```
2:00 PM: Buy $25 BTC
2:30 PM: BTC up 0.5% → sell for +$0.12 profit
2:31 PM: Buy another $25 BTC
3:15 PM: BTC up 0.5% → sell for +$0.12 profit
... repeat 6-12 times per day
```

6 cycles × $0.12 = $0.72/day. 30 days = $21.60 (4.3% monthly).

Not retirement money, but consistently green.

## Combining Grid + Mean-Reversion

Mean-reversion buys oversold dips. Grid buys at regular intervals. Together they cover different entry types:

```
Mean-reversion: "BTC just crashed 1% → RSI oversold → BUY"
Grid:          "It's been 2 hours → BUY $25 regardless"
```

When one is wrong, the other might be right. Both use the SAME exit system: take-profit 0.15-0.3% + stop-loss 0.5%.

## Expected Daily P&L

| Source | Cycles/Day | Profit/Cycle | Daily |
|--------|-----------|-------------|-------|
| Grid buys | 6-8 | $0.10-0.15 | $0.60-1.20 |
| Mean-rev entries | 2-4 | $0.05-0.15 | $0.10-0.60 |
| **Total** | | | **$0.70-1.80** |

Weekly: $5-13. Monthly: $20-50. 4-10% monthly return.

## Why This Works Where Momentum Failed

- Grid: buys at ALL times — catches every dip without timing
- Mean-reversion: buys at SPECIFIC good times (RSI oversold)
- Both exit ONLY at profit (take-profit) or controlled loss (stop-loss)
- No momentum noise, no micro-reversal exits
- Simple, boring, consistently green

## Implementation

- `engine/grid.go` — new strategy (~40 lines)
- `config.yaml` — grid config block (~5 lines)  
- `cmd/main.go` — register strategy (~3 lines)
- Total: ~50 lines of new code