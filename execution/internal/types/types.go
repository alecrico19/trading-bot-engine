package types

import "time"

type OrderSide string
type OrderType string
type OrderStatus string
type SignalType string
type SignalDirection string
type DecisionAction string

const (
	SideBuy  OrderSide = "buy"
	SideSell OrderSide = "sell"

	TypeLimit  OrderType = "limit"
	TypeMarket OrderType = "market"

	StatusPending          OrderStatus = "pending"
	StatusPlaced           OrderStatus = "placed"
	StatusPartiallyFilled  OrderStatus = "partially_filled"
	StatusFilled           OrderStatus = "filled"
	StatusCancelled        OrderStatus = "cancelled"
	StatusRejected         OrderStatus = "rejected"

	SignalTypeBias      SignalType = "bias"
	SignalTypeAlert     SignalType = "alert"
	SignalTypeSentiment SignalType = "sentiment"

	SignalDirectionLong    SignalDirection = "long"
	SignalDirectionShort   SignalDirection = "short"
	SignalDirectionNeutral SignalDirection = "neutral"

	ActionBuy     DecisionAction = "buy"
	ActionSell    DecisionAction = "sell"
	ActionHold    DecisionAction = "hold"
	ActionClose   DecisionAction = "close"
	ActionReduce  DecisionAction = "reduce"
)

type Balance struct {
	Asset  string  `json:"asset"`
	Free   float64 `json:"free"`
	Locked float64 `json:"locked"`
}

type Ticker struct {
	Symbol    string  `json:"symbol"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	Last      float64 `json:"last"`
	Volume    float64 `json:"volume"`
	Timestamp int64   `json:"timestamp"`
}

type OrderBookLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
}

type OrderBook struct {
	Symbol    string            `json:"symbol"`
	Bids      []OrderBookLevel  `json:"bids"`
	Asks      []OrderBookLevel  `json:"asks"`
	Timestamp int64             `json:"timestamp"`
}

type Trade struct {
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	Timestamp int64   `json:"timestamp"`
}

type Order struct {
	ID            string      `json:"id"`
	ExchangeID    string      `json:"exchange_id"`
	Exchange      string      `json:"exchange"`
	Symbol        string      `json:"symbol"`
	Side          OrderSide   `json:"side"`
	Type          OrderType   `json:"type"`
	Status        OrderStatus `json:"status"`
	Amount        float64     `json:"amount"`
	Filled        float64     `json:"filled"`
	Price         float64     `json:"price"`
	AvgPrice      float64     `json:"avg_price"`
	Fee           float64     `json:"fee"`
	Strategy      string      `json:"strategy"`
	SignalID      string      `json:"signal_id"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type Position struct {
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"`
	Amount        float64 `json:"amount"`
	AvgPrice      float64 `json:"avg_price"`
	UnrealizedPnL float64 `json:"unrealized_pnl"`
	RealizedPnL   float64 `json:"realized_pnl"`
}

type Signal struct {
	ID         string            `json:"id"`
	Timestamp  time.Time         `json:"timestamp"`
	Source     string            `json:"source"`
	Type       SignalType        `json:"type"`
	Symbol     string            `json:"symbol"`
	Direction  SignalDirection   `json:"direction"`
	Confidence float64           `json:"confidence"`
	Factors    map[string]any    `json:"factors"`
	Reason     string            `json:"reason"`
	TTLSeconds int               `json:"ttl_seconds"`
}

func (s *Signal) IsExpired() bool {
	if s.TTLSeconds <= 0 {
		return false
	}
	return time.Since(s.Timestamp).Seconds() > float64(s.TTLSeconds)
}

func (s *Signal) IsActionable() bool {
	return !s.IsExpired() && s.Confidence >= 0.6
}

type MarketState struct {
	Symbol    string
	Ticker    *Ticker
	OrderBook *OrderBook
	Trades    []Trade
	Positions []Position
	Balances  []Balance
	Signals   []Signal
}

type Decision struct {
	Action   DecisionAction `json:"action"`
	Symbol   string         `json:"symbol"`
	Side     OrderSide      `json:"side"`
	Amount   float64        `json:"amount"`
	Price    float64        `json:"price"`
	Type     OrderType      `json:"type"`
	Reason   string         `json:"reason"`
	SignalID string         `json:"signal_id"`
	Strategy string         `json:"strategy"`
}
