package kraken

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	krakenapi "github.com/beldur/kraken-go-api-client"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"trading-bot/execution/internal/config"
	"trading-bot/execution/internal/types"
)

var symbolMap = map[string]string{
	"BTC/USDT": "XXBTZUSDT",
	"ETH/USDT": "XETHZUSDT",
	"BTC/USD":  "XXBTZUSD",
	"ETH/USD":  "XETHZUSD",
	"BTC/EUR":  "XXBTZEUR",
	"ETH/EUR":  "XETHZEUR",
}

type Adapter struct {
	api    *krakenapi.KrakenAPI
	logger zerolog.Logger
}

func New(cfg config.ExchangeConfig, logger zerolog.Logger) *Adapter {
	api := krakenapi.New(cfg.APIKey, cfg.APISecret)
	return &Adapter{
		api:    api,
		logger: logger.With().Str("component", "kraken").Logger(),
	}
}

func (a *Adapter) Name() string { return "kraken" }

func (a *Adapter) toKrakenSymbol(symbol string) string {
	if mapped, ok := symbolMap[symbol]; ok {
		return mapped
	}
	return symbol
}

func (a *Adapter) FetchBalance(ctx context.Context) ([]types.Balance, error) {
	resp, err := a.api.Balance()
	if err != nil {
		return nil, fmt.Errorf("fetch balance: %w", err)
	}

	data, _ := json.Marshal(resp)
	var rawMap map[string]float64
	json.Unmarshal(data, &rawMap)

	balances := make([]types.Balance, 0)
	for asset, amount := range rawMap {
		if amount == 0 {
			continue
		}
		balances = append(balances, types.Balance{
			Asset: asset,
			Free:  amount,
		})
	}
	return balances, nil
}

func (a *Adapter) FetchTicker(ctx context.Context, symbol string) (*types.Ticker, error) {
	krakenSymbol := a.toKrakenSymbol(symbol)
	resp, err := a.api.Ticker(krakenSymbol)
	if err != nil {
		return nil, fmt.Errorf("fetch ticker: %w", err)
	}

	data, _ := json.Marshal(resp)
	var rawMap map[string]krakenapi.PairTickerInfo
	json.Unmarshal(data, &rawMap)

	bid := 0.0
	ask := 0.0
	last := 0.0
	volume := 0.0

	for _, ticker := range rawMap {
		if len(ticker.Bid) >= 3 {
			bid = parseFloat(ticker.Bid[0])
		}
		if len(ticker.Ask) >= 3 {
			ask = parseFloat(ticker.Ask[0])
		}
		if len(ticker.Close) >= 2 {
			last = parseFloat(ticker.Close[0])
		}
		if len(ticker.Volume) >= 2 {
			volume = parseFloat(ticker.Volume[1])
		}
		break
	}

	return &types.Ticker{
		Symbol:    symbol,
		Bid:       bid,
		Ask:       ask,
		Last:      last,
		Volume:    volume,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

func (a *Adapter) FetchOrderBook(ctx context.Context, symbol string, depth int) (*types.OrderBook, error) {
	krakenSymbol := a.toKrakenSymbol(symbol)
	ob, err := a.api.Depth(krakenSymbol, depth)
	if err != nil {
		return nil, fmt.Errorf("fetch order book: %w", err)
	}

	bids := make([]types.OrderBookLevel, len(ob.Bids))
	for i, b := range ob.Bids {
		bids[i] = types.OrderBookLevel{Price: b.Price, Quantity: b.Amount}
	}
	asks := make([]types.OrderBookLevel, len(ob.Asks))
	for i, a := range ob.Asks {
		asks[i] = types.OrderBookLevel{Price: a.Price, Quantity: a.Amount}
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
	krakenSymbol := a.toKrakenSymbol(symbol)

	direction := "buy"
	if side == types.SideSell {
		direction = "sell"
	}

	orderTypeStr := "limit"
	if orderType == types.TypeMarket {
		orderTypeStr = "market"
	}

	args := map[string]string{
		"userref": id,
	}

	resp, err := a.api.AddOrder(krakenSymbol, direction, orderTypeStr, fmt.Sprintf("%.8f", amount), args)
	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	exchangeID := ""
	if len(resp.TransactionIds) > 0 {
		exchangeID = resp.TransactionIds[0]
	}
	if exchangeID == "" && len(resp.Description.Order) > 0 {
		exchangeID = resp.Description.Order
	}

	a.logger.Info().
		Str("order_id", id).
		Str("exchange_id", exchangeID).
		Str("symbol", symbol).
		Str("side", string(side)).
		Msg("order created")

	return &types.Order{
		ID:         id,
		ExchangeID: exchangeID,
		Exchange:   a.Name(),
		Symbol:     symbol,
		Side:       side,
		Type:       orderType,
		Status:     types.StatusPlaced,
		Amount:     amount,
		Price:      price,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (a *Adapter) CancelOrder(ctx context.Context, symbol, orderID string) error {
	_, err := a.api.CancelOrder(orderID)
	return err
}

func (a *Adapter) FetchOrder(ctx context.Context, symbol, orderID string) (*types.Order, error) {
	resp, err := a.api.QueryOrders(orderID, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch order: %w", err)
	}

	for txID, o := range *resp {
		side := types.SideBuy
		if o.Description.Type == "sell" {
			side = types.SideSell
		}

		ordType := types.TypeLimit
		if o.Description.OrderType == "market" {
			ordType = types.TypeMarket
		}

		status := types.StatusPending
		switch o.Status {
		case "open":
			status = types.StatusPlaced
		case "closed":
			status = types.StatusFilled
		case "canceled":
			status = types.StatusCancelled
		}

		return &types.Order{
			ExchangeID: txID,
			Exchange:   a.Name(),
			Symbol:     symbol,
			Side:       side,
			Type:       ordType,
			Status:     status,
			Amount:     parseFloat(o.Volume),
			Filled:     o.VolumeExecuted,
			Price:      o.Price,
		}, nil
	}

	return nil, fmt.Errorf("order not found")
}

func (a *Adapter) FetchOpenOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	resp, err := a.api.OpenOrders(nil)
	if err != nil {
		return nil, fmt.Errorf("fetch open orders: %w", err)
	}

	result := make([]types.Order, 0, len(resp.Open))
	for txID, o := range resp.Open {
		side := types.SideBuy
		if o.Description.Type == "sell" {
			side = types.SideSell
		}
		result = append(result, types.Order{
			ExchangeID: txID,
			Exchange:   a.Name(),
			Symbol:     symbol,
			Side:       side,
			Status:     types.StatusPlaced,
			Amount:     parseFloat(o.Volume),
			Price:      o.Price,
		})
	}
	return result, nil
}

func (a *Adapter) SubscribeOrderBook(ctx context.Context, symbol string) (<-chan *types.OrderBook, error) {
	ch := make(chan *types.OrderBook, 100)

	go func() {
		defer close(ch)
		ticker := time.NewTicker(2 * time.Second)
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

func parseFloat(s string) float64 {
	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result
}
