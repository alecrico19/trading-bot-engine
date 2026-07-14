# Bot Lost on Small Sells, Missed $200 BTC Move

## What Happened

```
Journey: 19 trades, -$0.79 P&L
         Many small sells at micro-losses
         Positions depleted → bot idle
         BTC then goes UP $200+ → bot has no position, can't participate
```

The bot sold too early on many small reversals, accumulated -$0.79 in fees and slippage. Then when the real move came, it was out of the market.

## Root Cause: Exit Systems Compete with Each Other

Four exit systems fire independently:
1. **Peak-drop (0.02%)** — sells on any micro-pullback from peak
2. **Take-profit (0.15%)** — sells 50% on moderate gains
3. **Momentum reversal (40-45%)** — sells on downticks
4. **Stop-loss (0.5%)** — sells on sustained loss

In a choppy uptrend, exits 1, 2, and 3 ALL fire in rapid succession, depleting the position before the big move can develop. The bot takes microscopic profits that can't overcome fees + small losses from whipsaw reversals.

## What Professional Scalping Research Says

From the Zeus research report (15 pro scalper videos):
> "Never cut winners early. Let runners run to 2R. Scale out 50% at 1R, trail the rest."

Our bot does the opposite: it scales out continuously through peak-drop (0.02%), partial TP (0.15%), and momentum exits (any reversal). By the time the big move happens, there's nothing left.

## Fix Options

### Option A: Fewer Exits, Longer Holds
Remove peak-drop exit. Keep only:
- **Take-profit 0.15%/0.3%** — the 50% partial + 100% full
- **Stop-loss 0.5%** — safety net
- **Momentum reversal** — but only on MAJOR reversals (30%+, not 45%)

This means: buy, hold through micro-reversals, exit only on confirmed trend change or profit target.

### Option B: Scale Entries, Not Exits
Instead of selling 50% at every micro-move, the bot should BUY more on dips within the trend. If BTC is in an uptrend and dips 0.02%, buy more instead of selling.

### Option C: Trend-First Approach
Only trade in one direction (buy in uptrends, sell in downtrends). Never exit because of a micro-reversal — only exit on MAJOR reversal or profit target.

## Recommendation: Option A (Fewer Exits)

The simplest and most impactful. Remove peak-drop (0.02%) and tighten the momentum exit to only fire on clear reversals. Let the take-profit handle profitable exits. Let the position ride through noise.

## Files
- `engine.go` — disable peak-drop (1 line: `false &&`)
- `scalping.go` — raise momentum exit from 40%→30% downticks
- `tick_momentum.go` — raise momentum exit from 45%→35% downticks