import uuid
from datetime import datetime, timezone
from typing import Optional

from research.config import Config, SignalRule
from research.technical.indicators import IndicatorResult


class SignalGenerator:
    def __init__(self, config: Config):
        self.config = config

    def generate(self, symbol: str, indicators: IndicatorResult, sentiment: dict) -> list[dict]:
        signals = []

        for rule in self.config.signal_rules:
            sig = self._evaluate_rule(symbol, rule, indicators, sentiment)
            if sig:
                signals.append(sig)

        if not signals:
            sig = self._rule_based_signal(symbol, indicators, sentiment)
            if sig:
                signals.append(sig)

        return signals

    def _evaluate_rule(self, symbol: str, rule: SignalRule, indicators: IndicatorResult, sentiment: dict) -> Optional[dict]:
        for cond in rule.conditions:
            for key, threshold in cond.items():
                if not self._check_condition(key, threshold, indicators, sentiment):
                    return None
        return self._make_signal(symbol, rule.action, confidence=rule.confidence, ttl_seconds=rule.ttl_seconds,
                                  reason=f"rule: {rule.name}")

    def _check_condition(self, key: str, threshold, indicators: IndicatorResult, sentiment: dict) -> bool:
        if key == "rsi_14" and indicators.rsi is not None:
            return indicators.rsi < threshold
        if key == "rsi_above" and indicators.rsi is not None:
            return indicators.rsi > threshold
        if key == "volume_ratio" and indicators.volume_ratio is not None:
            return indicators.volume_ratio > threshold
        if key == "sentiment_above":
            return sentiment.get("combined", 0) > threshold
        if key == "sentiment_below":
            return sentiment.get("combined", 0) < threshold
        if key == "price_below_bb_lower" and indicators.bb_lower is not None and indicators.last_close is not None:
            return indicators.last_close < indicators.bb_lower
        if key == "price_above_bb_upper" and indicators.bb_upper is not None and indicators.last_close is not None:
            return indicators.last_close > indicators.bb_upper
        return False

    def _rule_based_signal(self, symbol: str, indicators: IndicatorResult, sentiment: dict) -> Optional[dict]:
        rsi = indicators.rsi
        sentiment_score = sentiment.get("combined", 0)
        bb_upper = indicators.bb_upper
        bb_lower = indicators.bb_lower

        if rsi is None:
            return None

        factors = {
            "rsi": round(rsi, 1) if rsi else None,
            "sentiment_score": round(sentiment_score, 2),
            "volume_ratio": round(indicators.volume_ratio, 2) if indicators.volume_ratio else None,
            "bb_upper": round(bb_upper, 2) if bb_upper else None,
            "bb_lower": round(bb_lower, 2) if bb_lower else None,
            "sma_20": round(indicators.sma_20, 2) if indicators.sma_20 else None,
        }

        if rsi < self.config.rsi_oversold and sentiment_score > -0.3:
            confidence = min(0.85, 0.6 + (self.config.rsi_oversold - rsi) / 100 + max(0, sentiment_score * 0.3))
            return self._make_signal(symbol, "bias_long", confidence=confidence, factors=factors,
                                      reason=f"RSI oversold ({rsi:.1f})")

        if rsi > self.config.rsi_overbought and sentiment_score < 0.3:
            confidence = min(0.85, 0.6 + (rsi - self.config.rsi_overbought) / 100 + max(0, -sentiment_score * 0.3))
            return self._make_signal(symbol, "bias_short", confidence=confidence, factors=factors,
                                      reason=f"RSI overbought ({rsi:.1f})")

        if sentiment_score > 0.5:
            return self._make_signal(symbol, "bias_long", confidence=0.55 + sentiment_score * 0.3,
                                      factors=factors, reason=f"Positive sentiment ({sentiment_score:.2f})")

        if sentiment_score < -0.5:
            return self._make_signal(symbol, "bias_short", confidence=0.55 + abs(sentiment_score) * 0.3,
                                      factors=factors, reason=f"Negative sentiment ({sentiment_score:.2f})")

        if indicators.volume_ratio and indicators.volume_ratio > 3.0 and sentiment_score > 0.2:
            return self._make_signal(symbol, "bias_long", confidence=0.65, factors=factors,
                                      reason=f"Volume spike ({indicators.volume_ratio:.1f}x)")

        return None

    @staticmethod
    def _make_signal(symbol: str, action: str, confidence: float = 0.6, ttl_seconds: int = 300,
                     factors: dict = None, reason: str = "") -> dict:
        direction = "neutral"
        source = "research"
        sig_type = "bias"

        if action == "bias_long":
            direction = "long"
        elif action == "bias_short":
            direction = "short"
        elif action == "alert":
            sig_type = "alert"
            direction = "neutral"

        return {
            "id": f"sig_{uuid.uuid4().hex[:8]}",
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "source": source,
            "type": sig_type,
            "symbol": symbol,
            "direction": direction,
            "confidence": round(min(max(confidence, 0.0), 1.0), 2),
            "factors": factors or {},
            "reason": reason,
            "ttl_seconds": ttl_seconds,
        }
