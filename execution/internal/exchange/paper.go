package exchange

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"trading-bot/execution/internal/types"
)

type PaperTrader struct {
	mu         sync.RWMutex
	balances   map[string]types.Balance
	orders     map[string]*types.Order
	positions  map[string]*types.Position
	orderBooks map[string]*types.OrderBook
	logger     zerolog.Logger
}

func NewPaperTrader(logger zerolog.Logger, initialBalance float64) *PaperTrader {
	pt := &PaperTrader{
		balances:   make(map[string]types.Balance),
		orders:     make(map[string]*types.Order),
		positions:  make(map[string]*types.Position),
		orderBooks: make(map[string]*types.OrderBook),
		logger:     logger.With().Str("component", "paper").Logger(),
	}
	pt.balances["USDT"] = types.Balance{Asset: "USDT", Free: initialBalance, Locked: 0}
	pt.balances["BTC"] = types.Balance{Asset: "BTC", Free: 0, Locked: 0}
	pt.balances["ETH"] = types.Balance{Asset: "ETH", Free: 0, Locked: 0}
	return pt
}

func (p *PaperTrader) Name() string { return "paper" }

func (p *PaperTrader) SetOrderBook(ob *types.OrderBook) {
	p.mu.Lock()
	p.orderBooks[ob.Symbol] = ob
	p.mu.Unlock()
}

func (p *PaperTrader) FetchBalance(ctx context.Context) ([]types.Balance, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	result := make([]types.Balance, 0, len(p.balances))
	for _, b := range p.balances {
		if b.Asset == "" {
			continue
		}
		result = append(result, b)
	}
	return result, nil
}

func (p *PaperTrader) FetchTicker(ctx context.Context, symbol string) (*types.Ticker, error) {
	p.mu.RLock()
	ob, ok := p.orderBooks[symbol]
	p.mu.RUnlock()
	if !ok {
		return &types.Ticker{Symbol: symbol}, nil
	}
	bestBid := 0.0
	bestAsk := 0.0
	if len(ob.Bids) > 0 {
		bestBid = ob.Bids[0].Price
	}
	if len(ob.Asks) > 0 {
		bestAsk = ob.Asks[0].Price
	}
	last := bestBid
	if last == 0 {
		last = bestAsk
	}

	return &types.Ticker{
		Symbol:    symbol,
		Bid:       bestBid,
		Ask:      bestAsk,
		Last:      last,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

func (p *PaperTrader) FetchOrderBook(ctx context.Context, symbol string, depth int) (*types.OrderBook, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	ob, ok := p.orderBooks[symbol]
	if !ok {
		return &types.OrderBook{Symbol: symbol, Bids: []types.OrderBookLevel{}, Asks: []types.OrderBookLevel{}}, nil
	}
	return ob, nil
}

func (p *PaperTrader) CreateOrder(ctx context.Context, symbol string, side types.OrderSide, orderType types.OrderType, amount, price float64) (*types.Order, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	id := uuid.New().String()
	now := time.Now()

	order := &types.Order{
		ID:         id,
		ExchangeID: id,
		Exchange:   "paper",
		Symbol:     symbol,
		Side:       side,
		Type:       orderType,
		Status:     types.StatusPlaced,
		Amount:     amount,
		Price:      price,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	p.simulateFill(order)
	p.orders[id] = order
	return order, nil
}

func (p *PaperTrader) simulateFill(order *types.Order) {
	ob, ok := p.orderBooks[order.Symbol]
	fillPrice := order.Price

	if ok && len(ob.Asks) > 0 && len(ob.Bids) > 0 {
		if order.Side == types.SideBuy {
			fillPrice = ob.Asks[0].Price
		} else {
			fillPrice = ob.Bids[0].Price
		}
	}
	if fillPrice <= 0 {
		fillPrice = order.Price
	}

	order.Filled = order.Amount
	order.AvgPrice = fillPrice
	order.Price = fillPrice
	order.Status = types.StatusFilled
	order.UpdatedAt = time.Now()

	base, quote := p.splitSymbol(order.Symbol)

	if order.Side == types.SideBuy {
		cost := order.Amount * fillPrice
		if !p.adjustBalance(quote, -cost) {
			order.Status = types.StatusRejected
			order.Filled = 0
			order.AvgPrice = 0
			p.logger.Warn().Float64("cost", cost).Str("quote", quote).Msg("insufficient balance for buy")
			return
		}
		p.adjustBalance(base, order.Amount*0.999)
		order.Fee = order.Amount * fillPrice * 0.001
	} else {
		if !p.adjustBalance(base, -order.Amount) {
			order.Status = types.StatusRejected
			order.Filled = 0
			order.AvgPrice = 0
			p.logger.Warn().Float64("amount", order.Amount).Str("base", base).Msg("insufficient balance for sell")
			return
		}
		p.adjustBalance(quote, order.Amount*fillPrice*0.999)
		order.Fee = order.Amount * fillPrice * 0.001
	}
}

func (p *PaperTrader) adjustBalance(asset string, delta float64) bool {
	b := p.balances[asset]
	if b.Asset == "" {
		b.Asset = asset
	}
	b.Free += delta
	if b.Free < 0 {
		return false
	}
	p.balances[asset] = b
	return true
}

func (p *PaperTrader) splitSymbol(symbol string) (string, string) {
	parts := strings.Split(symbol, "/")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return symbol, "USDT"
}

func (p *PaperTrader) CancelOrder(ctx context.Context, symbol, orderID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	order, ok := p.orders[orderID]
	if !ok {
		return fmt.Errorf("order %s not found", orderID)
	}
	if order.Status == types.StatusFilled || order.Status == types.StatusCancelled {
		return nil
	}
	order.Status = types.StatusCancelled
	order.UpdatedAt = time.Now()
	return nil
}

func (p *PaperTrader) FetchOrder(ctx context.Context, symbol, orderID string) (*types.Order, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	order, ok := p.orders[orderID]
	if !ok {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

func (p *PaperTrader) FetchOpenOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var result []types.Order
	for _, o := range p.orders {
		if o.Symbol == symbol && o.Status == types.StatusPlaced {
			result = append(result, *o)
		}
	}
	return result, nil
}

func (p *PaperTrader) SubscribeOrderBook(ctx context.Context, symbol string) (<-chan *types.OrderBook, error) {
	return nil, fmt.Errorf("paper trader: use external order book feed")
}

func (p *PaperTrader) SubscribeTrades(ctx context.Context, symbol string) (<-chan *types.Trade, error) {
	return nil, fmt.Errorf("paper trader: use external trade feed")
}

func (p *PaperTrader) SubscribeAccount(ctx context.Context) (<-chan *types.Order, error) {
	return nil, fmt.Errorf("paper trader: no account stream")
}
