package coinbase

import (
	"context"
	"fmt"
	"time"

	coinbasepro "github.com/preichenberger/go-coinbasepro/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"trading-bot/execution/internal/config"
	"trading-bot/execution/internal/types"
)

type Adapter struct {
	client *coinbasepro.Client
	logger zerolog.Logger
}

func New(cfg config.ExchangeConfig, logger zerolog.Logger) *Adapter {
	client := coinbasepro.NewClient()
	if cfg.Testnet {
		client.BaseURL = "https://api-public.sandbox.pro.coinbase.com"
	}
	client.Key = cfg.APIKey
	client.Passphrase = cfg.APISecret
	client.Secret = cfg.APISecret

	return &Adapter{
		client: client,
		logger: logger.With().Str("component", "coinbase").Logger(),
	}
}

func (a *Adapter) Name() string { return "coinbase" }

func (a *Adapter) FetchBalance(ctx context.Context) ([]types.Balance, error) {
	accounts, err := a.client.GetAccounts()
	if err != nil {
		return nil, fmt.Errorf("fetch accounts: %w", err)
	}

	balances := make([]types.Balance, 0, len(accounts))
	for _, acct := range accounts {
		free := parseFloat64(acct.Balance)
		locked := parseFloat64(acct.Hold)
		if free+locked == 0 {
			continue
		}
		balances = append(balances, types.Balance{
			Asset:  acct.Currency,
			Free:   free,
			Locked: locked,
		})
	}
	return balances, nil
}

func (a *Adapter) FetchTicker(ctx context.Context, symbol string) (*types.Ticker, error) {
	ticker, err := a.client.GetTicker(symbol)
	if err != nil {
		return nil, fmt.Errorf("fetch ticker: %w", err)
	}

	return &types.Ticker{
		Symbol:    symbol,
		Bid:       parseFloat64(ticker.Bid),
		Ask:       parseFloat64(ticker.Ask),
		Last:      parseFloat64(ticker.Price),
		Volume:    parseFloat64(string(ticker.Volume)),
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

func (a *Adapter) FetchOrderBook(ctx context.Context, symbol string, depth int) (*types.OrderBook, error) {
	book, err := a.client.GetBook(symbol, 2)
	if err != nil {
		return nil, fmt.Errorf("fetch order book: %w", err)
	}

	bids := make([]types.OrderBookLevel, min(len(book.Bids), depth))
	for i := 0; i < min(len(book.Bids), depth); i++ {
		bids[i] = types.OrderBookLevel{
			Price:    parseFloat64(book.Bids[i].Price),
			Quantity: parseFloat64(book.Bids[i].Size),
		}
	}

	asks := make([]types.OrderBookLevel, min(len(book.Asks), depth))
	for i := 0; i < min(len(book.Asks), depth); i++ {
		asks[i] = types.OrderBookLevel{
			Price:    parseFloat64(book.Asks[i].Price),
			Quantity: parseFloat64(book.Asks[i].Size),
		}
	}

	return &types.OrderBook{
		Symbol:    symbol,
		Bids:      bids,
		Asks:      asks,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

func (a *Adapter) CreateOrder(ctx context.Context, symbol string, side types.OrderSide, orderType types.OrderType, amount, price float64) (*types.Order, error) {
	id := uuid.New().String()
	now := time.Now()

	orderSide := "buy"
	if side == types.SideSell {
		orderSide = "sell"
	}

	orderTypeStr := "limit"
	if orderType == types.TypeMarket {
		orderTypeStr = "market"
	}

	order := coinbasepro.Order{
		Type:      orderTypeStr,
		Side:      orderSide,
		ProductID: symbol,
		Size:      fmt.Sprintf("%.8f", amount),
		ClientOID: id,
	}
	if orderType == types.TypeLimit {
		order.Price = fmt.Sprintf("%.2f", price)
		order.TimeInForce = "GTC"
	}

	result, err := a.client.CreateOrder(&order)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	a.logger.Info().
		Str("order_id", id).
		Str("exchange_id", result.ID).
		Str("symbol", symbol).
		Str("side", string(side)).
		Msg("order created")

	return &types.Order{
		ID:         id,
		ExchangeID: result.ID,
		Exchange:   a.Name(),
		Symbol:     symbol,
		Side:       side,
		Type:       orderType,
		Status:     parseStatus(result),
		Amount:     amount,
		Filled:     parseFloat64(result.Size),
		Price:      price,
		AvgPrice:   parseFloat64(result.Price),
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (a *Adapter) CancelOrder(ctx context.Context, symbol, orderID string) error {
	return a.client.CancelOrder(orderID)
}

func (a *Adapter) FetchOrder(ctx context.Context, symbol, orderID string) (*types.Order, error) {
	order, err := a.client.GetOrder(orderID)
	if err != nil {
		return nil, fmt.Errorf("fetch order: %w", err)
	}

	side := types.SideBuy
	if order.Side == "sell" {
		side = types.SideSell
	}
	orderType := types.TypeLimit
	if order.Type == "market" {
		orderType = types.TypeMarket
	}

	return &types.Order{
		ExchangeID: order.ID,
		Exchange:   a.Name(),
		Symbol:     symbol,
		Side:       side,
		Type:       orderType,
		Status:     parseStatus(order),
		Amount:     parseFloat64(order.Size),
		Filled:     parseFloat64(order.FilledSize),
		Price:      parseFloat64(order.Price),
		AvgPrice:   parseFloat64(order.ExecutedValue) / max(parseFloat64(order.FilledSize), 0.00000001),
	}, nil
}

func (a *Adapter) FetchOpenOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	cursor := a.client.ListOrders()
	var coinbaseOrders []coinbasepro.Order
	if err := cursor.NextPage(&coinbaseOrders); err != nil {
		return nil, fmt.Errorf("fetch open orders: %w", err)
	}

	result := make([]types.Order, 0)
	for _, o := range coinbaseOrders {
		if o.ProductID != symbol || (o.Status != "open" && o.Status != "pending") {
			continue
		}
		side := types.SideBuy
		if o.Side == "sell" {
			side = types.SideSell
		}
		result = append(result, types.Order{
			ExchangeID: o.ID,
			Exchange:   a.Name(),
			Symbol:     symbol,
			Side:       side,
			Status:     parseStatus(o),
			Amount:     parseFloat64(o.Size),
			Filled:     parseFloat64(o.FilledSize),
			Price:      parseFloat64(o.Price),
		})
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
				ob, err := a.FetchOrderBook(ctx, symbol, 20)
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

func parseFloat64(s string) float64 {
	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result
}

func parseStatus(o coinbasepro.Order) types.OrderStatus {
	switch o.Status {
	case "open", "pending":
		return types.StatusPlaced
	case "done":
		if o.DoneReason == "filled" {
			return types.StatusFilled
		}
		return types.StatusCancelled
	case "rejected":
		return types.StatusRejected
	default:
		return types.StatusPending
	}
}
