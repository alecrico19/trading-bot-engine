package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type ExchangeConfig struct {
	APIKey    string `mapstructure:"api_key"`
	APISecret string `mapstructure:"api_secret"`
	Testnet   bool   `mapstructure:"testnet"`
}

type RiskConfig struct {
	MaxPositionPct            float64 `mapstructure:"max_position_pct"`
	MaxDailyLossPct           float64 `mapstructure:"max_daily_loss_pct"`
	DailyProfitTargetPct      float64 `mapstructure:"daily_profit_target_pct"`
	MaxConcurrentPositions    int     `mapstructure:"max_concurrent_positions"`
	CircuitBreakerDrawdownPct float64 `mapstructure:"circuit_breaker_drawdown_pct"`
	MaxOrderSizeUSD           float64 `mapstructure:"max_order_size_usd"`
	MinOrderSizeUSD           float64 `mapstructure:"min_order_size_usd"`
	MaxPriceDeviationPct      float64 `mapstructure:"max_price_deviation_pct"`
	StopLossPct               float64 `mapstructure:"stop_loss_pct"`
	TrailingStopActivatePct   float64 `mapstructure:"trailing_stop_activate_pct"`
	TrailingStopDistancePct   float64 `mapstructure:"trailing_stop_distance_pct"`
	TakeProfit1RPct           float64 `mapstructure:"take_profit_1r_pct"`
	TakeProfitTargetPct       float64 `mapstructure:"take_profit_target_pct"`
}

type StrategyConfig struct {
	Enabled               bool     `mapstructure:"enabled"`
	Symbols               []string `mapstructure:"symbols"`
	OrderBookDepth        int      `mapstructure:"order_book_depth"`
	MinSpreadPct          float64  `mapstructure:"min_spread_pct"`
	PositionSizeUSD       float64  `mapstructure:"position_size_usd"`
	RSIPeriod             int      `mapstructure:"rsi_period"`
	RSIOversold           float64  `mapstructure:"rsi_oversold"`
	RSIOverbought         float64  `mapstructure:"rsi_overbought"`
	BBPeriod              int      `mapstructure:"bb_period"`
	BBStdDev              float64  `mapstructure:"bb_stddev"`
	VolumeSpikeMultiplier float64  `mapstructure:"volume_spike_multiplier"`
}

type ResearchConfig struct {
	UpdateIntervalSeconds int      `mapstructure:"update_interval_seconds"`
	SentimentSources      []string `mapstructure:"sentiment_sources"`
	SignalMinConfidence   float64  `mapstructure:"signal_min_confidence"`
}

type Config struct {
	Exchanges   map[string]ExchangeConfig `mapstructure:"exchanges"`
	Risk        RiskConfig                `mapstructure:"risk"`
	Strategies  map[string]StrategyConfig `mapstructure:"strategies"`
	Research    ResearchConfig            `mapstructure:"research"`
	LiveTrading bool                      `mapstructure:"live_trading"`
	Alerts      AlertConfig               `mapstructure:"alerts"`
}

type AlertConfig struct {
	TelegramToken  string `mapstructure:"telegram_token"`
	TelegramChatID string `mapstructure:"telegram_chat_id"`
	DiscordWebhook string `mapstructure:"discord_webhook"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	for name, ex := range cfg.Exchanges {
		ex.APIKey = resolveEnv(ex.APIKey)
		ex.APISecret = resolveEnv(ex.APISecret)
		cfg.Exchanges[name] = ex
	}

	cfg.ApplyDefaults()
	return &cfg, cfg.Validate()
}

func (c *Config) ApplyDefaults() {
	if c.Risk.MaxPositionPct == 0 {
		c.Risk.MaxPositionPct = 0.05
	}
	if c.Risk.MaxDailyLossPct == 0 {
		c.Risk.MaxDailyLossPct = 0.02
	}
	if c.Risk.MaxConcurrentPositions == 0 {
		c.Risk.MaxConcurrentPositions = 5
	}
	if c.Risk.CircuitBreakerDrawdownPct == 0 {
		c.Risk.CircuitBreakerDrawdownPct = 0.10
	}
	if c.Risk.MaxOrderSizeUSD == 0 {
		c.Risk.MaxOrderSizeUSD = 500
	}
	if c.Risk.MinOrderSizeUSD == 0 {
		c.Risk.MinOrderSizeUSD = 10
	}
	if c.Risk.MaxPriceDeviationPct == 0 {
		c.Risk.MaxPriceDeviationPct = 0.02
	}
	if c.Risk.StopLossPct == 0 {
		c.Risk.StopLossPct = 0.05
	}
	if c.Risk.TrailingStopActivatePct == 0 {
		c.Risk.TrailingStopActivatePct = 0.02
	}
	if c.Risk.TrailingStopDistancePct == 0 {
		c.Risk.TrailingStopDistancePct = 0.03
	}
	if c.Risk.TakeProfit1RPct == 0 {
		c.Risk.TakeProfit1RPct = 0.01
	}
	if c.Risk.TakeProfitTargetPct == 0 {
		c.Risk.TakeProfitTargetPct = 0.02
	}
	if c.Research.SignalMinConfidence == 0 {
		c.Research.SignalMinConfidence = 0.6
	}
	if c.Research.UpdateIntervalSeconds == 0 {
		c.Research.UpdateIntervalSeconds = 60
	}
}

func (c *Config) Validate() error {
	if len(c.Exchanges) == 0 {
		return fmt.Errorf("at least one exchange must be configured")
	}
	if c.Risk.MaxPositionPct <= 0 || c.Risk.MaxPositionPct > 1 {
		return fmt.Errorf("max_position_pct must be between 0 and 1")
	}
	if c.Risk.MaxDailyLossPct <= 0 || c.Risk.MaxDailyLossPct > 1 {
		return fmt.Errorf("max_daily_loss_pct must be between 0 and 1")
	}
	return nil
}

func (c *Config) ValidateLive() error {
	if !c.LiveTrading {
		return fmt.Errorf("live_trading must be explicitly set to true in config")
	}
	for name, ex := range c.Exchanges {
		if ex.APIKey == "" {
			return fmt.Errorf("exchange %s: api_key is required for live trading", name)
		}
		if ex.APISecret == "" {
			return fmt.Errorf("exchange %s: api_secret is required for live trading", name)
		}
		if ex.Testnet && c.LiveTrading {
			return fmt.Errorf("exchange %s: testnet must be false for live trading", name)
		}
	}
	return nil
}

func resolveEnv(s string) string {
	if !strings.HasPrefix(s, "$") {
		return s
	}
	s = strings.TrimPrefix(s, "${")
	s = strings.TrimPrefix(s, "$")
	s = strings.TrimSuffix(s, "}")
	return os.Getenv(s)
}
