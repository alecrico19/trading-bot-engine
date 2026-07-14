# Milestone Log — Balanced Profitability Profile

**Status:** 🟢 Running (paper) · **Profile:** Balanced · **Last updated:** 2026-07-14 21:52

Running log of significant changes and live events. Updated at every milestone.
Related: [[why-stops-at-3]] · [[honest-assessment]] · [[retrospective]] · [[dca-grid-plan]] · [[best-path]]

---

## Current config (Balanced)

| Setting | Value |
|---|---|
| Position size | ~$50/trade |
| Max concurrent positions | 3 (grid ≤2 BTC, mean-rev ≤1 ETH) |
| Exit engine | **event-driven** (every trade tick) + 3s poll backstop |
| Scale-out | sell **50% at +0.5%** into strength, once per position |
| Full take-profit backstop | +1.0% (remainder runs on trailing) |
| Trailing stop | arms +0.4%, trails 0.2% below peak |
| Stop-loss | −0.8% |
| Daily profit-lock | +3% → flatten + halt for day |
| Daily loss-stop | −3% → flatten + halt for day |
| Crash guard | block new buys if −2% over 30 min |
| Reported P&L | net of 0.2% round-trip fee |

Dashboard: http://localhost:8420 · Logs: `/tmp/trading-bot-logs/engine.log`

---

## Timeline

### 2026-07-14 — Session: grid fix + Balanced profile + crash guard

- **Root cause fixed — grid never traded.** `collectSymbols()` had no `grid` case,
  so BTCUSDT never got a trade-stream goroutine and `grid.Evaluate()` was never called.
  This — not any threshold — is why the grid "registered but never bought". See [[grid-not-buying]].
- **Exit/P&L accuracy.** Exits now read the order manager's volume-weighted `AvgPrice`
  instead of a hand-rolled `(prev+new)/2` average that skewed multi-leg grid math.
- **Grid stale-exit.** Grid legs held >2h are rotated out at market. See [[why-stops-at-3]].
- **Fee-aware exits (Balanced).**
  - Reported P&L now subtracts the ~0.2% round-trip fee, so the trade log agrees with
    realized equity (previously overstated ~0.2%/trade).
  - Disabled the 0.15% partial take-profit — it fired below the fee floor = guaranteed loss.
  - Full TP 0.6%, trailing arms 0.4%/trails 0.2%, stop 0.8%.
- **Daily profit-lock / loss-stop.** Flatten all + halt for the day at +3% (lock green)
  or −3% (cap red), measured on total equity vs the day's opening equity; auto-resume next day.
- **Crash-guard regime filter.** Refuse new long entries (grid adds AND mean-reversion
  dip-entries) when a symbol falls >2% over 30 min. Directly targets the documented
  "buy dips that keep dipping → one big red day" failure mode ([[honest-assessment]]).
  Dormant until ~15 min of price history exists (no false blocks at startup).
- **Commits (local → pushed to GitHub):**
  - `2549765` Balanced profitability profile: fee-aware exits + daily profit-lock
  - `0d461a2` Add crash-guard regime filter: block new buys in sharp downtrends

**Verified live:** build + vet pass, both symbols streaming, grid buys $50 legs, honest P&L
wiring active, no errors, crash guard correctly dormant during warmup.

**Not yet observed live** (watching for these):
- [ ] Crash guard actually blocking a buy (needs a real −2%/30min move)
- [ ] Daily lock firing (needs equity to reach +3% / −3% on the day)

---

### 2026-07-14 — Capture upticks: event-driven exits + scale-out + ETH short fix

Prompted by BTC wicking to ~$64,181 without the bot selling. Root cause: exits ran on a
**3-second poll**, so fast wicks passed between samples (even the trailing, armed far lower,
never fired — proof the poll never sampled the spike).

- **Event-driven exits.** Take-profit / scale-out / trailing / stop now evaluate on **every
  trade tick** (`checkExits`), with the 3s poll kept as a backstop. A per-symbol guard
  (`beginExit`/`endExit`) prevents the poll and stream from double-selling.
- **Scale-out ladder.** Sell **50% at +0.5%** into strength (once per position via
  `scaleOutDone`), let the remainder run under the trailing stop; hard backstop at +1.0%.
  This is the "sell some into the uptick" behaviour that was missing.
- **ETH short-spam fixed.** Mean-reversion was emitting short entries on a spot paper
  account that can't short → endless `insufficient balance for sell`. Made mean-reversion
  **long-only**. Verified: spam count now 0.
- **Verified live:** build + vet pass, grid buys $50 BTC leg, ETH spam gone, no errors.
- **Not yet observed live:** an actual scale-out / trailing capture (needs a favourable move).
- Commits (local, not yet pushed): see `git log`.

---

## Live event journal

_(Appended as notable exits / crash guard / daily lock fire or the bot restarts.)_

- 2026-07-14 21:14 — Engine restarted on Balanced+crash-guard build. First grid leg @ $63,792.75.
- 2026-07-14 21:50 — Restarted on event-driven-exits build (fresh DB, $1000).
  First grid leg: 0.000783 BTC @ $63,828. Watcher re-armed for scale-out/trailing/TP/guard/lock.
