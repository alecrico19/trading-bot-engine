package binance

import (
	"context"
	"fmt"
	"strings"
	"time"

	binance "github.com/adshao/go-binance/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"trading-bot/execution/internal/config"
	"trading-bot/execution/internal/types"
)

type Adapter struct {
	client  *binance.Client
	testnet bool
	logger  zerolog.Logger
}

func New(cfg config.ExchangeConfig, logger zerolog.Logger) *Adapter {
	binance.UseTestnet = cfg.Testnet
	client := binance.NewClient(cfg.APIKey, cfg.APISecret)
	return &Adapter{
		client:  client,
		testnet: cfg.Testnet,
		logger:  logger.With().Str("component", "binance").Logger(),
	}
}

func (a *Adapter) Name() string {
	if a.testnet {
		return "binance-testnet"
	}
	return "binance"
}

func (a *Adapter) FetchBalance(ctx context.Context) ([]types.Balance, error) {
	account, err := a.client.NewGetAccountService().Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch account: %w", err)
	}

	balances := make([]types.Balance, 0, len(account.Balances))
	for _, b := range account.Balances {
		free := parseFloat(b.Free)
		locked := parseFloat(b.Locked)
		if free+locked == 0 {
			continue
		}
		balances = append(balances, types.Balance{
			Asset:  b.Asset,
			Free:   free,
			Locked: locked,
		})
	}
	return balances, nil
}

func (a *Adapter) FetchTicker(ctx context.Context, symbol string) (*types.Ticker, error) {
	tickers, err := a.client.NewListPricesService().Symbol(symbol).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch ticker: %w", err)
	}
	if len(tickers) == 0 {
		return nil, fmt.Errorf("no ticker for %s", symbol)
	}

	bookTicker, err := a.client.NewListBookTickersService().Symbol(symbol).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch book ticker: %w", err)
	}
	if len(bookTicker) == 0 {
		return nil, fmt.Errorf("no book ticker for %s", symbol)
	}

	stats, err := a.client.NewListPriceChangeStatsService().Symbol(symbol).Do(ctx)
	volume := 0.0
	if err == nil && len(stats) > 0 {
		volume = parseFloat(stats[0].Volume)
	}

	return &types.Ticker{
		Symbol:    symbol,
		Bid:       parseFloat(bookTicker[0].BidPrice),
		Ask:       parseFloat(bookTicker[0].AskPrice),
		Last:      parseFloat(tickers[0].Price),
		Volume:    volume,
		Timestamp: time.Now().UnixMilli(),
	}, nil
}

