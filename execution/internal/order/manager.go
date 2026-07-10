package order

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"trading-bot/execution/internal/exchange"
	"trading-bot/execution/internal/types"
)

type Manager struct {
	exchange   exchange.Exchange
	logger     zerolog.Logger
	mu         sync.RWMutex
	orders     map[string]*types.Order
	positions  map[string]*types.Position
	positionMu sync.RWMutex
}

func NewManager(ex exchange.Exchange, logger zerolog.Logger) *Manager {
	return &Manager{
		exchange:  ex,
		logger:    logger.With().Str("component", "order-manager").Logger(),
		orders:    make(map[string]*types.Order),
		positions: make(map[string]*types.Position),
	}
}

func (m *Manager) PlaceOrder(ctx context.Context, symbol string, side types.OrderSide, orderType types.OrderType, amount, price float64, strategy string) (*types.Order, error) {
	id := uuid.New().String()
	now := time.Now()

	order, err := m.exchange.CreateOrder(ctx, symbol, side, orderType, amount, price)
	if err != nil {
		return nil, err
	}

	order.ID = id
	order.Strategy = strategy
	order.CreatedAt = now
	order.UpdatedAt = now

	m.mu.Lock()
	m.orders[id] = order
	m.mu.Unlock()

	if order.Status == types.StatusFilled || order.Status == types.StatusPartiallyFilled {
		m.updatePosition(order)
	}

	m.logger.Info().
		Str("order_id", id).
		Str("symbol", symbol).
		Str("side", string(side)).
		Str("type", string(orderType)).
		Float64("amount", amount).
		Float64("price", price).
		Str("strategy", strategy).
		Msg("order placed")

	return order, nil
}

func (m *Manager) CancelOrder(ctx context.Context, symbol, orderID string) error {
	m.mu.RLock()
	order, ok := m.orders[orderID]
	m.mu.RUnlock()
	if !ok {
		return m.exchange.CancelOrder(ctx, symbol, orderID)
	}
		return m.exchange.CancelOrder(ctx, symbol, order.ExchangeID)
}

func (m *Manager) CancelAllOpen(ctx context.Context, symbol string) error {
	orders, err := m.exchange.FetchOpenOrders(ctx, symbol)
	if err != nil {
		return err
	}
	for _, o := range orders {
		if err := m.exchange.CancelOrder(ctx, symbol, o.ExchangeID); err != nil {
			m.logger.Error().Err(err).Str("order_id", o.ExchangeID).Msg("cancel failed")
		}
	}
	return nil
}

func (m *Manager) GetOrder(orderID string) (*types.Order, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	o, ok := m.orders[orderID]
	return o, ok
}

func (m *Manager) GetOrders(symbol string) []*types.Order {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*types.Order
	for _, o := range m.orders {
		if o.Symbol == symbol {
			result = append(result, o)
		}
	}
	return result
}

func (m *Manager) GetPositions() []types.Position {
	m.positionMu.RLock()
	defer m.positionMu.RUnlock()
	result := make([]types.Position, 0, len(m.positions))
	for _, p := range m.positions {
		result = append(result, *p)
	}
	return result
}

func (m *Manager) GetPosition(symbol string) (*types.Position, bool) {
	m.positionMu.RLock()
	defer m.positionMu.RUnlock()
	p, ok := m.positions[symbol]
	return p, ok
}

func (m *Manager) FlatPositions() bool {
	m.positionMu.RLock()
	defer m.positionMu.RUnlock()
	return len(m.positions) == 0
}

func (m *Manager) updatePosition(order *types.Order) {
	if order.Side != types.SideBuy && order.Side != types.SideSell {
		return
	}
	m.positionMu.Lock()
	defer m.positionMu.Unlock()

	p, exists := m.positions[order.Symbol]
	if !exists {
		p = &types.Position{Symbol: order.Symbol}
	}

	if order.Side == types.SideBuy {
		newAmount := p.Amount + order.Filled
		if p.Amount < 0 {
			m.closeShort(p, order)
		} else {
			newCost := p.AvgPrice*p.Amount + order.AvgPrice*order.Filled
			p.Amount = newAmount
			if newAmount > 0 {
				p.AvgPrice = newCost / newAmount
			}
		}
	} else {
		newAmount := p.Amount - order.Filled
		if p.Amount > 0 {
			p.RealizedPnL += (order.AvgPrice - p.AvgPrice) * order.Filled
			p.Amount = newAmount
			if p.Amount <= 0 {
				p = &types.Position{Symbol: order.Symbol, RealizedPnL: p.RealizedPnL}
			}
		} else {
			newAmount := p.Amount - order.Filled
			newCost := p.AvgPrice*p.Amount + order.AvgPrice*(-order.Filled)
			p.Amount = newAmount
			if newAmount < 0 && p.Amount < 0 {
				p.AvgPrice = newCost / newAmount
			}
		}
	}

	if abs(p.Amount) < 0.00000001 {
		delete(m.positions, order.Symbol)
	} else {
		p.Side = "long"
		if p.Amount < 0 {
			p.Side = "short"
		}
		m.positions[order.Symbol] = p
	}
}

func (m *Manager) closeShort(p *types.Position, order *types.Order) {
	closeQty := min(abs(p.Amount), order.Filled)
	p.RealizedPnL += (p.AvgPrice - order.AvgPrice) * closeQty
	p.Amount += closeQty
	if abs(p.Amount) < 0.00000001 {
		p.Amount = 0
		p.AvgPrice = 0
	}
	if order.Filled > closeQty {
		p.Amount = order.Filled - closeQty
		p.AvgPrice = order.AvgPrice
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
