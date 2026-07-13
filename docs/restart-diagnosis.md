# Why Restarts Don't Apply Fixes — Diagnosis

## Possible Causes

1. **Old engine persists on port 8420** — the launcher's `fuser -k 8420/tcp` might fail silently. New `go run` can't bind, old engine keeps serving stale data from a previous binary.

2. **`go run` caches old build** — Go caches compiled binaries. If the source hash doesn't change (some edits don't change the hash), `go run` reuses the cached binary.

3. **Launcher fails silently** — if `go run` exits with an error, the launcher continues but the old engine stays on 8420. User sees dashboard but it's the OLD engine.

4. **Same `trading.db` across restarts** — the `rm -f trading.db` was recently added. Earlier restarts re-used the same DB with old $0.00 P&L trades mixed in.

5. **The fixes aren't fixing the right problem** — the reversed signals, disabled time-exit, and 4-decimal P&L are all in the latest binary. But the UNDERLYING issue (buying high, selling low) isn't resolved by these fixes.

## Plan: Verify What's Actually Running

Instead of telling you to restart, let me check:
- What binary is serving the API right now
- When it was compiled
- What commit it was built from
- Whether the latest launcher script executed cleanly
- Whether the engine log shows the fixes in action

If the fixes ARE applied and you're still losing money, the strategy itself needs fundamental rethinking, not more patches.