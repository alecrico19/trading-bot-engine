# Why Bot Buys High, Sells Low

## The Signal Problem

The scalping strategy uses `ratio = bid_volume / ask_volume`:
- ratio > 1.05 → heavy bids → BUY
- ratio < 0.95 → heavy asks → SELL

This is **following the crowd** — buying when buyers dominate, selling when sellers dominate. But on 100ms timescales, the crowd is already too late.

## The Timing Problem

```
[Buyers start jumping in] → ratio tips > 1.05 → bot: "BUY!"
                                           |
                                    But price already went UP
                                    |
[Bot buys at the peak] → price reverses
                                    |
[Sellers appear] → ratio tips < 0.95 → bot: "SELL!"
                                    |
                             Bot sells at the bottom
```

The ratio change is a LAGGING indicator. It confirms what ALREADY happened, not what WILL happen.

## What Professionals Do (Reverse Logic)

| Crowd Behavior | Bot Does (Wrong) | Pro Does (Right) |
|---------------|-----------------|-----------------|
| Heavy bids | BUY (buy high) | SELL into strength |
| Heavy asks | SELL (sell low) | BUY into weakness |
| Balanced book | Nothing | WAIT |

## The Fix: Reverse the Signal

```go
// Current (wrong):
if ratio > 1.05 → BUY   // buying when price is high
if ratio < 0.95 → SELL  // selling when price is low

// Fixed (correct):
if ratio < 0.95 → BUY   // buying when price is low (sellers are in control)
if ratio > 1.05 → SELL  // selling when price is high (buyers are in control)
```

This turns the strategy from momentum-following (lose money) to mean-reverting (make money). You buy when there's selling pressure (good price) and sell when there's buying pressure (good exit).

## Expected Effect
- Buy at the trough when sellers push price down → good entry
- Sell at the peak when buyers push price up → good exit
- Each trade captures the mean reversion instead of following it off a cliff