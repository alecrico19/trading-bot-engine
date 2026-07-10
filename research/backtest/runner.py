"""
Backtesting engine — walk-forward simulation of trading strategies against historical OHLCV data.
Uses numpy/pandas (zero heavy dependencies). Works on Python 3.14.
"""

import json
import sys
import time
from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Optional

import ccxt
import numpy as np
import pandas as pd


@dataclass
class BacktestResult:
    symbol: str
    strategy: str
    params: dict = field(default_factory=dict)
    total_trades: int = 0
    wins: int = 0
    losses: int = 0
    win_rate: float = 0.0
    total_return: float = 0.0
    max_drawdown: float = 0.0
    sharpe_ratio: float = 0.0
    profit_factor: float = 0.0
    avg_win: float = 0.0
    avg_loss: float = 0.0
    equity_curve: list[float] = field(default_factory=list)


def compute_rsi(close: np.ndarray, period: int = 14) -> np.ndarray:
    if len(close) < period + 1:
        return np.full(len(close), 50.0)
    changes = np.diff(close, prepend=close[0])
    gains = np.where(changes > 0, changes, 0)
    losses = np.where(changes < 0, -changes, 0)
    avg_gain = np.convolve(gains, np.ones(period) / period, mode="full")[:len(close)]
    avg_loss = np.convolve(losses, np.ones(period) / period, mode="full")[:len(close)]
    rsi = np.zeros(len(close))
    mask = avg_loss > 0
    rsi[mask] = 100 - (100 / (1 + avg_gain[mask] / avg_loss[mask]))
    rsi[~mask] = 100
    return rsi


def compute_bb(close: np.ndarray, period: int = 20, stddev: float = 2.0) -> tuple:
    if len(close) < period:
        return np.zeros(len(close)), np.zeros(len(close)), np.zeros(len(close))
    rolling = np.array([np.mean(close[max(0, i - period + 1):i + 1]) for i in range(len(close))])
    rolling_std = np.array([np.std(close[max(0, i - period + 1):i + 1]) for i in range(len(close))])
    upper = rolling + stddev * rolling_std
    lower = rolling - stddev * rolling_std
    return upper, rolling, lower


def compute_volume_ratio(volume: np.ndarray, short_period: int = 5, long_period: int = 20) -> np.ndarray:
    if len(volume) < long_period:
        return np.ones(len(volume))
    short_ma = np.array([np.mean(volume[max(0, i - short_period + 1):i + 1]) for i in range(len(volume))])
    long_ma = np.array([np.mean(volume[max(0, i - long_period + 1):i + 1]) for i in range(len(volume))])
    ratio = np.zeros(len(volume))
    mask = long_ma > 0
    ratio[mask] = short_ma[mask] / long_ma[mask]
    return ratio


def evaluate_scalping(df: pd.DataFrame, params: dict) -> np.ndarray:
    """
    Simplified scalping simulation using price action instead of order book.
    Uses micro-momentum: buy on upward micro-trend, sell on downward.
    Returns array of positions: 1=long, -1=short, 0=flat.
    """
    close = df["close"].values
    volume = df["volume"].values if "volume" in df.columns else np.ones(len(close))

    vol_ratio = compute_volume_ratio(volume)
    threshold = params.get("volume_spike_multiplier", 1.5)

    returns = np.diff(close, prepend=close[0])
    signals = np.zeros(len(close))

    for i in range(1, len(close)):
        if vol_ratio[i] > threshold and returns[i] > 0:
            signals[i] = 1
        elif vol_ratio[i] > threshold and returns[i] < 0:
            signals[i] = -1

    if params.get("use_smoothing", False):
        window = 3
        smoothed = np.zeros(len(signals))
        for i in range(len(signals)):
            start = max(0, i - window + 1)
            smoothed[i] = np.sign(np.sum(signals[start:i + 1]))
        signals = smoothed

    return signals


