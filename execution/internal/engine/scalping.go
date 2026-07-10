package engine

import (
	"github.com/rs/zerolog"

	"trading-bot/execution/internal/config"
	"trading-bot/execution/internal/types"
)

type ScalpingStrategy struct {
	cfg    config.StrategyConfig
	logger zerolog.Logger
}

func NewScalpingStrategy(cfg config.StrategyConfig, logger zerolog.Logger) *ScalpingStrategy {
	return &ScalpingStrategy{
		cfg:    cfg,
		logger: logger.With().Str("strategy", "scalping").Logger(),
	}
}

func (s *ScalpingStrategy) Name() string { return "scalping" }

func (s *ScalpingStrategy) Evaluate(state *types.MarketState) *types.Decision {
	if state.OrderBook == nil || len(state.OrderBook.Bids) == 0 || len(state.OrderBook.Asks) == 0 {
		return nil
	}

	depth := s.cfg.OrderBookDepth
	if depth == 0 {
		depth = 10
	}
	if len(state.OrderBook.Bids) < depth || len(state.OrderBook.Asks) < depth {
		return nil
	}

	bidVolume := 0.0
	for i := 0; i < depth; i++ {
		bidVolume += state.OrderBook.Bids[i].Quantity
	}
	askVolume := 0.0
	for i := 0; i < depth; i++ {
		askVolume += state.OrderBook.Asks[i].Quantity
	}

	if bidVolume == 0 || askVolume == 0 {
		return nil
	}

	bestBid := state.OrderBook.Bids[0].Price
	bestAsk := state.OrderBook.Asks[0].Price
	if bestBid <= 0 || bestAsk <= 0 {
		return nil
	}

	spread := (bestAsk - bestBid) / bestBid * 100
	maxSpread := s.cfg.MinSpreadPct
	if maxSpread <= 0 {
		maxSpread = 0.10
	}
	if spread > maxSpread*100 {
		return nil
	}

	ratio := bidVolume / askVolume

	signal := s.getSignal(state)
	signalID := ""
	confidence := 0.0
	if signal != nil && signal.IsActionable() {
		signalID = signal.ID
		confidence = signal.Confidence
	}

	if ratio > 2.0 {
		return &types.Decision{
			Action:   types.ActionBuy,
			Symbol:   state.Symbol,
			Side:     types.SideBuy,
			Amount:   0,
			Price:    bestAsk,
			Type:     types.TypeLimit,
			Reason:   "strong order book imbalance: heavy bid volume",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	if ratio > 1.2 {
		if signal != nil && signal.Direction == types.SignalDirectionShort {
			s.logger.Debug().Float64("ratio", ratio).Msg("scalp long signal overridden by research short bias")
			return nil
		}
		return &types.Decision{
			Action:   types.ActionBuy,
			Symbol:   state.Symbol,
			Side:     types.SideBuy,
			Amount:   0,
			Price:    bestAsk,
			Type:     types.TypeLimit,
			Reason:   "order book imbalance: heavy bid volume",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	if ratio < 0.5 {
		return &types.Decision{
			Action:   types.ActionSell,
			Symbol:   state.Symbol,
			Side:     types.SideSell,
			Amount:   0,
			Price:    bestBid,
			Type:     types.TypeLimit,
			Reason:   "strong order book imbalance: heavy ask volume",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	if ratio < 0.83 {
		if signal != nil && signal.Direction == types.SignalDirectionLong {
			s.logger.Debug().Float64("ratio", ratio).Msg("scalp short signal overridden by research long bias")
			return nil
		}
		return &types.Decision{
			Action:   types.ActionSell,
			Symbol:   state.Symbol,
			Side:     types.SideSell,
			Amount:   0,
			Price:    bestBid,
			Type:     types.TypeLimit,
			Reason:   "order book imbalance: heavy ask volume",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	if confidence > 0.7 && signal != nil && signal.Type == types.SignalTypeAlert {
		return &types.Decision{
			Action: types.ActionReduce,
			Symbol: state.Symbol,
			Reason: "alert signal: risk reduction",
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
