package engine

import (
	"time"

	"github.com/rs/zerolog"

	"trading-bot/execution/internal/config"
	"trading-bot/execution/internal/types"
)

type GridStrategy struct {
	cfg       config.StrategyConfig
	logger    zerolog.Logger
	lastBuy  map[string]time.Time
}

func NewGridStrategy(cfg config.StrategyConfig, logger zerolog.Logger) *GridStrategy {
	return &GridStrategy{
		cfg:     cfg,
		logger:  logger.With().Str("strategy", "grid").Logger(),
		lastBuy: make(map[string]time.Time),
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
		buyUSD = 25
	}
	maxPositions := s.cfg.OrderBookDepth
	if maxPositions == 0 {
		maxPositions = 3
	}

	since := time.Since(s.lastBuy[symbol])
	if since < interval {
		s.logger.Debug().Str("symbol", symbol).Dur("since_last", since).Dur("interval", interval).Msg("grid: not time yet")
		return nil
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

	ticker := state.Ticker
	if ticker == nil || ticker.Last <= 0 {
		s.logger.Debug().Str("symbol", symbol).Msg("grid: no ticker")
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