func (a *Adapter) FetchOrderBook(ctx context.Context, symbol string, depth int) (*types.OrderBook, error) {
	ob, err := a.client.NewDepthService().Symbol(symbol).Limit(depth).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch order book: %w", err)
	}

	bids := make([]types.OrderBookLevel, len(ob.Bids))
	for i, b := range ob.Bids {
		bids[i] = types.OrderBookLevel{
			Price:    parseFloat(b.Price),
			Quantity: parseFloat(b.Quantity),
		}
	}

	asks := make([]types.OrderBookLevel, len(ob.Asks))
	for i, a := range ob.Asks {
		asks[i] = types.OrderBookLevel{
			Price:    parseFloat(a.Price),
			Quantity: parseFloat(a.Quantity),
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

	var binanceSide binance.SideType
	var binanceType binance.OrderType

	switch side {
	case types.SideBuy:
		binanceSide = binance.SideTypeBuy
	case types.SideSell:
		binanceSide = binance.SideTypeSell
	default:
		return nil, fmt.Errorf("invalid side: %s", side)
	}

	var result *binance.CreateOrderResponse
	var err error

	switch orderType {
	case types.TypeLimit:
		binanceType = binance.OrderTypeLimit
		result, err = a.client.NewCreateOrderService().
			Symbol(symbol).
			Side(binanceSide).
			Type(binanceType).
			TimeInForce(binance.TimeInForceTypeGTC).
			Quantity(fmt.Sprintf("%.8f", amount)).
			Price(fmt.Sprintf("%.2f", price)).
			Do(ctx)
	case types.TypeMarket:
		binanceType = binance.OrderTypeMarket
		result, err = a.client.NewCreateOrderService().
			Symbol(symbol).
			Side(binanceSide).
			Type(binanceType).
			Quantity(fmt.Sprintf("%.8f", amount)).
			Do(ctx)
	default:
		return nil, fmt.Errorf("unsupported order type: %s", orderType)
	}

	if err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}

	a.logger.Info().
		Str("order_id", id).
		Str("exchange_id", fmt.Sprintf("%d", result.OrderID)).
		Str("symbol", symbol).
		Str("side", string(side)).
		Str("type", string(orderType)).
		Float64("amount", amount).
		Float64("price", price).
		Msg("order created")

	return &types.Order{
		ID:         id,
		ExchangeID: fmt.Sprintf("%d", result.OrderID),
		Exchange:   a.Name(),
		Symbol:     symbol,
		Side:       side,
		Type:       orderType,
		Status:     parseOrderStatus(string(result.Status)),
		Amount:     amount,
		Filled:     parseFloat(result.ExecutedQuantity),
		Price:      price,
		AvgPrice:   parseFloat(result.Price),
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (a *Adapter) CancelOrder(ctx context.Context, symbol, orderID string) error {
	_, err := a.client.NewCancelOrderService().Symbol(symbol).OrderID(toInt64(orderID)).Do(ctx)
	if err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}
	a.logger.Info().Str("order_id", orderID).Str("symbol", symbol).Msg("order cancelled")
	return nil
}

func (a *Adapter) FetchOrder(ctx context.Context, symbol, orderID string) (*types.Order, error) {
	o, err := a.client.NewGetOrderService().Symbol(symbol).OrderID(toInt64(orderID)).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch order: %w", err)
	}

	avgPrice := 0.0
	filled := parseFloat(o.ExecutedQuantity)
	if filled > 0 {
		avgPrice = parseFloat(o.CummulativeQuoteQuantity) / filled
	}

	return &types.Order{
		ExchangeID: fmt.Sprintf("%d", o.OrderID),
		Exchange:   a.Name(),
		Symbol:     symbol,
		Side:       parseOrderSide(string(o.Side)),
		Type:       parseOrderType(string(o.Type)),
		Status:     parseOrderStatus(string(o.Status)),
		Amount:     parseFloat(o.OrigQuantity),
		Filled:     filled,
		Price:      parseFloat(o.Price),
		AvgPrice:   avgPrice,
	}, nil
}

func (a *Adapter) FetchOpenOrders(ctx context.Context, symbol string) ([]types.Order, error) {
	orders, err := a.client.NewListOpenOrdersService().Symbol(symbol).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch open orders: %w", err)
	}

	result := make([]types.Order, len(orders))
	for i, o := range orders {
		result[i] = types.Order{
			ExchangeID: fmt.Sprintf("%d", o.OrderID),
			Exchange:   a.Name(),
			Symbol:     symbol,
			Side:       parseOrderSide(string(o.Side)),
			Type:       parseOrderType(string(o.Type)),
			Status:     parseOrderStatus(string(o.Status)),
			Amount:     parseFloat(o.OrigQuantity),
			Filled:     parseFloat(o.ExecutedQuantity),
			Price:      parseFloat(o.Price),
		}
	}
	return result, nil
}

func (a *Adapter) SubscribeOrderBook(ctx context.Context, symbol string) (<-chan *types.OrderBook, error) {
	symbolLower := strings.ToLower(symbol)
	ch := make(chan *types.OrderBook, 100)

	errHandler := func(err error) {
		a.logger.Error().Err(err).Str("symbol", symbol).Msg("order book stream error")
	}

	depthHandler := func(event *binance.WsDepthEvent) {
		bids := make([]types.OrderBookLevel, len(event.Bids))
		for i, b := range event.Bids {
			bids[i] = types.OrderBookLevel{
				Price:    parseFloat(b.Price),
				Quantity: parseFloat(b.Quantity),
			}
		}
		asks := make([]types.OrderBookLevel, len(event.Asks))
		for i, a := range event.Asks {
			asks[i] = types.OrderBookLevel{
				Price:    parseFloat(a.Price),
				Quantity: parseFloat(a.Quantity),
			}
		}
		select {
		case ch <- &types.OrderBook{
			Symbol:    symbol,
			Bids:      bids,
			Asks:      asks,
			Timestamp: time.Now().UnixMilli(),
		}:
		case <-ctx.Done():
		}
	}

	done, stop, err := binance.WsDepthServe100Ms(symbolLower, depthHandler, errHandler)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("subscribe order book: %w", err)
	}

	go func() {
		select {
		case <-ctx.Done():
		case <-done:
		}
		select {
		case stop <- struct{}{}:
		default:
		}
	}()

	return ch, nil
}

