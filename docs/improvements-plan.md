# Improvements Plan

## 1. Stop-loss (small)

**What:** Auto-close paper positions losing >5%. Prevents drawdown spirals.

**How:**
- New goroutine in engine: checks all holdings every 3 seconds
- For each non-USDT balance, computes P&L% from last fill price vs current ticker
- P&L < -5% → sends market sell/buy to close
- Threshold configurable: `stop_loss_pct: 0.05` in config

**Files:** `engine.go` (+25 lines), `config.go` (+1 field)

## 2. Strategy P&L Breakdown (small)

**What:** Dashboard shows P&L per strategy ("scalping: +$2.40, mean-reversion: -$1.80")

**How:**
- Track `strategyPnL map[string]float64` in engine
- Increment on every `executeDecision` 
- Expose in `/status` as `strategy_pnl: {scalping: 2.40, mean-reversion: -1.80}`
- Dashboard shows a card with per-strategy P&L rows

**Files:** `engine.go` (+15 lines), `server.go` (+1 field), `dashboard.html` (+15 lines)

## 3. Trade History Table (small)

**What:** Scrolling last 20 trades in dashboard (symbol, side, price, P&L, strategy, time)

**How:**
- Add `GET /trades` endpoint returning last 20 rows from SQLite
- Dashboard polls it every 5 seconds
- Compact table at bottom of page with color-coded P&L

**Files:** `server.go` (+8 lines), `dashboard.html` (+40 lines)

## 4. Trailing Stop-loss (medium)

**What:** Locks in profits. Once position is +2%, trail a stop at X% below the highest price.

**How:**
- Track `highWaterMark map[string]float64` per symbol in engine
- In stop-loss goroutine: if position P&L > +2%, activate trail
- Trail = highWaterMark × (1 - trail_pct) — if price drops to trail, close
- Configurable: `trailing_stop_activate_pct: 0.02`, `trailing_stop_distance_pct: 0.03`

**Files:** `engine.go` (+40 lines), `config.go` (+2 fields)

## 5. Auto-start on Boot (medium)

**What:** Bot starts automatically when your machine boots. No manual launcher.

**How:**
- Create `trading-bot.service` in `~/.config/systemd/user/`
- Runs `~/trading-bot/scripts/launcher.sh` headless
- Enables with `systemctl --user enable trading-bot`
- Logs to journald

**Files:** 1 new file (`trading-bot.service`, ~15 lines)

## Effort Summary

| Item | Lines | Time |
|------|-------|------|
| 1. Stop-loss | ~25 | 10 min |
| 2. Strategy P&L | ~30 | 10 min |
| 3. Trade history | ~50 | 15 min |
| 4. Trailing stop | ~40 | 15 min |
| 5. Systemd service | ~15 | 5 min |
| **Total** | **~160** | **~1 hr** |
