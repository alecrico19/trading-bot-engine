# Code Audit Report — Trading Bot

## Critical Bugs (Fix First)

### 1. Config env-var substitution broken (`config.go:80-90`)
`$VAR` and `${VAR}` patterns BOTH fail to resolve — API keys become empty strings.
- `"$VAR"` → TrimPrefix("${") does nothing → envKey stays `"$VAR"` → `os.Getenv("$VAR")` → `""`
- `"${VAR}"` → TrimPrefix("${") gives `"VAR}"` → TrimPrefix("}") doesn't match → envKey stays `"VAR}"` → `os.Getenv("VAR}")` → `""`
**Fix:** Use `os.ExpandEnv()` or a proper resolver.

### 2. Telegram bot open to anyone (`bot.py:32-38`)
Empty `AUTHORIZED_USERS` → `authorized is None` → `is_authorized()` returns `True` for ALL users.
**Fix:** Default to deny — empty set blocks all commands until configured.

### 3. HTTP API no authentication (`server.go`)
`/pause`, `/resume`, `/kill`, `/status`, `/positions`, `/balance` — zero auth. Anyone on the network can control the bot.
**Fix:** Add `X-API-Token` header check. Bind to `127.0.0.1` only.

### 4. MACD signal line completely broken (`indicators.py:84-102`)
`macd_signal` always equals `macd` — the entire MACD indicator is useless.
**Fix:** Compute full MACD series, then 9-period EMA of it.

### 5. BB conditions always return True (`signals.py:47-50`)
`price_below_bb_lower` and `price_above_bb_upper` return `True` whenever BB is computable — never compares price to band.
**Fix:** Add `last_close` to IndicatorResult, compare it to band values.

### 6. Backtest can never short (`runner.py:146-188`)
No branch opens a short position. Short signals from mean-reversion are silently ignored. Win rate / Sharpe are biased.
**Fix:** Add short-entry and short-close branches.

### 7. Backtest equity curve is wrong (`runner.py:156-207`)
Built post-hoc from FINAL state, not bar-by-bar. Curve is either flat (position closed) or fictitious (retroactively applies final position to all bars). Sharpe and max drawdown are wrong.
**Fix:** Record `equity_curve.append(capital + position * close[i])` inside the trade loop.

### 8. Binance WebSocket panics on shutdown (`binance/adapter.go:286-298`)
`stop <- struct{}{}` can send on a closed channel → panic. `done` channel ignored → no reconnection on stream termination.
**Fix:** Capture `done`, use non-blocking send, implement reconnect with backoff.

### 9. Binance nil-pointer on unknown order type (`binance/adapter.go:125-192`)
Unhandled `orderType` → `result` stays nil → `result.OrderID` panics.
**Fix:** Add `default:` case returning error.

### 10. CancelOrder swaps arguments (`order/manager.go:68-76`)
Calls `CancelOrder(ctx, order.ExchangeID, symbol)` — symbol and orderID are reversed. Every cancellation fails.
**Fix:** `CancelOrder(ctx, symbol, order.ExchangeID)`.

### 11. Short position flip discarded (`order/manager.go:156-163`)
Selling more than the long amount resets `p.Amount = 0` instead of keeping the negative short. Oversold quantity vanishes.
**Fix:** Only reset AvgPrice when fully flat, keep `p.Amount` as-is when flipping.

---

## High Severity

| # | File | Issue |
|---|------|-------|
| 12 | `engine.go:373` | `tradeCount++` data race — no mutex on increment |
| 13 | `engine.go:645/667` | Stop-loss orders bypass all trade/PnL/risk bookkeeping |
| 14 | `engine.go:645/667` | Stop-loss PlaceOrder error ignored → position "forgotten" while still open |
| 15 | `engine.go:380` | `entryPrice` reset to 0 on partial/rejected sells → stop-loss blind to remaining position |
| 16 | `engine.go:456-464` | `calculateEquity` ignores crypto holdings (used for sizing & risk) |
| 17 | `engine.go:334-343` | Amount-0 SELL decisions open unmanaged shorts |
| 18 | `scalping.go:71-101` | Strong-imbalance branches unreachable (dead code — ratio>3.0 implies ratio>1.5) |
| 19 | `scalping.go:57` | `MinSpreadPct` logic inverted — default 0 makes scalping near-inert |
| 20 | `paper.go:140-144` | Rejected orders still report Filled and fill price |
| 21 | `consumer.go:64` | Signals with `TTLSeconds==0` treated as expired on arrival |
| 22 | `binance/adapter.go:396-400` | `parseFloat` via `fmt.Sscanf` loses precision, swallows errors |
| 23 | `engine.go:653-660` | `highWater` not reset on buy → stale trailing baseline |
| 24 | `telegram/bot.py:60` | Blocking `requests` in async handlers freezes event loop |
| 25 | `main.go:160-162` | `signalCon.Close()` before `<-engineDone` → use-after-close race |
| 26 | `main.go:143-154` | If TUI exits without ctx cancel, process hangs forever |
| 27 | `paper.go:178-184` | `splitSymbol("BTCUSDT")` returns `("BTCUSDT","USDT")` — Binance symbols without `/` break |
| 28 | `risk/manager.go:58-78` | Daily P&L reset only in `RecordTrade`, not `CanOpenPosition` → first batch of new day skips loss limit |
| 29 | `risk/manager.go:138-156` | `ValidateOrder` market-order notional branch is dead code (overridden by `if price > 0`) |

