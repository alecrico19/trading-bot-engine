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

        ema12 = TechnicalAnalyzer._ema(prices, 12)
        ema26 = TechnicalAnalyzer._ema(prices, 26)
        if ema12 is None or ema26 is None:
            return None, None

        macd_line = ema12 - ema26

        macd_prices = np.array([macd_line])
        for i in range(len(prices) - 1):
            prev = macd_prices[-1]
            macd_prices = np.append(macd_prices, prev + (2 / (9 + 1)) * (prev - prev))

        signal = float(np.mean(macd_prices[-9:])) if len(macd_prices) >= 9 else None
        return macd_line, signal

    @staticmethod
    def _ema(prices: np.ndarray, period: int) -> Optional[float]:
        if len(prices) < period:
            return None
        multiplier = 2 / (period + 1)
        ema = float(np.mean(prices[:period]))
        for price in prices[period:]:
            ema = (price - ema) * multiplier + ema
        return ema
