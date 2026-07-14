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

### 2026-07-14 — BUG FOUND & FIXED: exits filled at entry, not market price

The very first scale-out (ETH, triggered at +0.65% / $1879.78) filled at **$1865.69 =
entry × 0.999**, booking a +0.65% winner as a −$0.075 loss. Root cause: exit orders passed
`entry*0.999` as the order price, and the paper trader's `simulateFill` falls back to the
order price when its order-book snapshot for that symbol is stale/empty (ETH's was). So
**every exit was filling at the entry price and discarding the actual move** — a pre-existing
bug that only became visible once event-driven scale-out started firing.

- **Fix:** thread the live trigger `price` into `exitMarket` and the scale-out order (was
  `entry*0.999`). Exits now fill near current market whether or not the book snapshot is fresh.
- Verified: build + vet pass, clean restart, grid buy @ $63,721. Correct fill to be confirmed
  on the next live exit (watcher re-armed).

---

## Live event journal

_(Appended as notable exits / crash guard / daily lock fire or the bot restarts.)_

- 2026-07-14 21:14 — Restarted on Balanced+crash-guard build. First grid leg @ $63,792.75.
- 2026-07-14 21:50 — Restarted on event-driven-exits build. First grid leg @ $63,828.
- 2026-07-14 22:06 — ⚠️ First scale-out (ETH) fired but filled at entry×0.999 → exposed the
  exit-fill bug above.
- 2026-07-14 22:11 — Restarted on exit-fill-price-fix build (fresh DB, $1000).
  First grid leg: 0.000785 BTC @ $63,721.39. Watcher re-armed.
- 2026-07-14 22:18 — ✅ FIX CONFIRMED. ETH ladder ran clean: scale-out sold 50% at
  +0.56% ($1870.71, filled $1870.71, +$0.089), remainder ran to +1.04% and hit the full-TP
  backstop ($1879.78, filled $1879.78, +$0.210). Fills now match trigger prices; both
  tranches net-positive. Equity $1000.35, 2 wins. This is the scale-out-into-strength +
  let-the-rest-run behaviour working as designed. Watcher narrowed back to crash-guard /
  daily-lock only (exits now fire routinely — no need to notify per exit).
