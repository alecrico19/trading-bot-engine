package risk

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"trading-bot/execution/internal/config"
	"trading-bot/execution/internal/exchange"
	"trading-bot/execution/internal/order"
	"trading-bot/execution/internal/types"
)

type Manager struct {
	cfg         config.RiskConfig
	exchange    exchange.Exchange
	orderMgr    *order.Manager
	logger      zerolog.Logger
	mu          sync.RWMutex
	dailyPnL    float64
	dailyDate   string
	tradeCount  int
	wins        int
	breach      bool
	startEquity float64
}

func NewManager(cfg config.RiskConfig, ex exchange.Exchange, om *order.Manager, logger zerolog.Logger) *Manager {
	return &Manager{
		cfg:      cfg,
		exchange: ex,
		orderMgr: om,
		logger:   logger.With().Str("component", "risk-manager").Logger(),
	}
}

func (m *Manager) SetStartEquity(equity float64) {
	m.mu.Lock()
	m.startEquity = equity
	m.mu.Unlock()
}

func (m *Manager) IsBreached() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.breach
}

func (m *Manager) MarkBreach() {
	m.mu.Lock()
	m.breach = true
	m.mu.Unlock()
	m.logger.Warn().Msg("risk circuit breaker tripped")
}

func (m *Manager) CanOpenPosition(symbol string, equity float64, confidence float64) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.breach {
		return fmt.Errorf("circuit breaker active: trading halted")
	}

	today := time.Now().Format("2006-01-02")
	if m.dailyDate == today && m.dailyPnL < -equity*m.cfg.MaxDailyLossPct {
		return fmt.Errorf("daily loss limit reached: PnL=%.2f, limit=%.2f%%", m.dailyPnL, m.cfg.MaxDailyLossPct*100)
	}

	positions := m.orderMgr.GetPositions()
	drawdown := 0.0
	if m.startEquity > 0 {
		drawdown = (m.startEquity - equity) / m.startEquity
	}
	if drawdown > m.cfg.CircuitBreakerDrawdownPct {
		return fmt.Errorf("drawdown circuit breaker: %.1f%% > %.1f%%", drawdown*100, m.cfg.CircuitBreakerDrawdownPct*100)
	}

	activeCount := 0
	for _, p := range positions {
		if abs(p.Amount) > 0 {
			activeCount++
		}
	}
	if activeCount >= m.cfg.MaxConcurrentPositions {
		return fmt.Errorf("max concurrent positions reached: %d", activeCount)
	}

	positionSize := equity * m.cfg.MaxPositionPct
	if positionSize > equity {
		return fmt.Errorf("position size %.2f exceeds equity %.2f", positionSize, equity)
	}

	return nil
}

func (m *Manager) RecordTrade(pnl float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	today := time.Now().Format("2006-01-02")
	if m.dailyDate != today {
		m.dailyDate = today
		m.dailyPnL = 0
		m.tradeCount = 0
		m.wins = 0
	}

	m.dailyPnL += pnl
	m.tradeCount++
	if pnl > 0 {
		m.wins++
	}
}

func (m *Manager) DailyStats() (pnl float64, trades int, wins int) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dailyPnL, m.tradeCount, m.wins
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func CalculatePositionSize(symbol string, equity float64, maxPct float64, ticker *types.Ticker) float64 {
	if ticker == nil || ticker.Last <= 0 {
		return 0
	}
	usdSize := equity * maxPct
	return usdSize / ticker.Last
}

func (m *Manager) ValidateOrder(symbol string, side types.OrderSide, orderType types.OrderType, amount, price float64, ticker *types.Ticker) error {
	if ticker == nil {
		return nil
	}

	notional := amount * price
	if orderType == types.TypeMarket && ticker.Last > 0 {
		notional = amount * ticker.Last
	}
	if price > 0 {
		notional = amount * price
	}

	if notional > m.cfg.MaxOrderSizeUSD {
		return fmt.Errorf("order notional $%.2f exceeds max $%.2f", notional, m.cfg.MaxOrderSizeUSD)
	}
	if notional < m.cfg.MinOrderSizeUSD {
		return fmt.Errorf("order notional $%.2f below min $%.2f", notional, m.cfg.MinOrderSizeUSD)
	}

	if price > 0 && ticker.Last > 0 {
		marketPrice := ticker.Last
		pctDev := abs(price-marketPrice) / marketPrice * 100
		if pctDev > m.cfg.MaxPriceDeviationPct*100 {
			return fmt.Errorf("price $%.2f deviates %.1f%% from market $%.2f (max %.1f%%)",
				price, pctDev, marketPrice, m.cfg.MaxPriceDeviationPct*100)
		}
	}

	return nil
}
