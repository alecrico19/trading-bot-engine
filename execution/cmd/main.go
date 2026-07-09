package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	ossignal "os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog"

	"trading-bot/execution/internal/alert"
	"trading-bot/execution/internal/config"
	"trading-bot/execution/internal/db"
	"trading-bot/execution/internal/engine"
	"trading-bot/execution/internal/exchange"
	binanceadapter "trading-bot/execution/internal/exchange/binance"
	"trading-bot/execution/internal/httpapi"
	"trading-bot/execution/internal/order"
	"trading-bot/execution/internal/risk"
	signalcon "trading-bot/execution/internal/signal"
	"trading-bot/execution/tui"
)

func main() {
	configPath := flag.String("config", "../../config/config.yaml", "path to config file")
	paperMode := flag.Bool("paper", true, "paper trading mode")
	exchangeName := flag.String("exchange", "", "exchange to use (overrides paper mode)")
	redisAddr := flag.String("redis", "localhost:6379", "redis address")
	dbPath := flag.String("db", "trading.db", "sqlite database path")
	initialBalance := flag.Float64("balance", 1000, "initial paper trading balance in USD")
	verbose := flag.Bool("verbose", false, "debug logging")
	useTUI := flag.Bool("tui", true, "launch terminal UI")
	apiPort := flag.Int("api", 0, "HTTP API port (0 = disabled)")
	flag.Parse()

	logger := zerolog.New(os.Stderr).With().Timestamp().Str("service", "execution").Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
	if *verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	store, err := db.NewStore(*dbPath, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to open database")
	}

	var ex exchange.Exchange
	var marketData exchange.Exchange
	if *exchangeName != "" {
		exCfg, ok := cfg.Exchanges[*exchangeName]
		if !ok {
			logger.Fatal().Str("exchange", *exchangeName).Msg("exchange not configured")
		}
		ex = binanceadapter.New(exCfg, logger)
		marketData = ex
		logger.Info().Str("exchange", *exchangeName).Msg("using exchange")
	} else if *paperMode {
		ex = exchange.NewPaperTrader(logger, *initialBalance)
		binanceCfg, hasBinance := cfg.Exchanges["binance"]
		if hasBinance {
			marketData = binanceadapter.New(binanceCfg, logger)
			logger.Info().Msg("paper trading with binance market data")
		} else {
			marketData = ex
			logger.Warn().Msg("paper trading without real market data (add binance exchange to config)")
		}
	} else {
		logger.Fatal().Msg("no exchange configured and paper mode disabled")
	}

	if *exchangeName != "" || !*paperMode {
		if err := cfg.ValidateLive(); err != nil {
			logger.Fatal().Err(err).Msg("live trading validation failed — set live_trading: true and ensure testnet: false in config")
		}
		logger.Warn().Msg("========================================")
		logger.Warn().Msg("  LIVE TRADING MODE — REAL MONEY")
		logger.Warn().Msg("  Confirm by setting live_trading: true")
		logger.Warn().Msg("========================================")
	}

	orderMgr := order.NewManager(ex, logger)
	riskMgr := risk.NewManager(cfg.Risk, ex, orderMgr, logger)
	signalCon := signalcon.NewConsumer(*redisAddr, logger)
	alertSvc := alert.NewService(cfg.Alerts.TelegramToken, cfg.Alerts.TelegramChatID, cfg.Alerts.DiscordWebhook, logger)

	eng := engine.New(cfg, ex, marketData, orderMgr, riskMgr, signalCon, store, alertSvc, logger)

	for name, sc := range cfg.Strategies {
		if !sc.Enabled {
			continue
		}
		switch name {
		case "scalping":
			eng.RegisterStrategy(engine.NewScalpingStrategy(sc, logger))
		case "mean-reversion":
			eng.RegisterStrategy(engine.NewMeanReversionStrategy(sc, logger))
		default:
			logger.Warn().Str("strategy", name).Msg("unknown strategy, skipping")
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	ossignal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		logger.Info().Str("signal", sig.String()).Msg("received shutdown signal")
		cancel()
	}()

	if *apiPort > 0 {
		server := httpapi.NewServer(eng)
		go func() {
			addr := fmt.Sprintf(":%d", *apiPort)
			logger.Info().Str("addr", addr).Msg("HTTP API started")
			if err := http.ListenAndServe(addr, server.Handler()); err != nil && err != http.ErrServerClosed {
				logger.Error().Err(err).Msg("HTTP API error")
			}
		}()
	}

	engineDone := make(chan struct{})
	go func() {
		defer close(engineDone)
		if err := eng.Run(ctx); err != nil {
			logger.Error().Err(err).Msg("engine error")
		}
	}()

	if *useTUI {
		m := tui.New(eng)
		p := tea.NewProgram(m, tea.WithAltScreen())

		go func() {
			<-ctx.Done()
			p.Quit()
		}()

		if _, err := p.Run(); err != nil {
			logger.Error().Err(err).Msg("TUI error")
		}
	} else {
		fmt.Printf("\n%s\n\n", eng.GetSummary())
		<-ctx.Done()
	}

	signalCon.Close()
	<-engineDone
	store.Close()
	logger.Info().Msg("engine stopped")
}
