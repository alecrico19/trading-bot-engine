package engine

import (
	"github.com/rs/zerolog"

	"trading-bot/execution/internal/config"
	"trading-bot/execution/internal/types"
)

type MeanReversionStrategy struct {
	cfg    config.StrategyConfig
	logger zerolog.Logger
	prices map[string]*PriceHistory
}

func NewMeanReversionStrategy(cfg config.StrategyConfig, logger zerolog.Logger) *MeanReversionStrategy {
	rsiPeriod := cfg.RSIPeriod
	if rsiPeriod == 0 {
		rsiPeriod = 14
	}
	bbPeriod := cfg.BBPeriod
	if bbPeriod == 0 {
		bbPeriod = 20
	}
	maxLen := rsiPeriod
	if bbPeriod > maxLen {
		maxLen = bbPeriod
	}
	maxLen += 10

	return &MeanReversionStrategy{
		cfg:    cfg,
		logger: logger.With().Str("strategy", "mean-reversion").Logger(),
		prices: make(map[string]*PriceHistory),
	}
}

func (s *MeanReversionStrategy) Name() string { return "mean-reversion" }

func (s *MeanReversionStrategy) Evaluate(state *types.MarketState) *types.Decision {
	ticker := state.Ticker
	if ticker == nil || ticker.Last <= 0 {
		return nil
	}

	hist, ok := s.prices[state.Symbol]
	if !ok {
		hist = NewPriceHistory(100)
		s.prices[state.Symbol] = hist
	}
	hist.Add(ticker.Last)

	rsiPeriod := s.cfg.RSIPeriod
	if rsiPeriod == 0 {
		rsiPeriod = 14
	}
	bbPeriod := s.cfg.BBPeriod
	if bbPeriod == 0 {
		bbPeriod = 20
	}
	bbStdDev := s.cfg.BBStdDev
	if bbStdDev == 0 {
		bbStdDev = 2.0
	}
	rsiOversold := s.cfg.RSIOversold
	if rsiOversold == 0 {
		rsiOversold = 30
	}
	rsiOverbought := s.cfg.RSIOverbought
	if rsiOverbought == 0 {
		rsiOverbought = 70
	}

	if hist.Len() < rsiPeriod+1 || hist.Len() < bbPeriod {
		return nil
	}

	rsi := hist.RSI(rsiPeriod)
	_, upper, lower := hist.BollingerBands(bbPeriod, bbStdDev)

	if lower <= 0 || upper <= 0 {
		return nil
	}

	s.logger.Debug().
		Float64("price", ticker.Last).
		Float64("rsi", rsi).
		Float64("bb_lower", lower).
		Float64("bb_upper", upper).
		Str("symbol", state.Symbol).
		Msg("mean reversion check")

	signal := s.getSignal(state)
	signalID := ""
	if signal != nil && signal.IsActionable() {
		signalID = signal.ID
	}

	for _, p := range state.Positions {
		if p.Symbol == state.Symbol && abs(p.Amount) > 0.00001 {
			if p.Side == "long" && (rsi > rsiOverbought || ticker.Last > upper) {
				return &types.Decision{
					Action:   types.ActionSell,
					Symbol:   state.Symbol,
					Side:     types.SideSell,
					Amount:   p.Amount,
					Price:    ticker.Last,
					Type:     types.TypeLimit,
					Reason:   "take profit: RSI overbought or above upper BB",
					SignalID: signalID,
					Strategy: s.Name(),
				}
			}
			if p.Side == "short" && (rsi < rsiOversold || ticker.Last < lower) {
				return &types.Decision{
					Action:   types.ActionBuy,
					Symbol:   state.Symbol,
					Side:     types.SideBuy,
					Amount:   abs(p.Amount),
					Price:    ticker.Last,
					Type:     types.TypeLimit,
					Reason:   "cover short: RSI oversold or below lower BB",
					SignalID: signalID,
					Strategy: s.Name(),
				}
			}
			return nil
		}
	}

	if rsi < rsiOversold && ticker.Last <= lower {
		if signal != nil && signal.Direction == types.SignalDirectionShort {
			return nil
		}
		return &types.Decision{
			Action:   types.ActionBuy,
			Symbol:   state.Symbol,
			Side:     types.SideBuy,
			Amount:   0,
			Price:    ticker.Last,
			Type:     types.TypeLimit,
			Reason:   "entry: RSI oversold + below lower BB",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	if rsi > rsiOverbought && ticker.Last >= upper {
		if signal != nil && signal.Direction == types.SignalDirectionLong {
			return nil
		}
		return &types.Decision{
			Action:   types.ActionSell,
			Symbol:   state.Symbol,
			Side:     types.SideSell,
			Amount:   0,
			Price:    ticker.Last,
			Type:     types.TypeLimit,
			Reason:   "entry: RSI overbought + above upper BB",
			SignalID: signalID,
			Strategy: s.Name(),
		}
	}

	return nil
}

func (s *MeanReversionStrategy) getSignal(state *types.MarketState) *types.Signal {
	for _, sig := range state.Signals {
		if sig.Symbol == state.Symbol && sig.IsActionable() {
			return &sig
		}
	}
	return nil
}
