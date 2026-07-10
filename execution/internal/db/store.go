package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"

	"trading-bot/execution/internal/types"
)

type Store struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewStore(path string, logger zerolog.Logger) (*Store, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	s := &Store{db: db, logger: logger.With().Str("component", "db").Logger()}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS trades (
			id TEXT PRIMARY KEY,
			exchange TEXT NOT NULL,
			symbol TEXT NOT NULL,
			side TEXT NOT NULL,
			amount REAL NOT NULL,
			price REAL NOT NULL,
			fee REAL DEFAULT 0,
			strategy TEXT DEFAULT '',
			signal_id TEXT DEFAULT '',
			pnl REAL DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS signal_log (
			id TEXT PRIMARY KEY,
			source TEXT NOT NULL,
			type TEXT NOT NULL,
			symbol TEXT NOT NULL,
			direction TEXT NOT NULL,
			confidence REAL NOT NULL,
			action_taken TEXT DEFAULT 'logged',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS daily_pnl (
			date TEXT PRIMARY KEY,
			realized_pnl REAL DEFAULT 0,
			trade_count INTEGER DEFAULT 0,
			win_count INTEGER DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_trades_created ON trades(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_trades_symbol ON trades(symbol)`,
	}
	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return fmt.Errorf("exec %q: %w", q, err)
		}
	}
	return nil
}

func (s *Store) RecordTrade(trade *types.Order, pnl float64) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO trades (id, exchange, symbol, side, amount, price, fee, strategy, signal_id, pnl, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		trade.ID, trade.Exchange, trade.Symbol, string(trade.Side),
		trade.Filled, trade.AvgPrice, trade.Fee,
		trade.Strategy, trade.SignalID, pnl, trade.CreatedAt,
	)
	return err
}

func (s *Store) RecordSignal(sig *types.Signal, action string) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO signal_log (id, source, type, symbol, direction, confidence, action_taken, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		sig.ID, sig.Source, string(sig.Type), sig.Symbol,
		string(sig.Direction), sig.Confidence, action, sig.Timestamp,
	)
	return err
}

func (s *Store) RecordDailyPnL(date string, pnl float64, trades int, wins int) error {
	_, err := s.db.Exec(
		`INSERT INTO daily_pnl (date, realized_pnl, trade_count, win_count)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(date) DO UPDATE SET realized_pnl=excluded.realized_pnl, trade_count=excluded.trade_count, win_count=excluded.win_count`,
		date, pnl, trades, wins,
	)
	return err
}

func (s *Store) GetDailyPnL(date string) (float64, int, int, error) {
	var pnl float64
	var trades, wins int
	err := s.db.QueryRow(
		`SELECT realized_pnl, trade_count, win_count FROM daily_pnl WHERE date = ?`, date,
	).Scan(&pnl, &trades, &wins)
	if err == sql.ErrNoRows {
		return 0, 0, 0, nil
	}
	return pnl, trades, wins, err
}

func (s *Store) GetRecentTrades(symbol string, limit int) ([]types.Order, error) {
	query := `SELECT id, exchange, symbol, side, amount, price, fee, strategy, signal_id, pnl, created_at
		 FROM trades ORDER BY created_at DESC LIMIT ?`
	var rows *sql.Rows
	var err error
	if symbol != "" {
		query = `SELECT id, exchange, symbol, side, amount, price, fee, strategy, signal_id, pnl, created_at
		 FROM trades WHERE symbol = ? ORDER BY created_at DESC LIMIT ?`
		rows, err = s.db.Query(query, symbol, limit)
	} else {
		rows, err = s.db.Query(query, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []types.Order = make([]types.Order, 0)
	for rows.Next() {
		var o types.Order
		var createdAt time.Time
		if err := rows.Scan(&o.ID, &o.Exchange, &o.Symbol, &o.Side, &o.Filled, &o.AvgPrice, &o.Fee, &o.Strategy, &o.SignalID, &o.Pnl, &createdAt); err != nil {
			return nil, err
		}
		o.CreatedAt = createdAt
		trades = append(trades, o)
	}
	return trades, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
