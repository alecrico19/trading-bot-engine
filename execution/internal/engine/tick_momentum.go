package engine

import (
	"github.com/rs/zerolog"

	"trading-bot/execution/internal/config"
	"trading-bot/execution/internal/types"
)

type TickMomentumStrategy struct {
	cfg       config.StrategyConfig
	logger    zerolog.Logger
	prices    map[string]*PriceHistory
	lastSide  map[string]string
}

func NewTickMomentumStrategy(cfg config.StrategyConfig, logger zerolog.Logger) *TickMomentumStrategy {
	return &TickMomentumStrategy{
		cfg:      cfg,
		logger:   logger.With().Str("strategy", "tick-momentum").Logger(),
		prices:   make(map[string]*PriceHistory),
		lastSide: make(map[string]string),
	}
}

func (s *TickMomentumStrategy) Name() string { return "tick-momentum" }

func (s *TickMomentumStrategy) FeedTrade(trade *types.Trade) {
	hist, ok := s.prices[trade.Symbol]
	if !ok {
		hist = NewPriceHistory(20)
		s.prices[trade.Symbol] = hist
	}
	hist.Add(trade.Price)

	minTicks := s.cfg.OrderBookDepth
	if minTicks == 0 {
		minTicks = 5
	}
	if hist.Len() < minTicks {
		return
	}

	vals := hist.Values()
	recent := vals[len(vals)-minTicks:]

	upCount := 0
	downCount := 0
	for i := 1; i < len(recent); i++ {
		if recent[i] > recent[i-1] {
			upCount++
		}
		if recent[i] < recent[i-1] {
			downCount++
		}
	}

	s.logger.Debug().
		Str("symbol", trade.Symbol).
		Int("up", upCount).
		Int("down", downCount).
		Float64("price", trade.Price).
		Msg("tick momentum check")
}

func (s *TickMomentumStrategy) Evaluate(state *types.MarketState) *types.Decision {
	hist, ok := s.prices[state.Symbol]
	if !ok {
		return nil
	}

	minTicks := s.cfg.OrderBookDepth
	if minTicks == 0 {
		minTicks = 5
	}
	if hist.Len() < minTicks {
		return nil
	}

	vals := hist.Values()
	recent := vals[len(vals)-minTicks:]

	upCount := 0
	downCount := 0
	for i := 1; i < len(recent); i++ {
		if recent[i] > recent[i-1] {
			upCount++
		}
		if recent[i] < recent[i-1] {
			downCount++
		}
	}
	total := upCount + downCount
	if total == 0 {
		return nil
	}

	hasPosition := false
	for _, p := range state.Positions {
		if p.Symbol == state.Symbol && abs(p.Amount) > 0.00001 {
			hasPosition = true
			break
		}
	}

	lastSide, _ := s.lastSide[state.Symbol]
	ticker := state.Ticker
	if ticker == nil || ticker.Last <= 0 {
		return nil
	}

	signal := s.getSignal(state)
	signalID := ""
	if signal != nil && signal.IsActionable() {
		signalID = signal.ID
	}

	upPct := float64(upCount) / float64(total)

	if upPct >= 0.60 && !hasPosition {
		s.lastSide[state.Symbol] = "buy"
		return &types.Decision{
			Action:   types.ActionBuy,
			Symbol:   state.Symbol,
			Side:     types.SideBuy,
			Amount:   0,
			Price:    ticker.Last,
			Type:     types.TypeLimit,
			Reason:   "tick momentum: 75%+ consecutive upticks",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	if upPct <= 0.40 && hasPosition {
		s.lastSide[state.Symbol] = "sell"
		return &types.Decision{
			Action:   types.ActionSell,
			Symbol:   state.Symbol,
			Side:     types.SideSell,
			Amount:   0,
			Price:    ticker.Last,
			Type:     types.TypeLimit,
			Reason:   "tick momentum: 75%+ consecutive downticks (exit)",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	if lastSide == "buy" && upPct <= 0.5 && hasPosition {
		return &types.Decision{
			Action:   types.ActionSell,
			Symbol:   state.Symbol,
			Side:     types.SideSell,
			Amount:   0,
			Price:    ticker.Last,
			Type:     types.TypeLimit,
			Reason:   "tick momentum: reverse signal (exit)",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	return nil
}

func (s *TickMomentumStrategy) getSignal(state *types.MarketState) *types.Signal {
	for _, sig := range state.Signals {
		if sig.Symbol == state.Symbol && sig.IsActionable() {
			return &sig
		}
	}
	return nil
}
