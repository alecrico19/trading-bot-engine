package alpaca

import (
	"context"
	"fmt"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"

	"trading-bot/execution/internal/config"
	"trading-bot/execution/internal/types"
)

type Adapter struct {
	trading *alpaca.Client
	market  *marketdata.Client
	logger  zerolog.Logger
}

func New(cfg config.ExchangeConfig, logger zerolog.Logger) *Adapter {
	baseURL := "https://paper-api.alpaca.markets"
	if !cfg.Testnet {
		baseURL = "https://api.alpaca.markets"
	}

	tradingClient := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    cfg.APIKey,
		APISecret: cfg.APISecret,
		BaseURL:   baseURL,
	})

	marketClient := marketdata.NewClient(marketdata.ClientOpts{
		APIKey:    cfg.APIKey,
		APISecret: cfg.APISecret,
	})

	return &Adapter{
		trading: tradingClient,
		market:  marketClient,
		logger:  logger.With().Str("component", "alpaca").Logger(),
	}
}

func (a *Adapter) Name() string { return "alpaca" }

func (a *Adapter) FetchBalance(ctx context.Context) ([]types.Balance, error) {
	account, err := a.trading.GetAccount()
	if err != nil {
		return nil, fmt.Errorf("fetch account: %w", err)
	}

	cash, _ := account.Cash.Float64()
	buyingPower, _ := account.BuyingPower.Float64()

	return []types.Balance{
		{Asset: "USD", Free: cash, Locked: buyingPower - cash},
	}, nil
}

func (a *Adapter) FetchTicker(ctx context.Context, symbol string) (*types.Ticker, error) {
	quotes, err := a.market.GetQuotes(symbol, marketdata.GetQuotesRequest{})
	if err != nil {
		return nil, fmt.Errorf("fetch quote: %w", err)
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quote for %s", symbol)
	}

	q := quotes[len(quotes)-1]
	return &types.Ticker{
		Symbol:    symbol,
		Bid:       q.BidPrice,
		Ask:       q.AskPrice,
		Last:      (q.BidPrice + q.AskPrice) / 2,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

func (a *Adapter) FetchOrderBook(ctx context.Context, symbol string, depth int) (*types.OrderBook, error) {
	quotes, err := a.market.GetQuotes(symbol, marketdata.GetQuotesRequest{})
	if err != nil {
		return nil, fmt.Errorf("fetch order book: %w", err)
	}
	if len(quotes) == 0 {
		return &types.OrderBook{Symbol: symbol}, nil
	}

	q := quotes[len(quotes)-1]
	return &types.OrderBook{
		Symbol: symbol,
		Bids: []types.OrderBookLevel{
			{Price: q.BidPrice, Quantity: float64(q.BidSize)},
		},
		Asks: []types.OrderBookLevel{
			{Price: q.AskPrice, Quantity: float64(q.AskSize)},
		},
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

func (a *Adapter) CreateOrder(ctx context.Context, symbol string, side types.OrderSide, orderType types.OrderType, amount, price float64) (*types.Order, error) {
	id := uuid.New().String()
	now := time.Now()

	alpacaSide := alpaca.Buy
	if side == types.SideSell {
		alpacaSide = alpaca.Sell
	}

	alpacaType := alpaca.Limit
	if orderType == types.TypeMarket {
		alpacaType = alpaca.Market
	}

	qty := decimal.NewFromFloat(amount)
	limitPrice := decimal.NewFromFloat(price)

	req := alpaca.PlaceOrderRequest{
		Symbol:        symbol,
		Qty:           &qty,
		Side:          alpacaSide,
		Type:          alpacaType,
		TimeInForce:   alpaca.Day,
		ClientOrderID: id,
	}
	if orderType == types.TypeLimit {
		req.LimitPrice = &limitPrice
	}

	order, err := a.trading.PlaceOrder(req)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	orderQty, _ := order.Qty.Float64()
	orderFilledQty, _ := order.FilledQty.Float64()

	orderSide := types.SideBuy
	if order.Side == alpaca.Sell {
		orderSide = types.SideSell
	}
	orderTypeResult := types.TypeLimit
	if order.Type == alpaca.Market {
		orderTypeResult = types.TypeMarket
	}

	a.logger.Info().
		Str("order_id", id).
		Str("exchange_id", order.ID).
		Str("symbol", symbol).
		Str("side", string(orderSide)).
		Msg("order created")

	return &types.Order{
		ID:         id,
		ExchangeID: order.ID,
		Exchange:   a.Name(),
		Symbol:     symbol,
		Side:       orderSide,
		Type:       orderTypeResult,
		Status:     parseStatus(order.Status),
		Amount:     orderQty,
		Filled:     orderFilledQty,
		Price:      price,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (a *Adapter) CancelOrder(ctx context.Context, symbol, orderID string) error {
	return a.trading.CancelOrder(orderID)
}

func (a *Adapter) FetchOrder(ctx context.Context, symbol, orderID string) (*types.Order, error) {
	order, err := a.trading.GetOrder(orderID)
	if err != nil {
		return nil, fmt.Errorf("fetch order: %w", err)
	}
	return a.convertOrder(order), nil
}

func (a *Adapter) FetchOpenOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	orders, err := a.trading.GetOrders(alpaca.GetOrdersRequest{
		Status:  "open",
		Symbols: []string{symbol},
	})
	if err != nil {
		return nil, fmt.Errorf("fetch open orders: %w", err)
	}

	result := make([]types.Order, len(orders))
	for i, o := range orders {
		result[i] = *a.convertOrder(&o)
	}
	return result, nil
}

func (a *Adapter) SubscribeOrderBook(ctx context.Context, symbol string) (<-chan *types.OrderBook, error) {
	ch := make(chan *types.OrderBook, 100)

	go func() {
		defer close(ch)
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ob, err := a.FetchOrderBook(ctx, symbol, 1)
				if err != nil {
					continue
				}
				select {
				case ch <- ob:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

func (a *Adapter) SubscribeTrades(ctx context.Context, symbol string) (<-chan *types.Trade, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *Adapter) SubscribeAccount(ctx context.Context) (<-chan *types.Order, error) {
	return nil, fmt.Errorf("not implemented")
}

func (a *Adapter) convertOrder(order *alpaca.Order) *types.Order {
	side := types.SideBuy
	if order.Side == alpaca.Sell {
		side = types.SideSell
	}
	orderType := types.TypeLimit
	if order.Type == alpaca.Market {
		orderType = types.TypeMarket
	}

	qty, _ := order.Qty.Float64()
	filledQty, _ := order.FilledQty.Float64()
	limitPrice, _ := order.LimitPrice.Float64()

	return &types.Order{
		ExchangeID: order.ID,
		Exchange:   a.Name(),
		Symbol:     order.Symbol,
		Side:       side,
		Type:       orderType,
		Status:     parseStatus(order.Status),
		Amount:     qty,
		Filled:     filledQty,
		Price:      limitPrice,
	}
}

func parseStatus(status string) types.OrderStatus {
	switch status {
	case "new", "accepted", "pending_new", "accepted_for_bidding":
		return types.StatusPlaced
	case "partially_filled":
		return types.StatusPartiallyFilled
	case "filled":
		return types.StatusFilled
	case "canceled", "expired", "done_for_day":
		return types.StatusCancelled
	case "rejected", "suspended", "stopped":
		return types.StatusRejected
	default:
		return types.StatusPending
	}
}
