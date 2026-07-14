# Gold Trading (XAU/USD) via Forex Broker

## What Those Ads Do

High-frequency scalping on gold — order book imbalance detection, rapid entries/exits, small profit targets (0.1-0.3%). Same pattern our scalping strategy already uses, just on a different instrument.

## Adaptation Path

### Option A: OANDA (easiest API)
- Free demo account, no minimum deposit
- REST API + streaming prices
- Go client: no official one, build from OpenAPI spec (~150 lines)
- Spreads on XAU/USD: ~0.2-0.3%
- Commissions: none (spread-only)

### Option B: Interactive Brokers (pro-level)
- Multi-asset (stocks, futures, forex, crypto)
- Full order book depth (Level 2)
- Go client: `ibapi` or build from TWS API
- More complex, but supports everything from one account
- Minimum: $2,000 deposit

### Option C: MetaTrader 5 bridge
- Most forex brokers support MT5
- Python bridge: `MetaTrader5` package
- Run a Python service that bridges MT5 → Redis → our Go engine
- Zero Go code changes, just a new research-style service
- Fastest to build if MT5 is available

## Our Code Changes

Only one new file per option:

| Option | File | Lines |
|--------|------|-------|
| A (OANDA) | `internal/exchange/oanda/adapter.go` | ~150 |
| B (IBKR) | `internal/exchange/ibkr/adapter.go` | ~200 |
| C (MT5) | `research/mt5_bridge.py` | ~100 |

No changes to strategies, risk manager, engine, or dashboard.

## Recommendation

Option C (MT5 bridge) is fastest to build and works with any broker. Option A (OANDA) is best for a clean Go-only stack. Option B (IBKR) is overkill unless you want stocks + futures + forex from one account.

Which direction?
