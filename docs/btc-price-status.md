# BTC Prices — What the Bot Shows

## Mean-Reversion Entry (Just Fired!)

```
9:48 PM — BTC bought at $62,272.29 (mean-reversion)
9:49 PM — ETH bought at $1,765.41 (mean-reversion)
```

Both entries fired because RSI hit oversold (<35) + price touched lower BB (1.5 stddev). The bot is now holding both positions, waiting for exit conditions.

## Holdings

| Asset | Amount | 
|-------|--------|
| BTC | 0.001283 |
| ETH | 0.045265 |
| USDT | $840 (remaining cash) |

## What Happens Next

Bot waits for:
1. **RSI > 65 + price > upper BB** — mean-reversion exit (sell at profit)
2. **Take-profit 0.15%** — BTC at ~$62,365, ETH at ~$1,768
3. **Take-profit 0.3%** — BTC at ~$62,459, ETH at ~$1,771
4. **Stop-loss 0.5%** — BTC at ~$61,961, ETH at ~$1,757

## For Live BTC Prices

I can't fetch real-time Binance data from here. Open Binance and check the current price. Compare to the entry at $62,272 to see if you're in profit or loss right now.