# Is This the Best Path? — Three Options

## Where We Are

- Grid and mean-reversion fighting on BTC (cancel each other out)
- Bot idle 1+ hour with $998 in cash
- 21 trades, -$0.05 — close to flat

## Three Paths Forward

### Option A: Separate Symbols (Config Only)
```
Grid: BTC only, buys every 2 hours
Mean-rev: ETH only, trades RSI + BB
```
| Pro | Con |
|-----|-----|
| Zero code changes | Each strategy covers only 1 symbol |
| No more fighting | Less total trading |
| Simpler to debug | If ETH doesn't move, mean-rev sits idle |

### Option B: Grid Only (Simplest)
```
Grid: BTC only, buys every 2 hours
Mean-rev: disabled
```
| Pro | Con |
|-----|-----|
| Only 1 strategy = zero fighting | No mean-reversion entries |
| Most consistent: buys every 2h, sells at profit | Slower — only grid cycles |
| Proven: DCA always wins over time | Less engaging to watch |

### Option C: Per-Strategy Tracking (Most Complex)
```
Both strategies on BTC
Positions tagged with strategy name
Mean-rev only sells its own entries
Grid entries managed by take-profit/stop-loss
```
| Pro | Con |
|-----|-----|
| Both on BTC = max trades | ~30 lines of new code |
| Most sophisticated | More code = more bugs |
| Professional-grade position management | Takes time to build |

## My Recommendation: Option A

It solves the immediate problem (fighting strategies) with zero code changes. Both strategies run independently. If mean-reversion does well on ETH and grid does well on BTC, you have two profit sources.

After letting Option A run for a few hours, you can decide: keep both, scale one up, or switch to Option B.

## Files (Option A)

`config/config.yaml` — 2 lines changed (BTC→ETH in mean-rev symbols)