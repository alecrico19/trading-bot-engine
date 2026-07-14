# Dashboard Fix: Portfolio Equity & Positions

## Problem

Dashboard shows misleading numbers:
1. **Equity** shows only USDT cash, not total portfolio (ignores BTC/ETH held)
2. **Positions panel** empty — paper trader balances don't flow through order manager's tracker

## Fix

### Engine (`engine.go`)
- Add `GetHoldings()` — returns full balance array from exchange
- Rewrite `GetEquity()` — sum USDT cash + (BTC balance × BTC price) + (ETH balance × ETH price)
- Expose `equity` (total) and `cash` (available USDT) in status

### HTTP API (`server.go`)
- Add `GET /balance` endpoint
- Add `cash` and `holdings` fields to `/status` response

### Dashboard (`dashboard.html`)
- Show "Total Portfolio: $1,000" at top, then "Cash available: $793"
- Positions panel from balance data: symbol, qty, current price, value, unrealized P&L
- Both `dashboard.html` files (project root + embed copy) kept in sync

## Expected Result

| Before | After |
|--------|-------|
| Equity: $793 | Portfolio: $1,000 (+ BTC/ETH value) |
| Positions: 0 | Positions: 2 (BTC × 0.0016 + ETH × 0.058) |
| Cash not shown | Cash: $793 available |

3 files to change, ~15 lines per file.
