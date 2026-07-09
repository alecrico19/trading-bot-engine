from vaderSentiment.vaderSentiment import SentimentIntensityAnalyzer
from textblob import TextBlob


class SentimentAnalyzer:
    def __init__(self):
        self._vader = SentimentIntensityAnalyzer()
        self._vader.lexicon.pop("BTC", None)
        self._vader.lexicon.pop("ETH", None)

    def analyze_headline(self, text: str) -> dict:
        vader_scores = self._vader.polarity_scores(text)

        try:
            blob = TextBlob(text)
            textblob_polarity = blob.sentiment.polarity
        except Exception:
            textblob_polarity = 0.0

        combined = (vader_scores["compound"] + textblob_polarity) / 2

        return {
            "vader_compound": vader_scores["compound"],
            "vader_pos": vader_scores["pos"],
            "vader_neg": vader_scores["neg"],
            "vader_neu": vader_scores["neu"],
            "textblob": textblob_polarity,
            "combined": combined,
            "label": self._label(combined),
        }

    def analyze_articles(self, articles: list[dict]) -> dict:
        if not articles:
            return {"combined": 0.0, "label": "neutral", "count": 0}

        scores = []
        total_weight = 0.0
        now = __import__("datetime").datetime.now(__import__("datetime").timezone.utc)

        for article in articles:
            text = article.get("title", "") + " " + article.get("summary", "")
            if not text.strip():
                continue
            result = self.analyze_headline(text)

            hours_ago = max((now - article["published"]).total_seconds() / 3600, 0)
            weight = max(0.1, 1.0 - hours_ago / 24.0)

            scores.append(result["combined"] * weight)
            total_weight += weight

        if total_weight == 0:
            return {"combined": 0.0, "label": "neutral", "count": 0}

        combined = sum(scores) / total_weight
        return {
            "combined": combined,
            "label": self._label(combined),
            "count": len(scores),
        }

    @staticmethod
    def _label(score: float) -> str:
        if score > 0.3:
            return "positive"
        if score < -0.3:
            return "negative"
        return "neutral"
