# Take-Profit Assessment + Peak-Drop Exit

## Is 1-2% Take-Profit Realistic?

**No, not for short-term scalping.** It's safe/conservative — meant to ensure you don't cut winners too early. But on our timeframe (seconds to minutes):
- 2% on BTC = $1,260 — might take days
- Position sits open, exposed to reversal the whole time
- The take-profit at 1-2% was designed for swing trading, not scalping

**What's realistic for our timeframe:**
- Scalping (seconds): 0.05-0.2% per trade
- Tick-momentum (minutes): 0.2-0.5% per trade
- Mean-reversion (hours): 1-3% per trade

## The Peak-Drop Exit

**More realistic than take-profit for our use case.** A 0.02% drop from the peak ($12.50 on BTC) is achievable in seconds. It catches the transition from "going up" to "going down" — exactly what you observed at $62,782 → declining.

## How All Exits Work Together

| Exit | Triggers At | Catches |
|------|------------|---------|
| **Peak-drop (new)** | 0.02% from peak | Micro-reversal: "price just stopped going up" |
| **Momentum reversal** | 40-45% downticks | Sustained downtrend: "multiple trades going down" |
| **Stop-loss** | 0.5% loss | Continued failure: "this trade is wrong" |
| **Take-profit** | 1-2% gain | Rare big moves: "massive win" |

## Recommendation

**Build the peak-drop exit now.** It addresses your specific observation: the bot bought, price peaked, then gently declined with no sell. The take-profit at 1-2% can stay as a safety net for rare big moves, but the peak-drop is the real workhorse for our timeframe.

## Files
`engine.go` — add peak-drop check in `checkStopLoss` (~10 lines)