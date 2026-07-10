package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"trading-bot/execution/internal/alert"
	"trading-bot/execution/internal/config"
	"trading-bot/execution/internal/db"
	"trading-bot/execution/internal/exchange"
	"trading-bot/execution/internal/order"
	"trading-bot/execution/internal/risk"
	"trading-bot/execution/internal/signal"
	"trading-bot/execution/internal/types"
)

type EquityPoint struct {
	Time  time.Time `json:"t"`
	Value float64   `json:"v"`
}

type Engine struct {
	cfg        *config.Config
	exchange   exchange.Exchange
	marketData exchange.Exchange
	orderMgr   *order.Manager
	riskMgr    *risk.Manager
	signalCon  *signal.Consumer
	db         *db.Store
	alerts     *alert.Service
	logger     zerolog.Logger

	signals     []types.Signal
	strategies  []Strategy
	stratPaused map[string]bool
	stratPnL    map[string]float64
	tradeCount  int
	highWater   map[string]float64
	entryPrice  map[string]float64
	entryTime   map[string]time.Time
	equityHist  []EquityPoint

	mu     sync.RWMutex
	running bool
}

func New(cfg *config.Config, ex exchange.Exchange, marketData exchange.Exchange, om *order.Manager, rm *risk.Manager, sc *signal.Consumer, store *db.Store, alerts *alert.Service, logger zerolog.Logger) *Engine {
	if marketData == nil {
		marketData = ex
	}
	return &Engine{
		cfg:         cfg,
		exchange:    ex,
		marketData:  marketData,
		orderMgr:    om,
		riskMgr:     rm,
		signalCon:   sc,
		db:          store,
		alerts:      alerts,
		logger:      logger.With().Str("component", "engine").Logger(),
		stratPaused: make(map[string]bool),
		stratPnL:    make(map[string]float64),
		highWater:   make(map[string]float64),
		entryPrice:  make(map[string]float64),
		entryTime:   make(map[string]time.Time),
	}
}

func (e *Engine) RegisterStrategy(s Strategy) {
	e.strategies = append(e.strategies, s)
	e.stratPaused[s.Name()] = false
	e.logger.Info().Str("strategy", s.Name()).Msg("strategy registered")
}

func (e *Engine) Run(ctx context.Context) error {
	e.mu.Lock()
	e.running = true
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		e.running = false
		e.mu.Unlock()
	}()

	balances, err := e.exchange.FetchBalance(ctx)
	if err != nil {
		return fmt.Errorf("fetch initial balance: %w", err)
	}
	equity := calculateEquity(balances)
	e.riskMgr.SetStartEquity(equity)

	e.reconcilePositions(ctx)

	e.logger.Info().Float64("equity", equity).Int("strategies", len(e.strategies)).Msg("engine starting")
	e.alerts.Send(alert.LevelInfo, "Engine Started",
		fmt.Sprintf("Equity: $%.2f | Strategies: %d | Symbols: %v", equity, len(e.strategies), e.collectSymbols()))

	signalCh, err := e.signalCon.Subscribe(ctx)
	if err != nil {
		e.logger.Warn().Err(err).Msg("signal subscription failed, continuing without signals")
	}

	symbols := e.collectSymbols()
	if len(symbols) == 0 {
		return fmt.Errorf("no symbols configured for strategies")
	}

	var wg sync.WaitGroup
	decisionCh := make(chan *types.Decision, 10)
	tickerCh := make(chan struct{}, 1)
	tickerCh <- struct{}{}

	for _, symbol := range symbols {
		wg.Add(1)
		go e.runOrderBookLoop(ctx, &wg, symbol, decisionCh, tickerCh)
	}

	wg.Add(1)
	go e.runTradeLoop(ctx, &wg, symbols, tickerCh)

	wg.Add(1)
	go e.runSignalLoop(ctx, &wg, signalCh, tickerCh)

	wg.Add(1)
	go e.runDecisionLoop(ctx, &wg, decisionCh)

	wg.Add(1)
	go e.runStopLoss(ctx, &wg, symbols)

	wg.Add(1)
	go e.runEquityRecorder(ctx, &wg)

	wg.Add(1)
	go e.runPnLSnapshot(ctx, &wg)

	wg.Add(1)
	go e.runAlertMonitor(ctx, &wg)

	<-ctx.Done()
	e.logger.Info().Msg("engine shutting down")
	e.saveDailySnapshot()
	e.alerts.Send(alert.LevelInfo, "Engine Stopped", "Shutdown complete")
	wg.Wait()

	return nil
}

