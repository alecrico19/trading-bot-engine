# Live P&L Equity Curve Chart

## What
A real-time line graph on the dashboard showing portfolio value over time — same style as Alpaca's equity chart.

## How

### Engine (`engine.go`)
- New field: `equityHistory []struct{Time time.Time; Value float64}` — rolling buffer, max 500 points
- New goroutine: records equity every 5 seconds via `GetEquity()`
- New method: `GetEquityCurve(limit int) []{time, value}` — returns latest N points

### API (`server.go`)
- New endpoint: `GET /equity-curve?points=200`
- Returns JSON array: `[{"t":"09:30:00","v":1000.42}, ...]`

### Dashboard (`dashboard.html`)
- SVG line chart in a new card between Equity and Positions
- Green line for equity, red shaded area for drawdown
- Auto-updates every 5 seconds
- Zero dependencies — pure SVG generated in JavaScript
- Hover tooltip showing time + value

### Files
| File | Change |
|------|--------|
| `engine.go` | +30 lines (history buffer + recording goroutine + getter) |
| `server.go` | +5 lines (new endpoint) |
| `dashboard.html` | +50 lines (SVG chart section + JS renderer) |

~85 lines total.
