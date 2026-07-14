# Why Bot Stops at 3 Trades — Diagnosis

## The Pattern

Grid buys $25 BTC → entryCount = 1  
Grid buys $25 BTC → entryCount = 2  
Grid buys $25 BTC → entryCount = 3 = maxPositions → **stops**.

Now 3 positions open, $75 deployed, $425 idle. Exits (take-profit 0.15-0.3%, stop-loss 0.5%, mean-rev reversal) rarely fire because BTC moves in a tight range. Positions sit forever. Bot dead.

## Affected Strategies

Every strategy hits the same wall:
- **Grid**: 3 positions max → stops after 3 buys
- **Mean-reversion**: 1 position per symbol → stops after 1 buy
- **Tick-momentum**: 1 position per symbol → stops after 1 buy

Once the position cap is reached, no more entries until an exit fires.

## Why Exits Don't Fire

BTC ranges between $62,500-$62,700 ($200 range, 0.3%). The exits need:
- **Take-profit 0.15%**: $93 move needed. Range sometimes hits, sometimes doesn't.
- **Take-profit 0.3%**: $188 move needed. Rare.
- **Stop-loss 0.5%**: $313 move needed. Never.
- **Mean-rev exit**: RSI > 55 + BB upper. Needs sustained uptrend.

In a sideways range, none reliably fire. Positions sit open indefinitely.

## Fix Options

### 1. Rotate positions: sell oldest if at cap
When maxPositions is reached and a new buy signal fires, sell the oldest position (even at a small loss) to make room.
```go
if entryCount >= maxPositions:
    sell oldest position at market
    buy new position
```

### 2. Time-stale exit: close positions older than X hours
If a grid position is >1 hour old and hasn't hit take-profit, close it at market. Gets capital rotating.

### 3. Lower position cap
Reduce maxPositions from 3 to 2, raise buyUSD from 25 to 35. Same capital deployed, fewer positions = faster rotation.

### 4. Reduce maxPositions to 1
Grid keeps 1 position open. Buys, waits for exit, then buys again. Simplest rotation.

## Recommendation: Option 3 + Option 2

- maxPositions: 3→2, buyUSD: 25→35 ($70 deployed)
- Stale grid positions > 2 hours old get sold at market

This keeps capital moving while maintaining DCA discipline.