func (e *Engine) collectSymbols() []string {
	seen := make(map[string]bool)
	for _, s := range e.strategies {
		switch s.Name() {
		case "scalping":
			for _, sym := range e.cfg.Strategies["scalping"].Symbols {
				seen[sym] = true
			}
		case "mean-reversion":
			for _, sym := range e.cfg.Strategies["mean-reversion"].Symbols {
				seen[sym] = true
			}
		}
	}
	symbols := make([]string, 0, len(seen))
	for s := range seen {
		symbols = append(symbols, s)
	}
	return symbols
}

func (e *Engine) runOrderBookLoop(ctx context.Context, wg *sync.WaitGroup, symbol string, decisionCh chan<- *types.Decision, tickerCh chan<- struct{}) {
	defer wg.Done()

	obCh, err := e.marketData.SubscribeOrderBook(ctx, symbol)
	if err != nil {
		e.logger.Error().Err(err).Str("symbol", symbol).Msg("failed to subscribe order book, using polling")
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case ob, ok := <-obCh:
			if !ok {
				return
			}
			e.forwardOrderBook(ob)

			state := e.buildMarketState(symbol)
			state.OrderBook = ob

			for _, strat := range e.strategies {
				if !e.isStrategyForSymbol(strat.Name(), symbol) {
					continue
				}
				e.mu.RLock()
				paused := e.stratPaused[strat.Name()]
				e.mu.RUnlock()
				if paused {
					continue
				}
				decision := strat.Evaluate(state)
				if decision != nil {
					select {
					case decisionCh <- decision:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}
}

func (e *Engine) runTradeLoop(ctx context.Context, wg *sync.WaitGroup, symbols []string, tickerCh chan<- struct{}) {
	defer wg.Done()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	meanRevTicker := time.NewTicker(15 * time.Second)
	defer meanRevTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-meanRevTicker.C:
			for _, symbol := range symbols {
				e.evaluateMeanReversion(ctx, symbol)
			}
		case <-ticker.C:
			select {
			case tickerCh <- struct{}{}:
			default:
			}
		}
	}
}

func (e *Engine) evaluateMeanReversion(ctx context.Context, symbol string) {
	ticker, err := e.marketData.FetchTicker(ctx, symbol)
	if err != nil {
		e.logger.Error().Err(err).Str("symbol", symbol).Msg("fetch ticker failed")
		return
	}

	state := e.buildMarketState(symbol)
	state.Ticker = ticker

	for _, strat := range e.strategies {
		if strat.Name() != "mean-reversion" || !e.isStrategyForSymbol("mean-reversion", symbol) {
			continue
		}
		e.mu.RLock()
		paused := e.stratPaused[strat.Name()]
		e.mu.RUnlock()
		if paused {
			continue
		}
		decision := strat.Evaluate(state)
		if decision != nil {
			e.executeDecision(ctx, decision)
		}
	}
}

func (e *Engine) runSignalLoop(ctx context.Context, wg *sync.WaitGroup, signalCh <-chan *types.Signal, tickerCh chan<- struct{}) {
	defer wg.Done()

	if signalCh == nil {
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case sig, ok := <-signalCh:
			if !ok {
				return
			}
			e.mu.Lock()
			e.signals = append(e.signals, *sig)
			if len(e.signals) > 50 {
				valid := e.signals[:0]
				for _, s := range e.signals {
					if !s.IsExpired() {
						valid = append(valid, s)
					}
				}
				e.signals = valid
			}
			e.mu.Unlock()

			e.db.RecordSignal(sig, "logged")
		}
	}
}

func (e *Engine) runDecisionLoop(ctx context.Context, wg *sync.WaitGroup, decisionCh <-chan *types.Decision) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case decision, ok := <-decisionCh:
			if !ok {
				return
			}
			e.executeDecision(ctx, decision)
		}
	}
}

