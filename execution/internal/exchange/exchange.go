package exchange

import (
	"context"

	"trading-bot/execution/internal/types"
)

type Exchange interface {
	Name() string
	FetchBalance(ctx context.Context) ([]types.Balance, error)
	FetchTicker(ctx context.Context, symbol string) (*types.Ticker, error)
	FetchOrderBook(ctx context.Context, symbol string, depth int) (*types.OrderBook, error)
	CreateOrder(ctx context.Context, symbol string, side types.OrderSide, orderType types.OrderType, amount, price float64) (*types.Order, error)
	CancelOrder(ctx context.Context, symbol, orderID string) error
	FetchOrder(ctx context.Context, symbol, orderID string) (*types.Order, error)
	FetchOpenOrders(ctx context.Context, symbol string) ([]types.Order, error)
	SubscribeOrderBook(ctx context.Context, symbol string) (<-chan *types.OrderBook, error)
	SubscribeTrades(ctx context.Context, symbol string) (<-chan *types.Trade, error)
	SubscribeAccount(ctx context.Context) (<-chan *types.Order, error)
}
