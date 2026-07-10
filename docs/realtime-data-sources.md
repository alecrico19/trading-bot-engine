# Real-Time Data Sources — Current State

## Tick-Momentum (active) — YES, realtime ✅
- **Source:** Binance WebSocket trade stream (`SubscribeTrades`)
- **Data:** Every individual trade fill — real price, real quantity, real time
- **Latency:** ~50-200ms (home PC → Binance → home PC)
- **How used:** Tracks last 5-50 trade prices, computes momentum direction
- **Realtime:** YES — this IS the same data that creates the chart candles

## Mean-Reversion (active) — YES, realtime ✅
- **Source:** Binance REST ticker (`FetchTicker`) every 15 seconds
- **Data:** Current bid/ask/last price
- **How used:** Builds internal PriceHistory from ticker values, computes RSI/BB
- **Realtime:** Near-realtime (15-second polling)

## Scalping (disabled) — was realtime ❌
- **Source:** Binance WebSocket order book (`SubscribeOrderBook`) at 100ms
- **Data:** Bid/ask volumes at nearest price levels
- **Why disabled:** Order book snapshots at 100ms are **stale** by the time we see them.
  Pros trade on the data in <1ms. Our 50-200ms latency = the move already happened.
- **Realtime:** YES but uselessly late

## Research Service (background) — NOT realtime
- **Source:** Binance REST OHLCV (`fetch_ohlcv`) every 60s
- **Data:** 5-minute candles
- **Used for:** Sentiment analysis + optional trading signals
- **Realtime:** No (60s delay, 5-min candles)
- **Current output:** 0 signals (sentiment always neutral)

## Answer

The bot IS using realtime Binance data for its active strategies. Tick-momentum reads every trade from the WebSocket — the same data you see on the 1s chart. Mean-reversion pollsticker every 15s. Both are real market data, not stale snapshots.
