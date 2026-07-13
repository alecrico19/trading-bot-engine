package engine

import (
	"time"

	"github.com/rs/zerolog"

	"trading-bot/execution/internal/config"
	"trading-bot/execution/internal/types"
)

type ScalpingStrategy struct {
	cfg          config.StrategyConfig
	logger       zerolog.Logger
	prices       map[string]*PriceHistory
	lastSide     map[string]string
	lastDecision map[string]time.Time
}

func NewScalpingStrategy(cfg config.StrategyConfig, logger zerolog.Logger) *ScalpingStrategy {
	return &ScalpingStrategy{
		cfg:          cfg,
		logger:       logger.With().Str("strategy", "scalping").Logger(),
		prices:       make(map[string]*PriceHistory),
		lastSide:     make(map[string]string),
		lastDecision: make(map[string]time.Time),
	}
}

func (s *ScalpingStrategy) Name() string { return "scalping" }

func (s *ScalpingStrategy) FeedTrade(trade *types.Trade) {
	hist, ok := s.prices[trade.Symbol]
	if !ok {
		hist = NewPriceHistory(50)
		s.prices[trade.Symbol] = hist
	}
	hist.Add(trade.Price)
}

func (s *ScalpingStrategy) Evaluate(state *types.MarketState) *types.Decision {
	hist, ok := s.prices[state.Symbol]
	if !ok || hist.Len() < 10 {
		return nil
	}

	if time.Since(s.lastDecision[state.Symbol]) < 2*time.Second {
		return nil
	}

	window := s.cfg.OrderBookDepth
	if window == 0 {
		window = 5
	}

	vals := hist.Values()
	if len(vals) < window {
		return nil
	}
	recent := vals[len(vals)-window:]

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
	entryCount := 0
	var holding float64
	for _, p := range state.Positions {
		if p.Symbol == state.Symbol && abs(p.Amount) > 0.00001 {
			hasPosition = true
			entryCount++
			holding += abs(p.Amount)
		}
	}

	ticker := state.Ticker
	if ticker == nil || ticker.Last <= 0 {
		return nil
	}

	upPct := float64(upCount) / float64(total)
	lastSide, _ := s.lastSide[state.Symbol]

	signal := s.getSignal(state)
	signalID := ""
	if signal != nil && signal.IsActionable() {
		signalID = signal.ID
	}

	// Strong buying momentum + no position → enter long
	if upPct >= 0.80 && !hasPosition && entryCount < 10 {
		s.lastSide[state.Symbol] = "buy"
		s.lastDecision[state.Symbol] = time.Now()
		return &types.Decision{
			Action:   types.ActionBuy,
			Symbol:   state.Symbol,
			Side:     types.SideBuy,
			Amount:   0,
			Price:    ticker.Last,
			Type:     types.TypeLimit,
			Reason:   "scalp: strong uptick momentum",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	// Buy signal (momentum is up) + no position
	if upPct >= 0.60 && !hasPosition && entryCount < 10 {
		s.lastSide[state.Symbol] = "buy"
		s.lastDecision[state.Symbol] = time.Now()
		return &types.Decision{
			Action:   types.ActionBuy,
			Symbol:   state.Symbol,
			Side:     types.SideBuy,
			Amount:   0,
			Price:    ticker.Last,
			Type:     types.TypeLimit,
			Reason:   "scalp: uptick momentum",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	// Strong selling momentum + position held → exit
	if upPct <= 0.20 && hasPosition {
		s.lastSide[state.Symbol] = "sell"
		s.lastDecision[state.Symbol] = time.Now()
		return &types.Decision{
			Action:   types.ActionSell,
			Symbol:   state.Symbol,
			Side:     types.SideSell,
			Amount:   holding,
			Price:    ticker.Last,
			Type:     types.TypeLimit,
			Reason:   "scalp: strong downtick momentum (exit)",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	// Selling momentum + position held → exit
	if upPct <= 0.40 && hasPosition {
		s.lastSide[state.Symbol] = "sell"
		s.lastDecision[state.Symbol] = time.Now()
		return &types.Decision{
			Action:   types.ActionSell,
			Symbol:   state.Symbol,
			Side:     types.SideSell,
			Amount:   holding,
			Price:    ticker.Last,
			Type:     types.TypeLimit,
			Reason:   "scalp: downtick momentum (exit)",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	// Reversal from previous buy
	if lastSide == "buy" && upPct <= 0.5 && hasPosition {
		s.lastSide[state.Symbol] = "sell"
		s.lastDecision[state.Symbol] = time.Now()
		return &types.Decision{
			Action:   types.ActionSell,
			Symbol:   state.Symbol,
			Side:     types.SideSell,
			Amount:   holding,
			Price:    ticker.Last,
			Type:     types.TypeLimit,
			Reason:   "scalp: momentum reverse (exit)",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	return nil
}

func (s *ScalpingStrategy) getSignal(state *types.MarketState) *types.Signal {
	for _, sig := range state.Signals {
		if sig.Symbol == state.Symbol && sig.IsActionable() {
			return &sig
		}
	}
	return nil
}