func (a *Adapter) SubscribeTrades(ctx context.Context, symbol string) (<-chan *types.Trade, error) {
	symbolLower := strings.ToLower(symbol)
	ch := make(chan *types.Trade, 100)

	errHandler := func(err error) {
		a.logger.Error().Err(err).Str("symbol", symbol).Msg("trade stream error")
	}

	tradeHandler := func(event *binance.WsAggTradeEvent) {
		select {
		case ch <- &types.Trade{
			Symbol:    symbol,
			Side:      tradeSide(event.IsBuyerMaker),
			Price:     parseFloat(event.Price),
			Quantity:  parseFloat(event.Quantity),
			Timestamp: event.Time,
		}:
		case <-ctx.Done():
		}
	}

	done, stop, err := binance.WsAggTradeServe(symbolLower, tradeHandler, errHandler)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("subscribe trades: %w", err)
	}

	go func() {
		select {
		case <-ctx.Done():
		case <-done:
		}
		select {
		case stop <- struct{}{}:
		default:
		}
	}()

	return ch, nil
}

func (a *Adapter) SubscribeAccount(ctx context.Context) (<-chan *types.Order, error) {
	ch := make(chan *types.Order, 100)

	listenKey, err := a.client.NewStartUserStreamService().Do(ctx)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("start user stream: %w", err)
	}

	errHandler := func(err error) {
		a.logger.Error().Err(err).Msg("user data stream error")
	}

	execHandler := func(event *binance.WsUserDataEvent) {
		if event.Event == binance.UserDataEventTypeExecutionReport {
			update := event.OrderUpdate
			select {
			case ch <- &types.Order{
				ExchangeID: fmt.Sprintf("%d", update.Id),
				Exchange:   a.Name(),
				Symbol:     update.Symbol,
				Side:       parseOrderSide(update.Side),
				Type:       parseOrderType(update.Type),
				Status:     parseOrderStatus(update.Status),
				Amount:     parseFloat(update.Volume),
				Filled:     parseFloat(update.FilledVolume),
				Price:      parseFloat(update.Price),
				AvgPrice:   parseFloat(update.LatestPrice),
				UpdatedAt:  time.Now(),
			}:
			case <-ctx.Done():
			}
		}
	}

	done, stop, err := binance.WsUserDataServe(listenKey, execHandler, errHandler)
	if err != nil {
		close(ch)
		return nil, fmt.Errorf("subscribe user data: %w", err)
	}

	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				select { case stop <- struct{}{}: default: }
				return
			case <-done:
				select { case stop <- struct{}{}: default: }
				return
			case <-ticker.C:
				a.client.NewKeepaliveUserStreamService().ListenKey(listenKey).Do(ctx)
			}
		}
	}()

	return ch, nil
}

func parseFloat(s string) float64 {
	var result float64
	fmt.Sscanf(s, "%f", &result)
	return result
}

func parseOrderStatus(s string) types.OrderStatus {
	switch s {
	case "NEW":
		return types.StatusPlaced
	case "PARTIALLY_FILLED":
		return types.StatusPartiallyFilled
	case "FILLED":
		return types.StatusFilled
	case "CANCELED":
		return types.StatusCancelled
	case "REJECTED":
		return types.StatusRejected
	case "EXPIRED":
		return types.StatusCancelled
	default:
		return types.StatusPending
	}
}

func parseOrderSide(s string) types.OrderSide {
	switch s {
	case "BUY":
		return types.SideBuy
	case "SELL":
		return types.SideSell
	default:
		return types.SideBuy
	}
}

func parseOrderType(s string) types.OrderType {
	switch s {
	case "LIMIT":
		return types.TypeLimit
	case "MARKET":
		return types.TypeMarket
	default:
		return types.TypeLimit
	}
}

func tradeSide(isBuyerMaker bool) string {
	if isBuyerMaker {
		return string(types.SideSell)
	}
	return string(types.SideBuy)
}

func toInt64(s string) int64 {
	var result int64
	fmt.Sscanf(s, "%d", &result)
	return result
}
