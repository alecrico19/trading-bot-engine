# Backtesting Plan

## Why backtesting next

Currently you have no way to know if a strategy config is profitable before running it live. Backtesting:
- Runs 2 weeks of historical data in ~30 seconds
- Outputs: Sharpe ratio, max drawdown, win rate, profit factor, total return
- Lets you compare configs: "RSI 30 vs 35 — which had better risk-adjusted returns?"

## Approach

### 1. Clean up research backtest stub (`research/backtest/`)
Already has directory, no code. Wire up:

- **Data source:** ccxt `fetch_ohlcv()` — Binance historical 5m candles, free
- **Engine:** vectorbt `Portfolio.from_signals()` — portfolio-level simulation
- **Walk-forward:** train on N days, test on N+1, repeat for robustness
- **Output:** JSON with metrics per strategy config

### 2. Add CLI to research service

```bash
python3 research/backtest/runner.py --symbols BTC/USDT,ETH/USDT --days 14
```

Outputs:
```json
{
  "scalping": {"sharpe": 1.2, "max_dd": 0.04, "win_rate": 0.45, "profit_factor": 1.8},
  "mean-reversion": {"sharpe": 0.3, "max_dd": 0.08, "win_rate": 0.35, "profit_factor": 0.9}
}
```

### 3. Strategy config optimization
Grid-search over parameter ranges:
- RSI oversold: 20, 25, 30, 35
- RSI overbought: 65, 70, 75, 80
- BB period: 14, 20, 26
- Stop-loss: 3%, 5%, 7%
Find the config with the best Sharpe ratio.

### 4. Dashboard integration
Add `/backtest` endpoint that returns latest results, show summary card on dashboard.

## Files

| File | What |
|------|------|
| `research/backtest/runner.py` | Backtest engine with walk-forward |
| `research/backtest/optimizer.py` | Grid-search optimizer |
| `research/requirements.txt` | Add vectorbt (check if available on py3.14) |
| `scripts/start.sh` | Add `backtest` command |

~150 lines, 30 min.

## Risk
- vectorbt may not install on Python 3.14 (same numba issue as pandas-ta). Fallback: manual portfolio simulation with numpy, same as we did for indicators.
