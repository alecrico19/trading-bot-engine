#!/bin/bash
# Clean restart — kills everything, clears cache, starts fresh

echo "=== Killing all bot processes ==="
pkill -9 -f "go run.*main.go" 2>/dev/null
pkill -9 -f "launcher.sh" 2>/dev/null
pkill -9 -f "trading-bot" 2>/dev/null
pkill -9 -f "redis-server" 2>/dev/null
fuser -k 8420/tcp 2>/dev/null
sleep 2

echo "=== Starting fresh ==="
export PATH=$HOME/go/bin:$HOME/go-tools/bin:$PATH
export GOROOT=$HOME/go
go clean -cache 2>/dev/null
rm -f $HOME/trading-bot/execution/trading.db*

exec $HOME/Desktop/start-trading-bot.sh