func (e *Engine) executeDecision(ctx context.Context, decision *types.Decision) {
	if e.riskMgr.IsBreached() {
		e.logger.Debug().Str("strategy", decision.Strategy).Msg("decision blocked: circuit breaker")
		return
	}

	equity := e.GetEquity()

	switch decision.Action {
	case types.ActionBuy, types.ActionSell:
		if err := e.riskMgr.CanOpenPosition(decision.Symbol, equity, 0); err != nil {
			e.logger.Debug().Err(err).Str("strategy", decision.Strategy).Msg("position blocked by risk")
			return
		}

		amount := decision.Amount
		if amount == 0 {
			positionSize := equity * e.cfg.Risk.MaxPositionPct
			if decision.Price > 0 {
				amount = positionSize / decision.Price
			}
		}
		if amount <= 0 {
			return
		}

		ticker, _ := e.marketData.FetchTicker(ctx, decision.Symbol)
		if err := e.riskMgr.ValidateOrder(decision.Symbol, decision.Side, decision.Type, amount, decision.Price, ticker); err != nil {
			e.logger.Warn().Err(err).Str("strategy", decision.Strategy).Msg("order validation failed")
			return
		}

		order, err := e.orderMgr.PlaceOrder(ctx, decision.Symbol, decision.Side, decision.Type, amount, decision.Price, decision.Strategy)
		if err != nil {
			e.logger.Error().Err(err).Str("strategy", decision.Strategy).Msg("order failed")
			return
		}
		if decision.SignalID != "" {
			order.SignalID = decision.SignalID
		}

		if order.Status != types.StatusFilled && order.Status != types.StatusPartiallyFilled {
			return
		}

		pnl := 0.0
		if order.Side == types.SideBuy {
			pnl = 0
		} else {
			e.mu.RLock()
			entry := e.entryPrice[decision.Symbol]
			e.mu.RUnlock()
			if entry > 0 {
				pnl = (order.AvgPrice - entry) * order.Filled
			}
		}
		e.logger.Info().
			Str("symbol", decision.Symbol).
			Str("side", string(order.Side)).
			Float64("entry", func() float64 { e.mu.RLock(); defer e.mu.RUnlock(); return e.entryPrice[decision.Symbol] }()).
			Float64("avg_price", order.AvgPrice).
			Float64("filled", order.Filled).
			Float64("pnl", pnl).
			Msg("trade P&L")
		e.db.RecordTrade(order, pnl)
		e.riskMgr.RecordTrade(pnl)

		e.mu.Lock()
		e.tradeCount++
		e.stratPnL[decision.Strategy] += pnl
		if order.Side == types.SideBuy {
			prev := e.entryPrice[decision.Symbol]
			if prev > 0 {
				e.entryPrice[decision.Symbol] = (prev + order.AvgPrice) / 2
			} else {
				e.entryPrice[decision.Symbol] = order.AvgPrice
				e.entryTime[decision.Symbol] = time.Now()
			}
			e.highWater[decision.Symbol] = order.AvgPrice
		} else {
			pos, _ := e.orderMgr.GetPosition(decision.Symbol)
			if pos == nil || abs(pos.Amount) < 0.00001 {
				e.entryPrice[decision.Symbol] = 0
				e.highWater[decision.Symbol] = 0
				e.entryTime[decision.Symbol] = time.Time{}
			}
		}
		e.mu.Unlock()

		e.logger.Info().
			Str("strategy", decision.Strategy).
			Str("reason", decision.Reason).
			Str("order_id", order.ID).
			Float64("pnl", pnl).
			Msg("executed decision")

	case types.ActionClose:
		e.orderMgr.CancelAllOpen(ctx, decision.Symbol)

	case types.ActionReduce:
		e.orderMgr.CancelAllOpen(ctx, decision.Symbol)
		e.alerts.Send(alert.LevelWarn, "Position Reduced",
			fmt.Sprintf("%s: %s — cancelling all open orders", decision.Symbol, decision.Reason))
		e.logger.Warn().Str("reason", decision.Reason).Msg("positions reduced")

	case types.ActionHold:
	}
}

func (e *Engine) buildMarketState(symbol string) *types.MarketState {
	e.mu.RLock()
	signals := make([]types.Signal, len(e.signals))
	copy(signals, e.signals)
	e.mu.RUnlock()

	positions := e.orderMgr.GetPositions()

	return &types.MarketState{
		Symbol:    symbol,
		Signals:   signals,
		Positions: positions,
	}
}

func (e *Engine) isStrategyForSymbol(strategyName, symbol string) bool {
	cfg, ok := e.cfg.Strategies[strategyName]
	if !ok || !cfg.Enabled {
		return false
	}
	for _, s := range cfg.Symbols {
		if s == symbol {
			return true
		}
	}
	return false
}

