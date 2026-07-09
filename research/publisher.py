import json
import logging
from typing import Optional

logger = logging.getLogger(__name__)


class SignalPublisher:
    def __init__(self, host: str = "localhost", port: int = 6379):
        self.host = host
        self.port = port
        self._redis = None

    def connect(self):
        try:
            import redis
            self._redis = redis.Redis(host=self.host, port=self.port, db=0,
                                       socket_connect_timeout=3, decode_responses=True)
            self._redis.ping()
            logger.info("connected to redis at %s:%d", self.host, self.port)
        except Exception as e:
            logger.warning("redis not available (%s), signals will be logged only", e)
            self._redis = None

    def publish(self, signal: dict) -> bool:
        payload = json.dumps(signal)
        logger.info("signal: %s", payload)

        if self._redis:
            try:
                self._redis.publish("trading:signals", payload)
                return True
            except Exception as e:
                logger.error("redis publish failed: %s", e)
        return False

    def close(self):
        if self._redis:
            try:
                self._redis.close()
            except Exception:
                pass
