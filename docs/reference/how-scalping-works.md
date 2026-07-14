# How Scalping Works

## What Is Scalping?

Scalping is a trading strategy that aims to profit from **very small price changes**, typically 0.1% to 0.5% per trade. Scalpers make 20-100+ trades per day, each lasting seconds to minutes, banking tiny profits that add up over volume.

Think of it like picking up pennies in front of a steamroller — small, frequent wins that compound, but one bad move can wipe days of gains.

## How Scalpers Make Money

| Element | Scalping | Day Trading | Swing Trading |
|---------|----------|-------------|---------------|
| Hold time | Seconds to 2 min | 5 min to hours | Days to weeks |
| Profit target | 0.05-0.3% | 0.5-3% | 3-10% |
| Trades per day | 20-100+ | 2-10 | 1-5 per week |
| Decision speed | Milliseconds | Seconds | Minutes/hours |
| Relies on | Order book, tick data | Technical analysis, chart patterns | Fundamentals, trends |

The math: 50 trades × 0.1% profit each = 5% daily gain. Lose 0.2% on 20 trades = net positive.

## What Our Bot Did (Order Book Scalping)

Our scalping strategy used **order book imbalance**:

1. **Read the order book** — how many BTC are being bid (buyers) vs asked (sellers) at the nearest price levels
2. **Compute ratio** = bid_volume / ask_volume
3. **If ratio > 1.05** — more buyers than sellers → buy (momentum likely up)
4. **If ratio < 0.95** — more sellers than buyers → sell (momentum likely down)

This is called **order-flow scalping**. Pro HFT firms do this at <1ms latency with co-located servers. 

## Why It Failed (For Us)

The order book tells you what **just happened**, not what **will happen**. By the time our bot (50-200ms latency from home PC → Binance) sees the ratio tip, the pros have already traded, the price has moved, and the signal is worthless.

We proved this with 261 trades — 0 profitable sells, -$8.04 loss. The signal was always arriving too late.

## What We Switched To (Tick Momentum)

Instead of reading order book snapshots (stale), we now read **actual trade prices** from the same Binance WebSocket:

1. Every trade event: "0.5 BTC bought at $64,310"
2. Track the last 5-10 trade prices
3. If 60%+ are going up → buy (momentum confirmed by real fills)
4. If momentum reverses → sell

This uses REAL fills at REAL prices — the same data that creates the $40 price swings visible on the 1s chart.

## Professional Scalping (What Facebook Ads Show)

| What the ads show | Reality |
|-------------------|---------|
| "100 trades/second" | Colocated server, direct exchange data feed, $10k+/mo infrastructure |
| "99% win rate" | Marketing. Most pro scalpers have 50-65% win rates |
| "Click a button, make money" | Years of experience, custom ML models, $100k+ accounts |
| "Works on any device" | Need <1ms latency. Home internet can't compete |

## Can Retail Scalp?

**At sub-second speeds? No.** The infrastructure gap is too big.

**At 1-5 minute speeds? Yes.** Mean-reversion (RSI + Bollinger Bands) on 5-min candles works because 200ms latency doesn't matter on a 5-minute timeframe. The pros can't arbitrage a 5-minute signal.

This is why we're keeping mean-reversion enabled and adding tick-momentum as a middle-ground strategy (~1-5 second speed, uses real trade data).