def evaluate_mean_reversion(df: pd.DataFrame, params: dict) -> np.ndarray:
    close = df["close"].values
    rsi_period = params.get("rsi_period", 14)
    rsi_oversold = params.get("rsi_oversold", 30)
    rsi_overbought = params.get("rsi_overbought", 70)
    bb_period = params.get("bb_period", 20)
    bb_stddev = params.get("bb_stddev", 2.0)
    stop_loss = params.get("stop_loss_pct", 0.05)

    rsi = compute_rsi(close, rsi_period)
    upper, middle, lower = compute_bb(close, bb_period, bb_stddev)

    signals = np.zeros(len(close))
    entry_price = 0.0
    position = 0

    for i in range(bb_period, len(close)):
        if position == 0:
            if rsi[i] < rsi_oversold and close[i] <= lower[i]:
                signals[i] = 1
                position = 1
                entry_price = close[i]
            elif rsi[i] > rsi_overbought and close[i] >= upper[i]:
                signals[i] = -1
                position = -1
                entry_price = close[i]
        elif position == 1:
            pnl_pct = (close[i] - entry_price) / entry_price
            if rsi[i] > rsi_overbought or close[i] >= upper[i] or pnl_pct <= -stop_loss:
                signals[i] = -1
                position = 0
                entry_price = 0
        elif position == -1:
            pnl_pct = (entry_price - close[i]) / entry_price
            if rsi[i] < rsi_oversold or close[i] <= lower[i] or pnl_pct <= -stop_loss:
                signals[i] = 1
                position = 0
                entry_price = 0

    return signals


def run_backtest(df: pd.DataFrame, signals: np.ndarray, initial_capital: float = 1000.0,
                 fee_pct: float = 0.001) -> BacktestResult:
    close = df["close"].values
    position = 0.0
    entry_price = 0.0
    trades = 0
    wins = 0
    losses = 0
    pnl_list = []
    capital = initial_capital
    equity_curve = [capital]
    peak = capital

    for i in range(len(close)):
        if signals[i] == 1 and position <= 0:
            if position < 0:
                pnl = (entry_price - close[i]) * abs(position)
                pnl -= abs(position) * entry_price * fee_pct + abs(position) * close[i] * fee_pct
                capital += abs(position) * entry_price + pnl
                trades += 1
                pnl_list.append(pnl)
                if pnl > 0:
                    wins += 1
                else:
                    losses += 1
                position = 0
            size = capital * 0.1
            position = size / close[i]
            capital -= size
            capital -= size * fee_pct
            entry_price = close[i]

        elif signals[i] == -1 and position > 0:
            pnl = (close[i] - entry_price) * position
            pnl -= position * entry_price * fee_pct + position * close[i] * fee_pct
            capital += position * close[i] - position * close[i] * fee_pct
            trades += 1
            pnl_list.append(pnl)
            if pnl > 0:
                wins += 1
            else:
                losses += 1
            position = 0

    if position != 0:
        pnl = (close[-1] - entry_price) * position
        pnl -= position * entry_price * fee_pct + position * close[-1] * fee_pct
        trades += 1
        pnl_list.append(pnl)
        if pnl > 0:
            wins += 1
        else:
            losses += 1

    final_equity = capital + position * close[-1]
    total_return = (final_equity - initial_capital) / initial_capital

    for i in range(len(close)):
        val = capital
        if position != 0:
            val += position * close[i]
        equity_curve.append(float(val))
        peak = max(peak, val)

    max_dd = 0.0
    for val in equity_curve:
        peak = max(peak, val)
        dd = (peak - val) / peak if peak > 0 else 0
        max_dd = max(max_dd, dd)

    returns_arr = np.diff(equity_curve) / equity_curve[:-1]
    returns_arr = returns_arr[~np.isnan(returns_arr)]
    sharpe = 0.0
    if len(returns_arr) > 1 and np.std(returns_arr) > 0:
        sharpe = float(np.mean(returns_arr) / np.std(returns_arr) * np.sqrt(365 * 24 * 12))

    avg_win = 0.0
    avg_loss = 0.0
    if wins > 0:
        avg_win = sum(p for p in pnl_list if p > 0) / wins
    if losses > 0:
        avg_loss = sum(p for p in pnl_list if p < 0) / losses

    profit_factor = 1.0
    total_win = sum(p for p in pnl_list if p > 0)
    total_loss = abs(sum(p for p in pnl_list if p < 0))
    if total_loss > 0:
        profit_factor = total_win / total_loss

    return BacktestResult(
        symbol="",
        strategy="",
        total_trades=trades,
        wins=wins,
        losses=losses,
        win_rate=wins / trades if trades > 0 else 0,
        total_return=total_return,
        max_drawdown=max_dd,
        sharpe_ratio=sharpe,
        profit_factor=profit_factor,
        avg_win=avg_win,
        avg_loss=avg_loss,
        equity_curve=equity_curve,
    )


def walk_forward(df: pd.DataFrame, strategy_fn, params: dict, window_size: int = 500, step_size: int = 100) -> BacktestResult:
    all_signals = np.zeros(len(df))
    for start in range(0, len(df) - window_size, step_size):
        end = start + window_size
        train_df = df.iloc[start:end]
        test_start = end
        test_end = min(test_start + step_size, len(df))
        test_df = df.iloc[test_start:test_end]

        train_signals = strategy_fn(train_df, params)
        test_signals = strategy_fn(test_df, params)
        all_signals[test_start:test_end] = test_signals

    return run_backtest(df, all_signals)


