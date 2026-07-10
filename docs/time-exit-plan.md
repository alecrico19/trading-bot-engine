# Time-Based Profit Exit (Scalping)

## What It Does
If a scalping position is still open after 30 seconds AND the price is above the entry price → market sell to lock in the gain.

## What It Does NOT Do
Never sells at a loss. Losses are handled by the existing stop-loss (-5%).

## Logic
```
Every 3 seconds (in runStopLoss goroutine):
  For each open position:
    If strategy == "scalping" AND age > 30s AND price > entry:
      Market sell the entire position
```

## Why This Works
- Catches your $16 BTC gain that was stuck open
- Prevents position accumulation (no more "5 buys, 1 sell")
- Never sells at a loss (only fires when profitable)
- Works at any price level — zero config needed
- Matches pro scalper behavior: "the setup played or it didn't — get out"

## Files
`engine.go` — ~10 lines in checkStopLoss or a new checkTimeExit function
