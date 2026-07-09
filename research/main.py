#!/usr/bin/env python3
"""
Research Service — pulls news, computes sentiment and technical indicators,
emits signals to Redis for the execution engine.
"""

import logging
import signal
import sys
import time
from pathlib import Path

import ccxt
import pandas as pd

from research.config import load_config
from research.news.rss import RSSIngestor
from research.news.newsapi import NewsAPIIngestor
from research.news.reddit import RedditIngestor
from research.sentiment.vader import SentimentAnalyzer
from research.technical.indicators import TechnicalAnalyzer
from research.technical.signals import SignalGenerator
from research.publisher import SignalPublisher

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(name)s: %(message)s",
    stream=sys.stderr,
)
logger = logging.getLogger("research")


def main():
    config_path = sys.argv[1] if len(sys.argv) > 1 else "config.yaml"
    config = load_config(config_path)

    publisher = SignalPublisher(host=config.redis_host, port=config.redis_port)
    publisher.connect()

    rss = RSSIngestor(config.feed_urls)
    newsapi = NewsAPIIngestor(config.newsapi_key)
    reddit = RedditIngestor()
    sentiment = SentimentAnalyzer()
    tech = TechnicalAnalyzer(
        rsi_period=config.rsi_period,
        bb_period=config.bb_period,
        bb_stddev=config.bb_stddev,
    )
    signal_gen = SignalGenerator(config)

    exchange = ccxt.binance({"enableRateLimit": True})

    running = True

    def shutdown(sig, frame):
        nonlocal running
        logger.info("received signal %s, shutting down", sig)
        running = False

    signal.signal(signal.SIGINT, shutdown)
    signal.signal(signal.SIGTERM, shutdown)

    logger.info("research service started (symbols=%s, interval=%ds)",
                config.symbols, config.update_interval)

    while running:
        try:
            articles = rss.fetch()
            logger.debug("rss: %d articles", len(articles))

            if config.newsapi_key:
                api_articles = newsapi.fetch()
                articles.extend(api_articles)
                logger.debug("newsapi: %d articles", len(api_articles))

            reddit_articles = reddit.fetch()
            articles.extend(reddit_articles)
            logger.debug("reddit: %d articles", len(reddit_articles))

            sentiment_result = sentiment.analyze_articles(articles)
            logger.info("sentiment: combined=%.2f label=%s (from %d articles)",
                        sentiment_result["combined"], sentiment_result["label"], sentiment_result["count"])

            for symbol in config.symbols:
                try:
                    ohlcv = exchange.fetch_ohlcv(symbol, timeframe="5m", limit=100)
                    df = pd.DataFrame(ohlcv, columns=["timestamp", "open", "high", "low", "close", "volume"])
                    indicators = tech.compute(df)
                    logger.debug("%s: rsi=%.1f bb=(%.2f, %.2f) vol_ratio=%.2f",
                                 symbol, indicators.rsi or 0,
                                 indicators.bb_lower or 0, indicators.bb_upper or 0,
                                 indicators.volume_ratio or 0)

                    signals = signal_gen.generate(symbol, indicators, sentiment_result)
                    for sig in signals:
                        publisher.publish(sig)
                except Exception as e:
                    logger.error("%s fetch failed: %s", symbol, e)

        except Exception as e:
            logger.error("loop error: %s", e)

        for _ in range(config.update_interval):
            if not running:
                break
            time.sleep(1)

    publisher.close()
    logger.info("research service stopped")


if __name__ == "__main__":
    main()
