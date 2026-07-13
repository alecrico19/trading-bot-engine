# Backup Old DB Before Wiping

## Recommendation: Yes, keep it

Move to a dated backup instead of deleting. Simple and useful later:

```bash
mkdir -p ~/trading-bot/backups
mv trading.db "backups/trading-$(date +%Y%m%d-%H%M).db"
```

The launcher then starts fresh with a new DB. Old data preserved in `backups/`.

## Why Keep It

- Compare old vs new session P&L
- Reference if a bug surfaces later
- Audit trail of all trades ever made
- Small files (KB, not MB)

## Files to change

`scripts/launcher.sh` — replace `rm -f trading.db*` with `mkdir -p backups && mv trading.db backups/...` (~3 lines)
`scripts/restart.sh` — same change (~3 lines)