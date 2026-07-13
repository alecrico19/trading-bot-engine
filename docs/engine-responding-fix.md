# Engine IS Responding — Status Ticker Fix

## Diagnosed
Engine IS working — API responds, dashboard loads, port 8420 bound. The `Engine not responding...` in the launcher is a **false alarm**. The status ticker checks the API too early — before `go run` finishes compiling and the HTTP server binds.

## Evidence
- Port 8420: BOUND by PID 210145 (engine) ✓
- API `/status`: Returns valid JSON ✓
- Dashboard `/`: Returns HTML ✓
- Trade streams: BTCUSDT + ETHUSDT active ✓
- Strategies: mean-reversion + tick-momentum ✓

## Fix: Make the status ticker wait longer
The launcher's status loop starts checking immediately, but the engine takes ~15s to compile after `go clean -cache`. Add a retry with brief delay in the launcher's status ticker:

```bash
# Instead of immediate check, wait for engine to come online
for i in $(seq 1 10); do
    if curl -s http://localhost:8420/status &>/dev/null; then
        break
    fi
    sleep 2
done
```

## Files
`scripts/launcher.sh` — ~8 lines changed in the status ticker section