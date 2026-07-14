# ETH Dust Position Blocking Mean-Reversion — Fix

## Problem

ETH position is 0.000022 ($0.04) — too small for the $10 minimum order size. Mean-reversion keeps trying to sell it every few seconds, getting blocked by order validation.

## Fix: Skip Dust Positions

In mean-reversion's Evaluate, before generating a sell decision, check if the position amount × price meets the minimum. If not, return nil.

```go
if holding * ticker.Last < 10 {
    return nil  // position too small to sell
}
```

Or in executeDecision: skip the order if the notional is below minimum.

## Quick Workaround (No Code)

Reset the paper trader's ETH balance to 0. The dust position disappears. Restart bot fresh.

## Files

`engine/mean_reversion.go` — skip dust positions (~3 lines)