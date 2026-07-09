package signal

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"trading-bot/execution/internal/types"
)

const channelTradingSignals = "trading:signals"

type Consumer struct {
	redis  *redis.Client
	logger zerolog.Logger
	mu     sync.RWMutex
	cache  map[string]*types.Signal
}

func NewConsumer(addr string, logger zerolog.Logger) *Consumer {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "",
		DB:           0,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolSize:     1,
		MaxRetries:   1,
	})
	return &Consumer{
		redis:  rdb,
		logger: logger.With().Str("component", "signal-consumer").Logger(),
		cache:  make(map[string]*types.Signal),
	}
}

func (c *Consumer) Subscribe(ctx context.Context) (<-chan *types.Signal, error) {
	ch := make(chan *types.Signal, 100)
	pubsub := c.redis.Subscribe(ctx, channelTradingSignals)

	go func() {
		defer pubsub.Close()
		defer close(ch)

		msgCh := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgCh:
				if !ok {
					return
				}
				sig, err := parseSignal(msg.Payload)
				if err != nil {
					c.logger.Error().Err(err).Msg("parse signal failed")
					continue
				}
				if sig.IsExpired() {
					c.logger.Debug().Str("signal_id", sig.ID).Msg("signal expired on arrival")
					continue
				}

				c.mu.Lock()
				c.cache[sig.Symbol] = sig
				if len(c.cache) > 100 {
					for k := range c.cache {
						if c.cache[k].IsExpired() {
							delete(c.cache, k)
						}
					}
				}
				c.mu.Unlock()

				select {
				case ch <- sig:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	if err := c.redis.Ping(ctx).Err(); err != nil {
		c.logger.Warn().Err(err).Msg("redis not available, signal consumer running without signals")
	}

	return ch, nil
}

func (c *Consumer) GetActiveSignal(symbol string) *types.Signal {
	c.mu.RLock()
	defer c.mu.RUnlock()
	sig, ok := c.cache[symbol]
	if !ok {
		return nil
	}
	if sig.IsExpired() {
		return nil
	}
	return sig
}

func (c *Consumer) GetAllActive() []*types.Signal {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []*types.Signal
	for _, s := range c.cache {
		if !s.IsExpired() {
			result = append(result, s)
		}
	}
	return result
}

func (c *Consumer) Close() error {
	return c.redis.Close()
}

func parseSignal(payload string) (*types.Signal, error) {
	var raw struct {
		ID         string         `json:"id"`
		Timestamp  string         `json:"timestamp"`
		Source     string         `json:"source"`
		Type       string         `json:"type"`
		Symbol     string         `json:"symbol"`
		Direction  string         `json:"direction"`
		Confidence float64        `json:"confidence"`
		Factors    map[string]any `json:"factors"`
		Reason     string         `json:"reason"`
		TTLSeconds int            `json:"ttl_seconds"`
	}
	if err := json.Unmarshal([]byte(payload), &raw); err != nil {
		return nil, err
	}

	ts, err := time.Parse(time.RFC3339, raw.Timestamp)
	if err != nil {
		ts = time.Now()
	}

	return &types.Signal{
		ID:         raw.ID,
		Timestamp:  ts,
		Source:     raw.Source,
		Type:       types.SignalType(raw.Type),
		Symbol:     raw.Symbol,
		Direction:  types.SignalDirection(raw.Direction),
		Confidence: raw.Confidence,
		Factors:    raw.Factors,
		Reason:     raw.Reason,
		TTLSeconds: raw.TTLSeconds,
	}, nil
}