---

## Medium Severity

| # | File | Issue |
|---|------|-------|
| 30 | `engine.go:113-235` | `tickerCh` has no reader — dead plumbing |
| 31 | `engine.go:176-178` | "Using polling" logged but no polling exists — symbol silently inactive |
| 32 | `mean_reversion.go:46-51` | `prices` map + `PriceHistory` unsynchronized (latent race) |
| 33 | `strategy.go:78-100` | RSI is SMA-based (Cutler's), not Wilder's — diverges from exchange RSI |
| 34 | `engine.go:606-674` | TOCTOU race: stop-loss vs decision execution can double-close |
| 35 | `consumer.go:69-78` | Signal cache unbounded growth — no LRU or periodic sweep |
| 36 | `binance/adapter.go:421-441` | `parseOrderSide` defaults to `SideBuy` for unknown values |
| 37 | `binance/adapter.go:443-448` | `tradeSide` inverted — `IsBuyerMaker==true` means sell-initiated, not buy |
| 38 | `runner.py:254-263` | `walk_forward` computes `train_signals` and discards it (not real WFO) |
| 39 | `runner.py:268-287` | `optimize_params` is pure in-sample → overfitting |
| 40 | `dashboard.html:160` | Hardcoded `localhost:8080` breaks cross-host; use `window.location.origin` |
| 41 | `signals.py:63-68` | `if rsi` truthiness drops legitimate `0.0` values |

---

## Low Severity

| # | File | Issue |
|---|------|-------|
| 42 | `strategy.go:145-153` | Custom `sqrt` risks overflow; use `math.Sqrt` |
| 43 | `strategy.go:124-143` | Local `min`/`max` shadow Go 1.21+ builtins |
| 44 | `order/manager.go:199-204` | Local `abs` reimplemented instead of `math.Abs` |
| 45 | `risk/manager.go:123-128` | Same local `abs` duplication |
| 46 | `config.go:96-133` | `ApplyDefaults` can't distinguish unset from explicitly-0 |
| 47 | `config.go:135-145` | `Validate` only checks 2 fields — no sanity checks on strategy params |
| 48 | `signals.py:111` | Signal ID uses only 8 hex chars (32-bit) — collision risk after ~65k signals |
| 49 | `bot.py:167` | Hardcoded `$-50` alert threshold — not scaled to account size |

---

## Improvement Suggestions (Non-Bug)

1. **Reconnect WebSocket streams** — Currently no reconnection when Binance WS disconnects. Add backoff retry.
2. **Per-strategy position tracking** — Currently one `entryPrice` per symbol. Multiple strategies trading the same symbol collide.
3. **Use order manager's `Position.AvgPrice`** instead of separate `entryPrice` map for P&L calc.
4. **Add `math.Sqrt`** instead of custom Newton's method — hardware instruction, edge cases handled.
5. **Add CORS to HTTP API** for future web dashboard deployment on other hosts.
6. **Add `http.Server` with timeouts** to prevent slowloris attacks.
7. **Health check endpoint** (`GET /health`) for monitoring.
8. **Graceful HTTP shutdown** — Currently `ListenAndServe` is never shut down cleanly.
9. **Filter expired signals periodically** instead of only on new signal arrival.
10. **Detect flat-line prices** — RSI returns 100 for flat windows instead of 50 (neutral).

---

## Fix Priority

**Immediate (can crash or lose money):**
- #1 (config env vars)
- #8 (WS panics)
- #9 (nil-pointer panic)
- #10 (CancelOrder args)
- #11 (short flip lost)
- #2 (Telegram open)
- #3 (HTTP API open)

**Before going live:**
- #12-16 (engine races & P&L tracking)
- #17-19 (scalping logic bugs)
- #21 (signal TTL)
- #25-26 (shutdown races)
- #27 (symbol parsing)

**Before relying on backtest results:**
- #4-7 (MACD, BB conditions, short positions, equity curve)
- #38-39 (walk-forward, optimizer)