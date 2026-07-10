# Should Scalping + Tick-Momentum Work Together?

## Short Answer

Probably not. The order book at 100ms is **stale** compared to the trade stream. Trades happen first, then the order book updates. By the time our bot sees the order book, the trades already told us the direction. The order book adds no new information — it just confirms (with lag) what trades already showed.

```
Timeline:
t=0ms:    Big buyer hits the ask → price jumps
t=5ms:    Trade appears in our WebSocket stream → tick-momentum sees it
t=50ms:   Order book updates to show "heavy bid volume"
t=100ms:  Our bot's next order book snapshot → sees the ratio
```

Tick-momentum already reacted at t=5ms. Order book confirms at t=100ms — 95ms too late.

## Where Order Book COULD Help

### Rate-of-change, not ratio
Pros don't look at "is bid > ask." They look at "is the bid wall GROWING or SHRINKING?" This is predictive — if bids are accumulating at $64,300 faster than asks at $64,301, price is about to go UP. Our bot just checks the snapshot ratio.

To do this, we'd need to track the book depth PER PRICE LEVEL over multiple snapshots — much more complex.

### Confirmation, not entry
Order book could gate entries: "tick-momentum says buy AND order book confirms bullish" = higher confidence. But the order book always confirms bullish a few milliseconds after trades go bullish — it's a lagging indicator, not an independent one.

## What Would Actually Help

| Approach | Value |
|----------|-------|
| Track tick price + volume together | Weight bigger trades more heavily |
| SMALLER threshold for exit, BIGGER for entry | Enter on strong signal, exit on weak reversal |
| Multi-timeframe tick momentum | 5-tick window + 20-tick window — both must agree |
| Volume-weighted tick direction | 10 BTC bought > 0.1 BTC bought |

## Recommendation

**Don't re-enable scalping yet.** Let tick-momentum run for a few hours on paper. If it produces trades and shows a profit trend, great — tune it. If it also loses money, the problem isn't the signal source (order book vs trades) — it's the fundamental approach (momentum-following at retail latency).