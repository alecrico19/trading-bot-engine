# Honest Retrospective: What I'd Do Differently

## What We Actually Did (Wrong)

We built **60+ features** before validating a single strategy could make money:

```
Day 1:
├── Full Go engine with 4 exchange adapters
├── Bubble Tea TUI dashboard
├── Python research service (news, sentiment, indicators)
├── Redis pub/sub signaling
├── SQLite trade journal
├── Telegram bot
├── Obsidian vault sync
├── HTML live dashboard
├── Equity curve chart
├── Backtesting engine
├── Stop-loss, trailing stop, take-profit (3 types)
├── Time-exit, mini-profit exit (2 more exit types)
└── 261+ losing trades fighting scalping strategy

Result: Bot can't make money. Beautiful dashboard. Zero edge.
```

## What We Should Have Done (Minimal Path To Profit)

```
Hour 1-2:
├── Build ONLY mean-reversion strategy (RSI + BB)
├── Simple risk (fixed $40 position, max 5 concurrent)
├── Paper trade. No dashboard. No TUI. No Telegram.
└── Check P&L from SQLite or curl

Hour 3-4:
├── Run backtester → find optimal RSI/BB params
├── Apply optimized params → let run 2+ hours
└── If profitable → scale to $100/trade

Hour 5+:
├── Only then: add tick-momentum as SECONDARY
├── Only then: HTML dashboard (curl returns P&L was fine)
├── Only then: stop-loss (the ONE exit we needed)
└── Only then: Telegram bot for mobile monitoring
```

## Why This Order

| Wrong Order (What We Did) | Right Order |
|---------------------------|------------|
| Build exit types first | Verify strategy has edge first |
| Add 3 strategies at once | Run 1 strategy for hours |
| Change config 20+ times/day | Change 1 thing, measure, repeat |
| Build dashboard before profit | Profit first, dashboard second |
| Fix bugs reactively all day | Run stable, find issues naturally |
| Layer on features (TUI, Tailscale, vault) | Only build what's needed for the NEXT step |

## The Core Lesson

**80% of what we built (TUI, Telegram, research, charts, Obsidian docs, exchange adapters beyond Binance) was premature.** None of it helps the bot make money. The bot doesn't need a dashboard to be profitable — it needs a strategy with a verified edge, running stably, with minimal interference.

The RIGHT bot for maximum profit would be:
- 200 lines of Go (mean-reversion + simple risk)
- 0 lines of Python (no research service)
- 0 lines of HTML/CSS (no dashboard)
- 0 Telegram bot
- Just a `curl` endpoint for P&L
- Running untouched for 4+ hours

And it would have been profitable on Day 1, not Day 3.
