# Remaining Issues — Priority Stack

## P0: Disable time-exit (burning $1.41/hour)
Replace 90s time-based exit with take-profit (1-2%) + trailing stop (3%) as the only automated exits. Scalping reverse signals + take-profit = net positive.

## P0: P&L visibility — show more decimals
Dashboard formats P&L as `.toFixed(2)` → $0.012 shows as $0.01. $0.004 shows as $0.00. Change to `.toFixed(4)` so users can see micro-profits and realize they're not at $0.00.

## P1: Fresh DB on restart
The API returns 28+ trades from mixed sessions. User sees old pre-fix trades with real $0.00 P&L. Launcher should clean DB on each start or rotate logs.

## P1: Mean reversion verification
Has 0 trades across all sessions. Needs ~25 min to warm up (100 candles at 15s). Should be live by now. Need to verify it's evaluating and not stuck.

## P2: Config-based time-exit toggle
Add `time_exit_enabled: true/false` to config.yaml. Lets users turn it on/off without code changes.

## P2: Auto-adjust time-exit threshold based on trade performance
If time-exit P&L is negative, increase the time threshold. If positive, keep it.

## Files
| Item | File | Lines |
|------|------|-------|
| Disable time-exit | `engine.go` | Comment out the time-exit block (~8 lines) |
| P&L decimals | `dashboard.html` | 4 places instead of 2 (~2 lines) |
| Fresh DB | `launcher.sh` | rm trading.db before starting (~1 line) |
| Config toggle | `config.go` + `config.yaml` | Add field (~4 lines) |

## Recommendation
**Do P0 items only.** Disable time-exit → stop the bleeding. Fix P&L display → user sees real numbers. The rest can wait.