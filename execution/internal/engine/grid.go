package engine

import (
	"time"

	"github.com/rs/zerolog"

	"trading-bot/execution/internal/config"
	"trading-bot/execution/internal/types"
)

// Grid re-entry tunables. After a grid position fully closes, the grid re-arms and can
// re-buy once reEntryCooldown has elapsed — provided the short-term trend isn't falling.
// These are intentionally gentler/shorter than the engine crash guard (-2% / 30 min),
// which remains the global hard block on any buy.
const (
	reEntryCooldown    = 10 * time.Minute
	trendLookback      = 10 * time.Minute
	trendDropThreshold = 0.003 // skip re-buy if price fell >0.3% over the lookback window
)

type GridStrategy struct {
	cfg    config.StrategyConfig
	logger zerolog.Logger

	lastBuy   map[string]time.Time
	priceHist map[string][]pricePoint
}

func NewGridStrategy(cfg config.StrategyConfig, logger zerolog.Logger) *GridStrategy {
	return &GridStrategy{
		cfg:       cfg,
		logger:    logger.With().Str("strategy", "grid").Logger(),
		lastBuy:   make(map[string]time.Time),
		priceHist: make(map[string][]pricePoint),
	}
}

func (s *GridStrategy) Name() string { return "grid" }

func (s *GridStrategy) Evaluate(state *types.MarketState) *types.Decision {
	symbol := state.Symbol

	interval := time.Duration(s.cfg.RSIPeriod) * time.Hour
	if interval == 0 {
		interval = 2 * time.Hour
	}
	buyUSD := s.cfg.PositionSizeUSD
	if buyUSD == 0 {
		buyUSD = 50
	}
	maxPositions := s.cfg.OrderBookDepth
	if maxPositions == 0 {
		maxPositions = 2
	}

	entryCount := 0
	for _, p := range state.Positions {
		if p.Symbol == symbol && abs(p.Amount) > 0.00001 {
			entryCount++
		}
	}

	if entryCount >= maxPositions {
		s.logger.Debug().Str("symbol", symbol).Int("count", entryCount).Int("max", maxPositions).Msg("grid: position cap")
		return nil
	}

	// Position-aware timing gate, both measured from the last buy:
	//   - flat: re-arm after a short cooldown, so it re-buys promptly once a held position
	//     has closed (its last buy is old) while still throttling rapid re-fires before a
	//     fill registers.
	//   - holding >=1 leg: keep the full DCA interval before adding another leg.
	wait := interval
	if entryCount == 0 {
		wait = reEntryCooldown
	}
	if since := time.Since(s.lastBuy[symbol]); since < wait {
		s.logger.Debug().Str("symbol", symbol).Dur("since_last", since).Dur("wait", wait).Int("held", entryCount).Msg("grid: not time yet")
		return nil
	}

	ticker := state.Ticker
	if ticker == nil || ticker.Last <= 0 {
		s.logger.Debug().Str("symbol", symbol).Msg("grid: no ticker")
		return nil
	}

	// Record the price sample and hold off buying if the short-term trend is falling.
	s.recordPrice(symbol, ticker.Last)
	if drop, ok := s.shortDowntrend(symbol, ticker.Last); ok {
		s.logger.Info().Str("symbol", symbol).Float64("drop_pct", drop*100).Msg("grid: holding off, short downtrend")
		return nil
	}

	s.lastBuy[symbol] = time.Now()

	amount := buyUSD / ticker.Last

	s.logger.Info().Str("symbol", symbol).Float64("amount", amount).Float64("price", ticker.Last).Int("concurrent", entryCount+1).Msg("grid buy")

	return &types.Decision{
		Action:   types.ActionBuy,
		Symbol:   symbol,
		Side:     types.SideBuy,
		Amount:   amount,
		Price:    ticker.Last,
		Type:     types.TypeLimit,
		Reason:   "grid: interval buy",
		Strategy: s.Name(),
	}
}

// recordPrice appends a timestamped price and prunes samples older than the trend window.
func (s *GridStrategy) recordPrice(symbol string, price float64) {
	cutoff := time.Now().Add(-trendLookback - time.Minute)
	pts := append(s.priceHist[symbol], pricePoint{t: time.Now(), price: price})
	i := 0
	for i < len(pts) && pts[i].t.Before(cutoff) {
		i++
	}
	s.priceHist[symbol] = pts[i:]
}

// shortDowntrend reports whether price has fallen more than trendDropThreshold versus the
// oldest retained sample (~one lookback ago). Returns false until enough history exists so
// it never blocks right after startup or a restart.
func (s *GridStrategy) shortDowntrend(symbol string, price float64) (float64, bool) {
	pts := s.priceHist[symbol]
	if len(pts) < 2 {
		return 0, false
	}
	oldest := pts[0]
	if oldest.price <= 0 || time.Since(oldest.t) < trendLookback/2 {
		return 0, false
	}
	change := (price - oldest.price) / oldest.price
	return change, change <= -trendDropThreshold
}
