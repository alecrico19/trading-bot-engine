from dataclasses import dataclass, field
from typing import Optional
import yaml
import os


@dataclass
class SignalRule:
    name: str
    conditions: list[dict]
    action: str
    confidence: float = 0.7
    ttl_seconds: int = 300


@dataclass
class Config:
    symbols: list[str] = field(default_factory=lambda: ["BTC/USDT", "ETH/USDT"])
    redis_host: str = "localhost"
    redis_port: int = 6379
    update_interval: int = 60
    signal_min_confidence: float = 0.6
    rsi_period: int = 14
    rsi_oversold: float = 30
    rsi_overbought: float = 70
    bb_period: int = 20
    bb_stddev: float = 2.0
    volume_spike_multiplier: float = 1.5
    newsapi_key: str = ""
    feed_urls: list[str] = field(default_factory=lambda: [
        "https://cointelegraph.com/rss",
        "https://www.coindesk.com/arc/outboundfeeds/rss/",
        "https://cryptoslate.com/feed/",
    ])
    signal_rules: list[SignalRule] = field(default_factory=list)


def load_config(path: str) -> Config:
    if not os.path.exists(path):
        return Config()

    with open(path) as f:
        raw = yaml.safe_load(f) or {}

    cfg = Config()
    if "symbols" in raw:
        cfg.symbols = raw["symbols"]
    if "redis" in raw:
        cfg.redis_host = raw["redis"].get("host", cfg.redis_host)
        cfg.redis_port = raw["redis"].get("port", cfg.redis_port)
    if "update_interval" in raw:
        cfg.update_interval = raw["update_interval"]
    if "signal_min_confidence" in raw:
        cfg.signal_min_confidence = raw["signal_min_confidence"]
    if "newsapi_key" in raw:
        cfg.newsapi_key = raw["newsapi_key"]
    if "feed_urls" in raw:
        cfg.feed_urls = raw["feed_urls"]

    for key in ("rsi_period", "rsi_oversold", "rsi_overbought", "bb_period", "bb_stddev", "volume_spike_multiplier"):
        if key in raw:
            setattr(cfg, key, raw[key])

    for rule_data in raw.get("signal_rules", []):
        cfg.signal_rules.append(SignalRule(
            name=rule_data.get("name", ""),
            conditions=rule_data.get("conditions", []),
            action=rule_data.get("action", "bias_neutral"),
            confidence=rule_data.get("confidence", 0.7),
            ttl_seconds=rule_data.get("ttl_seconds", 300),
        ))

    return cfg