func (e *Engine) GetSummary() string {
	balances, _ := e.exchange.FetchBalance(context.Background())
	equity := calculateEquity(balances)
	dailyPnL, trades, wins := e.riskMgr.DailyStats()
	positions := e.orderMgr.GetPositions()
	activeSignals := e.signalCon.GetAllActive()

	status := "paused"
	e.mu.RLock()
	if e.running {
		status = "running"
	}
	if e.riskMgr.IsBreached() {
		status = "HALTED (circuit breaker)"
	}
	e.mu.RUnlock()

	return fmt.Sprintf(
		"Status: %s\nEquity: $%.2f\nDaily P&L: $%.2f | Trades: %d | Wins: %d\nPositions: %d | Active Signals: %d",
		status, equity, dailyPnL, trades, wins, len(positions), len(activeSignals),
	)
}

func calculateEquity(balances []types.Balance) float64 {
	total := 0.0
	for _, b := range balances {
		if b.Asset == "USDT" || b.Asset == "USD" || b.Asset == "USDC" {
			total += b.Free + b.Locked
		}
	}
	return total
}

func (e *Engine) GetHoldings() []types.Balance {
	balances, err := e.exchange.FetchBalance(context.Background())
	if err != nil {
		return nil
	}
	return balances
}

func (e *Engine) GetEquity() float64 {
	balances, err := e.exchange.FetchBalance(context.Background())
	if err != nil {
		return 0
	}
	total := 0.0
	for _, b := range balances {
		if b.Asset == "USDT" || b.Asset == "USD" || b.Asset == "USDC" {
			total += b.Free + b.Locked
			continue
		}
		amount := b.Free + b.Locked
		if amount <= 0 {
			continue
		}
		symbol := b.Asset + "USDT"
		ticker, err := e.marketData.FetchTicker(context.Background(), symbol)
		if err == nil && ticker != nil && ticker.Last > 0 {
			total += amount * ticker.Last
		} else {
			total += amount * 0
		}
	}
	return total
}

func (e *Engine) GetCash() float64 {
	balances, err := e.exchange.FetchBalance(context.Background())
	if err != nil {
		return 0
	}
	return calculateEquity(balances)
}

func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

func (e *Engine) IsBreached() bool {
	return e.riskMgr.IsBreached()
}

func (e *Engine) GetPositions() []types.Position {
	return e.orderMgr.GetPositions()
}

func (e *Engine) GetActiveSignals() []*types.Signal {
	return e.signalCon.GetAllActive()
}

func (e *Engine) GetRecentTrades(limit int) []types.Order {
	trades, _ := e.db.GetRecentTrades("", limit)
	return trades
}

func (e *Engine) GetDailyStats() (float64, int, int) {
	return e.riskMgr.DailyStats()
}

func (e *Engine) GetStrategyNames() []string {
	names := make([]string, len(e.strategies))
	for i, s := range e.strategies {
		names[i] = s.Name()
	}
	return names
}

func (e *Engine) GetStrategyPnL() map[string]float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make(map[string]float64, len(e.stratPnL))
	for k, v := range e.stratPnL {
		out[k] = v
	}
	return out
}

func (e *Engine) GetEquityCurve(limit int) []EquityPoint {
	e.mu.RLock()
	defer e.mu.RUnlock()
	start := len(e.equityHist) - limit
	if start < 0 {
		start = 0
	}
	out := make([]EquityPoint, len(e.equityHist[start:]))
	copy(out, e.equityHist[start:])
	return out
}

func (e *Engine) runEquityRecorder(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			eq := e.GetEquity()
			e.mu.Lock()
			e.equityHist = append(e.equityHist, EquityPoint{Time: time.Now(), Value: eq})
			if len(e.equityHist) > 1000 {
				e.equityHist = e.equityHist[len(e.equityHist)-1000:]
			}
			e.mu.Unlock()
		}
	}
}

func (e *Engine) runStopLoss(ctx context.Context, wg *sync.WaitGroup, symbols []string) {
	defer wg.Done()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if e.riskMgr.IsBreached() {
				continue
			}
			for _, symbol := range symbols {
				e.checkStopLoss(ctx, symbol)
			}
		}
	}
}

