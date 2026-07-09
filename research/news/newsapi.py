import hashlib
from datetime import datetime, timezone
from typing import Optional

import requests


class NewsAPIIngestor:
    BASE_URL = "https://newsapi.org/v2/everything"

    def __init__(self, api_key: str, timeout: int = 10):
        self.api_key = api_key
        self.timeout = timeout
        self._seen: set[str] = set()

    def fetch(self, query: str = "cryptocurrency OR bitcoin OR ethereum") -> list[dict]:
        if not self.api_key:
            return []

        articles = []
        try:
            resp = requests.get(
                self.BASE_URL,
                params={
                    "q": query,
                    "language": "en",
                    "sortBy": "publishedAt",
                    "pageSize": 10,
                    "apiKey": self.api_key,
                },
                timeout=self.timeout,
            )
            resp.raise_for_status()
            data = resp.json()
        except Exception:
            return articles

        for article in data.get("articles", []):
            title = article.get("title", "")
            url = article.get("url", "")
            url_hash = hashlib.sha256(url.encode()).hexdigest()
            if url_hash in self._seen:
                continue
            self._seen.add(url_hash)

            published_str = article.get("publishedAt", "")
            try:
                published = datetime.fromisoformat(published_str.replace("Z", "+00:00"))
            except Exception:
                published = datetime.now(timezone.utc)

            articles.append({
                "title": title,
                "link": url,
                "source": article.get("source", {}).get("name", "NewsAPI"),
                "published": published,
                "summary": article.get("description", ""),
            })

        if len(self._seen) > 1000:
            self._seen = set(list(self._seen)[-500:])

        return articles
