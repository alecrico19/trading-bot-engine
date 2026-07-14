# Integrate, Separate, or Just Scalping?

## Option A: Integrate (one strategy, dual confirmation)

```
Scalping window (3-5 trades) + Tick window (5-50 trades)
Must both agree: "3 upticks AND overall momentum up" → BUY
Exit: either window reverses
```

| Pro | Con |
|-----|-----|
| Higher quality entries | Fewer trades |
| Fewer false signals | More complex code |
| One position to manage | Both must agree |

## Option B: Separate (both run, share trade data)

```
Scalping: 3-5 trade window, fast entries, many trades
Tick-momentum: 5-50 trade window, confirmed entries, fewer trades
Both get trade data from same Binance WebSocket
```

| Pro | Con |
|-----|-----|
| More total trades | Two positions per symbol possible |
| Redundant — one succeeds if other fails | Competing for same capital |
| Tick-momentum was already profitable (+$0.05) | Potentially conflicting signals |

## Option C: Just scalping (remove tick-momentum)

```
Only scalping on trade data
Fast, many trades, simple
```

| Pro | Con |
|-----|-----|
| Simplest, most trades | Less confirmation |
| No signal conflicts | More false signals |
| Easy to tune one strategy | No proven profit (scalping was -$8/hour) |

## Recommendation: Option B

Tick-momentum already proved it can be profitable (+$0.05 on 9 trades). Scalping on trade data adds more opportunities. Keep both separate — if one fails, the other still works. They're complementary:
- **Scalping** catches $10-20 micro-moves (the ones you see on the 1s chart)
- **Tick-momentum** catches $50-100 macro-moves (sustained trends)

Both use the same real-time Binance trade data. No order books. No stale snapshots. Just actual fills at real prices.
