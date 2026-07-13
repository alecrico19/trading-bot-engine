# Catch Every Profit, No Matter How Small

## Goal

Every automated sell triggered at a profit. The bot currently has no exit for "price went up enough" on the micro-moves you see ($60 BTC swings).

## The Fix: Lower Take-Profit to Match Reality

| Exit | Current | Proposed | BTC $ at $62,600 | Triggers on |
|------|---------|----------|-------------------|-------------|
| **Partial 50% sell** | 1% | **0.1%** | ~$63 | Your $62,600 → $62,660 move ✅ |
| **Full 100% sell** | 2% | **0.3%** | ~$188 | Stronger sustained uptrend |

## How All Exits Work Together

```
Buy at $62,600
├── Price goes up $63 → Partial TP at 0.1%: sell 50% at +$0.04 profit
│   └── Remaining 50% rides
│       ├── Price keeps climbing → Full TP at 0.3%: sell rest at +$0.12 profit
│       └── Price peaks then drops 0.02% → Peak-drop: sell rest
├── Price peaks then immediately drops 0.02% → Peak-drop: sell 50%
└── Price drops 0.5% → Stop-loss: sell all at loss (rare, last resort)
```

## Every Sell Scenario

| Scenario | Exit | Result |
|----------|------|--------|
| Small uptick (+$63) | Partial TP 0.1% | +profit ✅ |
| Sustained uptick (+$188) | Full TP 0.3% | +profit ✅ |
| Peak and fade ($12 drop) | Peak-drop 0.02% | +profit (from earlier gain) ✅ |
| Trend reverses (downticks) | Momentum reversal | ±small (based on fill price) |
| Trade goes badly (-$315) | Stop-loss 0.5% | -loss (last resort) |

## Files
`config/config.yaml` — 2 lines changed
