#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== Trading Bot ==="
echo ""

case "${1:-run}" in
  run)
    echo "Starting paper trading mode (TUI)..."
    cd "$PROJECT_DIR/execution"
    export GOROOT=/home/alecr/go
    export GOPATH=/home/alecr/go-tools
    export PATH=$GOROOT/bin:$GOPATH/bin:$PATH
    go run cmd/main.go --paper --tui --config "$PROJECT_DIR/config/config.yaml"
    ;;

  headless)
    echo "Starting paper trading mode (headless + API on :8080)..."
    cd "$PROJECT_DIR/execution"
    export GOROOT=/home/alecr/go
    export GOPATH=/home/alecr/go-tools
    export PATH=$GOROOT/bin:$GOPATH/bin:$PATH
    go run cmd/main.go --paper --tui=false --api 8080 --config "$PROJECT_DIR/config/config.yaml"
    ;;

  live)
    echo "Starting live mode with Binance (headless)..."
    cd "$PROJECT_DIR/execution"
    export GOROOT=/home/alecr/go
    export GOPATH=/home/alecr/go-tools
    export PATH=$GOROOT/bin:$GOPATH/bin:$PATH
    go run cmd/main.go --exchange binance --tui=false --config "$PROJECT_DIR/config/config.yaml"
    ;;

  build)
    echo "Building execution engine..."
    mkdir -p "$PROJECT_DIR/bin"
    cd "$PROJECT_DIR/execution"
    export GOROOT=/home/alecr/go
    export GOPATH=/home/alecr/go-tools
    export PATH=$GOROOT/bin:$GOPATH/bin:$PATH
    go build -o "$PROJECT_DIR/bin/trading-bot" cmd/main.go
    echo "Binary: $PROJECT_DIR/bin/trading-bot"
    ;;

  redis)
    echo "Starting Redis..."
    if [ -x "$HOME/redis/redis-server" ]; then
      if "$HOME/redis/redis-cli" ping &>/dev/null; then
        echo "Redis already running"
      else
        "$HOME/redis/redis-server" --daemonize yes --port 6379 --logfile /dev/null
        sleep 1
        echo "Redis running on localhost:6379"
      fi
    else
      echo "Redis not installed at ~/redis/"
    fi
    ;;

  research)
    echo "Starting research service..."
    cd "$PROJECT_DIR"
    PYTHONPATH="$PROJECT_DIR:$PYTHONPATH" python3 research/main.py research/config.yaml
    ;;

  telegram)
    echo "Starting Telegram bot..."
    if [ -z "$TELEGRAM_TOKEN" ]; then
      echo "Error: TELEGRAM_TOKEN environment variable not set"
      exit 1
    fi
    cd "$PROJECT_DIR"
    export AUTHORIZED_USERS="${AUTHORIZED_USERS:-}"
    export API_URL="${API_URL:-http://localhost:8080}"
    python3 telegram/bot.py
    ;;

  backtest)
    echo "Running backtest..."
    cd "$PROJECT_DIR"
    PYTHONPATH="$PROJECT_DIR:$PYTHONPATH" python3 research/backtest/runner.py --symbols "${2:-BTC/USDT,ETH/USDT}" --days "${3:-7}" --optimize --json
    ;;

  all)
    echo "Starting all services (execution + research + telegram)..."
    "$0" redis
    "$0" headless &
    sleep 2
    "$0" research &
    "$0" telegram &
    wait
    ;;

  *)
    echo "Usage: $0 {run|headless|live|build|redis|research|telegram|backtest|all}"
    exit 1
    ;;
esac