func (e *Engine) checkStopLoss(ctx context.Context, symbol string) {
	holdings := e.GetHoldings()
	var baseHeld float64
	for _, h := range holdings {
		if h.Asset == "USDT" || h.Asset == "USD" || h.Asset == "USDC" {
			continue
		}
		base := h.Asset + "USDT"
		if base != symbol {
			continue
		}
		baseHeld = h.Free + h.Locked
	}
	if baseHeld <= 0.00001 {
		return
	}

	ticker, err := e.marketData.FetchTicker(ctx, symbol)
	if err != nil || ticker == nil || ticker.Last <= 0 {
		return
	}

	e.mu.RLock()
	entry := e.entryPrice[symbol]
	high := e.highWater[symbol]
	e.mu.RUnlock()

	if entry > 0 && (ticker.Last < entry*0.90 || ticker.Last > entry*1.10) {
		e.logger.Warn().Float64("ticker", ticker.Last).Float64("entry", entry).Str("symbol", symbol).Msg("ticker price anomaly, skipping")
		return
	}

	if entry <= 0 {
		return
	}

	pnlPct := (ticker.Last - entry) / entry

	if pnlPct <= -e.cfg.Risk.StopLossPct {
		amount := baseHeld
		e.logger.Warn().Float64("pnlPct", pnlPct*100).Str("symbol", symbol).Msg("stop-loss triggered")
		e.alerts.Send(alert.LevelWarn, "Stop-Loss Triggered",
			fmt.Sprintf("%s: %.1f%% loss, closing at $%.2f", symbol, pnlPct*100, ticker.Last))

		order, _ := e.orderMgr.PlaceOrder(ctx, symbol, types.SideSell, types.TypeMarket, amount, entry*0.999, "stop-loss")
		e.recordExit(order, entry, symbol, "stop-loss")
		e.mu.Lock()
		e.entryPrice[symbol] = 0
		e.highWater[symbol] = 0
		e.mu.Unlock()
		return
	}

	e.mu.RLock()
	scalpEntryTime := e.entryTime[symbol]
	e.mu.RUnlock()
	if !scalpEntryTime.IsZero() && time.Since(scalpEntryTime) > 30*time.Second && pnlPct > 0 {
		e.logger.Info().Float64("pnlPct", pnlPct*100).Str("symbol", symbol).Msg("time-based exit (30s, in profit)")
		order, _ := e.orderMgr.PlaceOrder(ctx, symbol, types.SideSell, types.TypeMarket, baseHeld, entry*0.999, "time-exit")
		e.recordExit(order, entry, symbol, "time-exit")
		e.mu.Lock()
		e.entryPrice[symbol] = 0
		e.highWater[symbol] = 0
		e.entryTime[symbol] = time.Time{}
		e.mu.Unlock()
		return
	}

	if pnlPct >= e.cfg.Risk.TakeProfitTargetPct {
		e.logger.Info().Float64("pnlPct", pnlPct*100).Str("symbol", symbol).Msg("full take-profit triggered")
		order, _ := e.orderMgr.PlaceOrder(ctx, symbol, types.SideSell, types.TypeMarket, baseHeld, entry*0.999, "take-profit-full")
		e.recordExit(order, entry, symbol, "take-profit-full")
		e.mu.Lock()
		e.entryPrice[symbol] = 0
		e.highWater[symbol] = 0
		e.mu.Unlock()
		return
	}

	if pnlPct >= e.cfg.Risk.TakeProfit1RPct {
		amount := baseHeld * 0.5
		if amount > 0.00001 {
			e.logger.Info().Float64("pnlPct", pnlPct*100).Str("symbol", symbol).Msg("partial take-profit (50%)")
			order, _ := e.orderMgr.PlaceOrder(ctx, symbol, types.SideSell, types.TypeMarket, amount, entry*0.999, "take-profit-50")
			e.recordExit(order, entry, symbol, "take-profit-50")
			e.entryPrice[symbol] = entry
		}
	}

	if pnlPct >= e.cfg.Risk.TrailingStopActivatePct {
		if ticker.Last > high {
			e.mu.Lock()
			e.highWater[symbol] = ticker.Last
			e.mu.Unlock()
			high = ticker.Last
		}
		trailPrice := high * (1 - e.cfg.Risk.TrailingStopDistancePct)
		if ticker.Last <= trailPrice {
			amount := baseHeld
			e.logger.Warn().Float64("high", high).Float64("current", ticker.Last).Str("symbol", symbol).Msg("trailing stop triggered")
			e.alerts.Send(alert.LevelWarn, "Trailing Stop Triggered",
				fmt.Sprintf("%s: locked profit, closing at $%.2f", symbol, ticker.Last))

			order, _ := e.orderMgr.PlaceOrder(ctx, symbol, types.SideSell, types.TypeMarket, amount, entry*0.999, "trailing-stop")
			e.recordExit(order, entry, symbol, "trailing-stop")
			e.mu.Lock()
			e.entryPrice[symbol] = 0
			e.highWater[symbol] = 0
			e.mu.Unlock()
		}
	}
}