def optimize_params(df: pd.DataFrame, strategy_fn, param_grid: dict) -> list[dict]:
    results = []
    keys = list(param_grid.keys())
    values = list(param_grid.values())

    def recurse(idx: int, current: dict):
        if idx == len(keys):
            params = current.copy()
            result = run_backtest(df, strategy_fn(df, params))
            result.params = params
            result.strategy = strategy_fn.__name__
            results.append(result)
            return
        for val in values[idx]:
            current[keys[idx]] = val
            recurse(idx + 1, current)

    recurse(0, {})
    results.sort(key=lambda r: r.sharpe_ratio, reverse=True)
    return results


def fetch_ohlcv(symbol: str, timeframe: str = "5m", limit: int = 1000) -> pd.DataFrame:
    exchange = ccxt.binance({"enableRateLimit": True})
    ohlcv = exchange.fetch_ohlcv(symbol, timeframe=timeframe, limit=limit)
    df = pd.DataFrame(ohlcv, columns=["timestamp", "open", "high", "low", "close", "volume"])
    df["timestamp"] = pd.to_datetime(df["timestamp"], unit="ms")
    return df


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="Backtest trading strategies")
    parser.add_argument("--symbols", nargs="+", default=["BTC/USDT", "ETH/USDT"])
    parser.add_argument("--strategy", choices=["scalping", "mean-reversion", "all"], default="all")
    parser.add_argument("--optimize", action="store_true", help="Grid-search optimal params")
    parser.add_argument("--days", type=int, default=7, help="Days of historical data")
    parser.add_argument("--json", action="store_true", help="Output as JSON")
    args = parser.parse_args()

    limit = (args.days * 24 * 60) // 5

    strategies = {
        "scalping": (evaluate_scalping, {
            "volume_spike_multiplier": [1.2, 1.5, 2.0, 3.0],
            "use_smoothing": [False, True],
        }),
        "mean-reversion": (evaluate_mean_reversion, {
            "rsi_period": [10, 14, 20],
            "rsi_oversold": [20, 25, 30, 35],
            "rsi_overbought": [65, 70, 75, 80],
            "bb_period": [14, 20, 26],
            "bb_stddev": [1.5, 2.0, 2.5],
            "stop_loss_pct": [0.03, 0.05, 0.07],
        }),
    }

    all_results = []

    for symbol in args.symbols:
        print(f"\n=== {symbol} ({args.days}d history, {limit} candles) ===", file=sys.stderr)
        df = fetch_ohlcv(symbol, limit=limit)
        print(f"  Fetched {len(df)} candles", file=sys.stderr)

        for name, (fn, grid) in strategies.items():
            if args.strategy != "all" and args.strategy != name:
                continue

            if args.optimize:
                print(f"  Optimizing {name} ({len(grid)} params, combos)...", file=sys.stderr)
                results = optimize_params(df, fn, grid)
                for r in results[:3]:
                    r.symbol = symbol
                    r.strategy = name
                    all_results.append(r)
                    print(f"    {r.params} → Sharpe={r.sharpe_ratio:.2f} Return={r.total_return*100:.1f}% "
                          f"WR={r.win_rate*100:.0f}% DD={r.max_drawdown*100:.1f}% PF={r.profit_factor:.2f}",
                          file=sys.stderr)
            else:
                result = run_backtest(df, fn(df, {}))
                result.symbol = symbol
                result.strategy = name
                all_results.append(result)
                print(f"  {name}: Sharpe={result.sharpe_ratio:.2f} Return={result.total_return*100:.1f}% "
                      f"WR={result.win_rate*100:.0f}% Trades={result.total_trades} DD={result.max_drawdown*100:.1f}%",
                      file=sys.stderr)

    if args.json and all_results:
        output = []
        for r in all_results:
            output.append({
                "symbol": r.symbol,
                "strategy": r.strategy,
                "params": r.params,
                "total_trades": r.total_trades,
                "wins": r.wins,
                "losses": r.losses,
                "win_rate": round(r.win_rate, 3),
                "total_return": round(r.total_return, 4),
                "max_drawdown": round(r.max_drawdown, 4),
                "sharpe_ratio": round(r.sharpe_ratio, 2),
                "profit_factor": round(r.profit_factor, 2),
                "avg_win": round(r.avg_win, 2),
                "avg_loss": round(r.avg_loss, 2),
            })
        print(json.dumps(output, indent=2))
