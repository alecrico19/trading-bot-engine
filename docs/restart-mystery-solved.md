# Why Fixes Don't Seem To Apply — The DB Mixing Problem

## The Truth

Your restarts DO apply the fixes. The `go clean -cache` + `rm -f trading.db` in the launcher clears everything. But:

1. **New session starts with 0 trades** — take-profit-50 exits don't repeat anymore
2. **But the old trades are still VISIBLE** if the DB wasn't wiped between sessions
3. **The old -$0.04 take-profit-50 sells from earlier ARE in the DB alongside the new trades**

You're seeing mixed data: old broken-session trades + new fixed-session trades.

## Evidence From Current Session

```
BTC sells in current DB:
  23:31: P&L=+$0.06  strat=mean-reversion  ← NEW (profitable!)
  23:08: P&L=-$0.01  strat=mean-reversion  ← NEW
  22:46: P&L=$0.00   strat=mean-reversion  ← NEW
```

No more take-profit-50 repeat-fire sells! The fix IS working. The old -$0.04 trades you saw earlier were from the PREVIOUS session (before restart).

## The Fix For Confusion

The launcher already runs `rm -f trading.db` before starting. But check if it's actually running:
- The DB might be in a different directory
- The launcher might be failing silently
- Or the restart script doesn't wipe the DB

## Verify

After a restart, check that the DB is fresh:
```bash
curl -s http://localhost:8420/trades | python3 -c "import sys,json; print(len(json.load(sys.stdin)))"
```

Should show 0 trades in a fresh session.

## Bottom Line

The fixes ARE applying. The DB is showing mixed old+new trades which is confusing. A truly clean restart (with DB wipe) would show only new trades with the fix.