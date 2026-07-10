package engine

import (
	"github.com/rs/zerolog"

	"trading-bot/execution/internal/config"
	"trading-bot/execution/internal/types"
)

type ScalpingStrategy struct {
	cfg    config.StrategyConfig
	logger zerolog.Logger
	prices *PriceHistory
}

func NewScalpingStrategy(cfg config.StrategyConfig, logger zerolog.Logger) *ScalpingStrategy {
	emaPeriod := cfg.RSIPeriod
	if emaPeriod == 0 {
		emaPeriod = 200
	}
	return &ScalpingStrategy{
		cfg:    cfg,
		logger: logger.With().Str("strategy", "scalping").Logger(),
		prices: NewPriceHistory(emaPeriod * 2),
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

	// Update trend tracker (EMA on mid-price for informational use)
	midPrice := (bestBid + bestAsk) / 2
	s.prices.Add(midPrice)

	hasPosition := false
	entryCount := 0
	for _, p := range state.Positions {
		if p.Symbol == state.Symbol && abs(p.Amount) > 0.00001 {
			hasPosition = true
			entryCount++
		}
	}

	signal := s.getSignal(state)
	signalID := ""
	confidence := 0.0
	if signal != nil && signal.IsActionable() {
		signalID = signal.ID
		confidence = signal.Confidence
	}

	if ratio > 1.5 {
		if !hasPosition {
			return nil
		}
		return &types.Decision{
			Action:   types.ActionSell,
			Symbol:   state.Symbol,
			Side:     types.SideSell,
			Amount:   0,
			Price:    bestAsk,
			Type:     types.TypeLimit,
			Reason:   "sell into strength: heavy bid volume",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	if ratio > 1.05 {
		if !hasPosition {
			return nil
		}
		if signal != nil && signal.Direction == types.SignalDirectionShort {
			s.logger.Debug().Float64("ratio", ratio).Msg("sell signal overridden by research short bias")
			return nil
		}
		return &types.Decision{
			Action:   types.ActionSell,
			Symbol:   state.Symbol,
			Side:     types.SideSell,
			Amount:   0,
			Price:    bestAsk,
			Type:     types.TypeLimit,
			Reason:   "sell into strength: heavy bid volume",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	if ratio < 0.8 {
		if entryCount >= 10 {
			return nil
		}
		return &types.Decision{
			Action:   types.ActionBuy,
			Symbol:   state.Symbol,
			Side:     types.SideBuy,
			Amount:   0,
			Price:    bestBid,
			Type:     types.TypeLimit,
			Reason:   "buy into weakness: heavy ask volume",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	if ratio < 0.95 {
		if entryCount >= 10 {
			return nil
		}
		if signal != nil && signal.Direction == types.SignalDirectionLong {
			s.logger.Debug().Float64("ratio", ratio).Msg("buy signal overridden by research long bias")
			return nil
		}
		return &types.Decision{
			Action:   types.ActionBuy,
			Symbol:   state.Symbol,
			Side:     types.SideBuy,
			Amount:   0,
			Price:    bestBid,
			Type:     types.TypeLimit,
			Reason:   "buy into weakness: heavy ask volume",
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

func (s *ScalpingStrategy) computeEMA(period int) float64 {
	vals := s.prices.Values()
	if len(vals) < period {
		return 0
	}
	multiplier := 2.0 / float64(period+1)
	ema := vals[0]
	for _, v := range vals[1:] {
		ema = (v-ema)*multiplier + ema
	}
	return ema
}