func (e *Engine) PauseAll() {
	e.mu.Lock()
	for k := range e.stratPaused {
		e.stratPaused[k] = true
	}
	e.mu.Unlock()
	e.logger.Info().Msg("all strategies paused")
}

func (e *Engine) ResumeAll() {
	e.mu.Lock()
	for k := range e.stratPaused {
		e.stratPaused[k] = false
	}
	e.mu.Unlock()
	e.logger.Info().Msg("all strategies resumed")
}

func (e *Engine) KillAll() {
	e.mu.Lock()
	for k := range e.stratPaused {
		e.stratPaused[k] = true
	}
	e.mu.Unlock()

	ctx := context.Background()
	symbols := e.collectSymbols()
	for _, symbol := range symbols {
		e.orderMgr.CancelAllOpen(ctx, symbol)
	}
	e.riskMgr.MarkBreach()
	e.alerts.Send(alert.LevelBreached, "Kill Switch Activated",
		"All strategies paused, open orders cancelled, circuit breaker tripped")
	e.logger.Warn().Msg("kill switch activated: all strategies paused, orders cancelled, breaker tripped")
}

func (e *Engine) reconcilePositions(ctx context.Context) {
	balances, err := e.exchange.FetchBalance(ctx)
	if err != nil {
		e.logger.Error().Err(err).Msg("reconciliation: failed to fetch balances")
		return
	}
	equity := calculateEquity(balances)
	e.logger.Info().Float64("equity", equity).Msg("reconciliation: balance fetched")

	for _, symbol := range e.collectSymbols() {
		orders, err := e.exchange.FetchOpenOrders(ctx, symbol)
		if err != nil {
			e.logger.Error().Err(err).Str("symbol", symbol).Msg("reconciliation: failed to fetch open orders")
			continue
		}
		if len(orders) > 0 {
			e.logger.Warn().Str("symbol", symbol).Int("count", len(orders)).Msg("reconciliation: open orders found on exchange")
		}
	}
}

func (e *Engine) runPnLSnapshot(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.saveDailySnapshot()
		}
	}
}

func (e *Engine) runAlertMonitor(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	lastTradeCount := 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if e.riskMgr.IsBreached() {
				continue
			}

			equity := e.GetEquity()
			dailyPnL, trades, _ := e.riskMgr.DailyStats()

			if equity > 0 {
				dailyLossPct := -dailyPnL / equity * 100
				if dailyLossPct > e.cfg.Risk.MaxDailyLossPct*50 {
					e.alerts.Send(alert.LevelWarn, "Approaching Daily Loss Limit",
						fmt.Sprintf("Daily P&L: $%.2f (%.1f%% of equity) | Trades: %d", dailyPnL, dailyLossPct, trades))
				}
			}

			if e.tradeCount > 0 && e.tradeCount-lastTradeCount >= 10 {
				e.alerts.Send(alert.LevelInfo, "Trade Milestone",
					fmt.Sprintf("%d trades executed | Daily P&L: $%.2f", e.tradeCount, dailyPnL))
				lastTradeCount = e.tradeCount
			}
		}
	}
}

func (e *Engine) saveDailySnapshot() {
	dailyPnL, trades, wins := e.riskMgr.DailyStats()
	date := time.Now().Format("2006-01-02")
	if err := e.db.RecordDailyPnL(date, dailyPnL, trades, wins); err != nil {
		e.logger.Error().Err(err).Msg("failed to save daily P&L snapshot")
	}
}

func (e *Engine) forwardOrderBook(ob *types.OrderBook) {
	type orderBookSetter interface {
		SetOrderBook(ob *types.OrderBook)
	}
	if setter, ok := e.exchange.(orderBookSetter); ok {
		setter.SetOrderBook(ob)
	}
}

func (e *Engine) recordExit(order *types.Order, entry float64, symbol, strat string) {
	if order == nil || order.Status != types.StatusFilled {
		return
	}
	pnl := (order.AvgPrice - entry) * order.Filled
	e.logger.Info().Float64("pnl", pnl).Str("symbol", symbol).Str("strategy", strat).Msg("exit trade P&L")
	e.db.RecordTrade(order, pnl)
	e.riskMgr.RecordTrade(pnl)
	e.mu.Lock()
	e.tradeCount++
	e.stratPnL[strat] += pnl
	e.mu.Unlock()
}
