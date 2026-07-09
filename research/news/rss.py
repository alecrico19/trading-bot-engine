import hashlib
from datetime import datetime, timezone
from typing import Optional

import feedparser
import requests


class RSSIngestor:
    def __init__(self, urls: list[str], timeout: int = 10):
        self.urls = urls
        self.timeout = timeout
        self._seen: set[str] = set()

    def fetch(self) -> list[dict]:
        articles = []
        for url in self.urls:
            try:
                feed = feedparser.parse(url)
            except Exception:
                continue

            for entry in feed.entries[:10]:
                title = entry.get("title", "")
                link = entry.get("link", "")
                url_hash = hashlib.sha256(link.encode()).hexdigest()
                if url_hash in self._seen:
                    continue
                self._seen.add(url_hash)

                published = None
                if hasattr(entry, "published_parsed") and entry.published_parsed:
                    try:
                        published = datetime(*entry.published_parsed[:6], tzinfo=timezone.utc)
                    except Exception:
                        published = datetime.now(timezone.utc)
                else:
                    published = datetime.now(timezone.utc)

                articles.append({
                    "title": title,
                    "link": link,
                    "source": feed.feed.get("title", url),
                    "published": published,
                    "summary": entry.get("summary", ""),
                })

            if len(self._seen) > 1000:
                self._seen = set(list(self._seen)[-500:])

        return articles
