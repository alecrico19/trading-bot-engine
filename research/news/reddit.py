import hashlib
from datetime import datetime, timezone

import requests


class RedditIngestor:
    BASE_URL = "https://www.reddit.com"

    def __init__(self, subreddits: list[str] = None, timeout: int = 10):
        self.subreddits = subreddits or ["CryptoCurrency", "BitcoinMarkets"]
        self.timeout = timeout
        self._seen: set[str] = set()
        self._session = requests.Session()
        self._session.headers["User-Agent"] = "TradingBot/1.0"

    def fetch(self) -> list[dict]:
        articles = []
        for sub in self.subreddits:
            try:
                resp = self._session.get(
                    f"{self.BASE_URL}/r/{sub}/hot.json",
                    params={"limit": 10},
                    timeout=self.timeout,
                )
                resp.raise_for_status()
                data = resp.json()
            except Exception:
                continue

            for child in data.get("data", {}).get("children", []):
                post = child.get("data", {})
                title = post.get("title", "")
                permalink = post.get("permalink", "")
                url = f"https://reddit.com{permalink}"

                url_hash = hashlib.sha256(url.encode()).hexdigest()
                if url_hash in self._seen:
                    continue
                self._seen.add(url_hash)

                created = post.get("created_utc", 0)
                try:
                    published = datetime.fromtimestamp(created, tz=timezone.utc)
                except Exception:
                    published = datetime.now(timezone.utc)

                articles.append({
                    "title": title,
                    "link": url,
                    "source": f"r/{sub}",
                    "published": published,
                    "summary": post.get("selftext", "")[:500],
                })

        if len(self._seen) > 1000:
            self._seen = set(list(self._seen)[-500:])

        return articles
