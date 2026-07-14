# Is Research Adding Noise?

## Current State

- Active signals: 0
- Sentiment: 0.03 (neutral) — below 0.5 threshold for generating any signal
- Signal log: empty — no signals ever recorded
- Research service: running, consuming CPU/network for no output

## Answer: Not adding noise. Not adding value either.

The research service computes sentiment from news RSS feeds and generates signals when confidence > 0.6. But:

1. **Sentiment is almost always neutral (0.0-0.1)** — crypto news sentiment rarely spikes enough to trigger the 0.5 threshold
2. **Technical indicators are redundant** — the engine's mean-reversion strategy computes its own RSI and BB from live data. The research service computes the same from 5-min candles with a 60-second delay.
3. **Known bugs** — MACD was broken (now fixed), BB conditions always returned True (now fixed), but neither matters since nothing fires

## Does It Hurt?

No. The research runs on its own, doesn't interfere with engine decisions when no signals are active. But it wastes:
- CPU cycles (fetching RSS, computing indicators)
- Network (RSS feeds, ccxt OHLCV)
- Memory (Python process)

## Recommendation

**Option A: Keep it running.** It's not hurting anything and might occasionally produce a useful signal. Low cost.

**Option B: Disable research.** Remove the research process from the launcher. Saves resources. The engine's built-in strategies don't need external signals — they compute their own indicators internally.

**Option C: Fix research to actually produce signals.** Lower the confidence threshold (0.6 → 0.3), fix the MACD and BB bugs (already done), and add more trigger conditions. But even then, the signals are just "buy bias" or "sell bias" — they don't replace the strategy's own logic.

**Recommendation:** Option A for now. It's harmless. If you want to reduce resource usage, Option B is fine — the engine doesn't depend on research.