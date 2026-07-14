# Can OpenCode See Images?

## Short Answer: Not with the current model

The deepseek-v4-pro model doesn't support image input. The Read tool can open image files, but the model can't process what's in them. This is a model limitation, not a configuration issue.

## Workarounds

1. **Describe what you see** — "P&L shows -$0.08, BTC balance is 0.0013, chart shows downtrend"
2. **Share trade data** — `curl http://localhost:8420/trades` gives me the raw data
3. **Share the dashboard URL** — I can fetch the API data myself
4. **Use a different model** — if you switch to a model with vision support (GPT-4V, Claude), I could see images

## What You Already Do Well

You've been great at describing what you see — "bot bought high, sold low," "220 insufficient balance warnings," "price peaked to $62,782 then dropped." That verbal description is often more useful than a screenshot because I can act on it directly.