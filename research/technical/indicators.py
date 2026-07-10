import numpy as np
import pandas as pd
from dataclasses import dataclass
from typing import Optional


@dataclass
class IndicatorResult:
    rsi: Optional[float] = None
    bb_upper: Optional[float] = None
    bb_middle: Optional[float] = None
    bb_lower: Optional[float] = None
    volume_ratio: Optional[float] = None
    sma_20: Optional[float] = None
    sma_50: Optional[float] = None
    macd: Optional[float] = None
    macd_signal: Optional[float] = None
    last_close: Optional[float] = None


class TechnicalAnalyzer:
    def __init__(self, rsi_period: int = 14, bb_period: int = 20, bb_stddev: float = 2.0):
        self.rsi_period = rsi_period
        self.bb_period = bb_period
        self.bb_stddev = bb_stddev

    def compute(self, df: pd.DataFrame) -> IndicatorResult:
        if df.empty or len(df) < self.bb_period:
            return IndicatorResult()

        close = df["close"].astype(float).values
        volume = df["volume"].astype(float).values if "volume" in df.columns else None

        result = IndicatorResult()
        result.last_close = float(close[-1]) if len(close) > 0 else None
        result.rsi = self._rsi(close, self.rsi_period)
        bb = self._bollinger_bands(close, self.bb_period, self.bb_stddev)
        result.bb_upper = bb[0]
        result.bb_middle = bb[1]
        result.bb_lower = bb[2]
        result.sma_20 = self._sma(close, 20)
        result.sma_50 = self._sma(close, 50) if len(close) >= 50 else None
        macd = self._macd(close)
        result.macd = macd[0]
        result.macd_signal = macd[1]

        if volume is not None and len(volume) >= 20:
            recent_vol = np.mean(volume[-5:])
            avg_vol = np.mean(volume[-20:])
            if avg_vol > 0:
                result.volume_ratio = recent_vol / avg_vol

        return result

    @staticmethod
    def _rsi(prices: np.ndarray, period: int) -> float:
        if len(prices) < period + 1:
            return 50.0

        changes = np.diff(prices[-period - 1:])
        gains = np.where(changes > 0, changes, 0)
        losses = np.where(changes < 0, -changes, 0)

        avg_gain = np.mean(gains)
        avg_loss = np.mean(losses)

        if avg_loss == 0:
            return 100.0

        rs = avg_gain / avg_loss
        return float(100 - (100 / (1 + rs)))

    @staticmethod
    def _bollinger_bands(prices: np.ndarray, period: int, stddev: float) -> tuple:
        window = prices[-period:]
        middle = float(np.mean(window))
        std = float(np.std(window, ddof=0))
        upper = middle + stddev * std
        lower = middle - stddev * std
        return upper, middle, lower

    @staticmethod
    def _sma(prices: np.ndarray, period: int) -> float:
        return float(np.mean(prices[-period:]))

    @staticmethod
    def _macd(prices: np.ndarray) -> tuple:
        if len(prices) < 26:
            return None, None

        ema12_series = TechnicalAnalyzer._ema_series(prices, 12)
        ema26_series = TechnicalAnalyzer._ema_series(prices, 26)
        if ema12_series is None or ema26_series is None:
            return None, None

        macd_series = ema12_series - ema26_series
        macd_line = float(macd_series[-1])

        # 9-period EMA of MACD series (signal line)
        if len(macd_series) < 9:
            return macd_line, None

        multiplier = 2 / (9 + 1)
        signal = float(np.mean(macd_series[:9]))
        for val in macd_series[9:]:
            signal = (val - signal) * multiplier + signal

        return macd_line, signal

    @staticmethod
    def _ema_series(prices: np.ndarray, period: int) -> Optional[np.ndarray]:
        if len(prices) < period:
            return None
        ema = np.zeros(len(prices))
        ema[period - 1] = np.mean(prices[:period])
        multiplier = 2 / (period + 1)
        for i in range(period, len(prices)):
            ema[i] = (prices[i] - ema[i - 1]) * multiplier + ema[i - 1]
        ema[:period - 1] = ema[period - 1]
        return ema

    @staticmethod
    def _ema(prices: np.ndarray, period: int) -> Optional[float]:
        if len(prices) < period:
            return None
        multiplier = 2 / (period + 1)
        ema = float(np.mean(prices[:period]))
        for price in prices[period:]:
            ema = (price - ema) * multiplier + ema
        return ema
