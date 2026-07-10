#!/bin/bash
# Trading Bot Launcher — starts all services in a terminal window

set -o pipefail

PROJECT_DIR="$HOME/trading-bot"
LOG_DIR="/tmp/trading-bot-logs"
mkdir -p "$LOG_DIR"

REDIS_BIN="$HOME/redis/redis-server"
REDIS_CLI="$HOME/redis/redis-cli"

export GOROOT="$HOME/go"
export GOPATH="$HOME/go-tools"
export PATH="$GOROOT/bin:$GOPATH/bin:$PATH"

BOLD="\033[1m"
GREEN="\033[32m"
YELLOW="\033[33m"
RED="\033[31m"
CYAN="\033[36m"
RESET="\033[0m"

cleanup() {
    echo ""
    echo -e "${YELLOW}Shutting down all services...${RESET}"
    kill $ENGINE_PID 2>/dev/null
    kill $RESEARCH_PID 2>/dev/null
    kill $TELEGRAM_PID 2>/dev/null
    wait $ENGINE_PID 2>/dev/null
    wait $RESEARCH_PID 2>/dev/null
    wait $TELEGRAM_PID 2>/dev/null
    echo -e "${GREEN}All services stopped.${RESET}"
    exit 0
}

trap cleanup SIGINT SIGTERM EXIT

echo ""
echo -e "${BOLD}${GREEN}╔══════════════════════════════════════╗${RESET}"
echo -e "${BOLD}${GREEN}║       TRADING BOT LAUNCHER           ║${RESET}"
echo -e "${BOLD}${GREEN}╚══════════════════════════════════════╝${RESET}"
echo ""

# Clean stale ports from previous runs
echo -ne "${CYAN}Cleaning stale ports...${RESET} "
fuser -k 8080/tcp 2>/dev/null
echo -e "${GREEN}done${RESET}"

# Start Redis
echo -ne "${CYAN}Starting Redis...${RESET} "
if [ -x "$REDIS_CLI" ] && $REDIS_CLI ping &>/dev/null; then
    echo -e "${GREEN}already running${RESET}"
elif [ -x "$REDIS_BIN" ]; then
    $REDIS_BIN --daemonize yes --port 6379 --logfile /dev/null &>/dev/null
    sleep 1
    if $REDIS_CLI ping &>/dev/null; then
        echo -e "${GREEN}OK${RESET}"
    else
        echo -e "${RED}FAILED${RESET}"
    fi
else
    echo -e "${YELLOW}not installed (signals disabled)${RESET}"
fi

# Start engine
echo -ne "${CYAN}Starting engine (paper mode + API)...${RESET} "
cd "$PROJECT_DIR/execution"
source ~/.bashrc 2>/dev/null
go run cmd/main.go --paper --tui=false --api 8080 --config "$PROJECT_DIR/config/config.yaml" > "$LOG_DIR/engine.log" 2>&1 &
ENGINE_PID=$!
sleep 3

if kill -0 $ENGINE_PID 2>/dev/null; then
    echo -e "${GREEN}OK${RESET} (PID $ENGINE_PID)"
else
    echo -e "${RED}FAILED${RESET}"
fi

# Start research
echo -ne "${CYAN}Starting research service...${RESET} "
cd "$PROJECT_DIR"
PYTHONPATH="$PROJECT_DIR:$PYTHONPATH" python3 research/main.py research/config.yaml > "$LOG_DIR/research.log" 2>&1 &
RESEARCH_PID=$!
sleep 2

if kill -0 $RESEARCH_PID 2>/dev/null; then
    echo -e "${GREEN}OK${RESET} (PID $RESEARCH_PID)"
else
    echo -e "${RED}FAILED${RESET}"
fi

# Start telegram (if configured)
TELEGRAM_PID=""
if [ -n "$TELEGRAM_TOKEN" ]; then
    echo -ne "${CYAN}Starting Telegram bot...${RESET} "
    export AUTHORIZED_USERS="${AUTHORIZED_USERS:-1473027968}"
    API_URL="http://localhost:8080" python3 telegram/bot.py > "$LOG_DIR/telegram.log" 2>&1 &
    TELEGRAM_PID=$!
    sleep 1
    if kill -0 $TELEGRAM_PID 2>/dev/null; then
        echo -e "${GREEN}OK${RESET} (PID $TELEGRAM_PID)"
    else
        echo -e "${RED}FAILED${RESET}"
    fi
else
    echo -e "${YELLOW}Telegram bot skipped (TELEGRAM_TOKEN not set)${RESET}"
fi

echo ""
echo -e "${BOLD}─────────────────────────────────────${RESET}"
echo -e "${GREEN}Engine API:${RESET}    http://localhost:8080/status"
echo -e "${GREEN}Logs:${RESET}         $LOG_DIR/"
echo -e "${YELLOW}Press Ctrl+C to stop all services${RESET}"
echo -e "${BOLD}─────────────────────────────────────${RESET}"
echo ""

# Show live status every 5 seconds
while true; do
    STATUS=$(curl -s http://localhost:8080/status 2>/dev/null)
    if [ -n "$STATUS" ]; then
        TRADES=$(echo "$STATUS" | python3 -c "import sys,json; print(json.load(sys.stdin).get('trades',0))" 2>/dev/null || echo "?")
        PNL=$(echo "$STATUS" | python3 -c "import sys,json; print(json.load(sys.stdin).get('daily_pnl',0))" 2>/dev/null || echo "?")
        EQUITY=$(echo "$STATUS" | python3 -c "import sys,json; print(json.load(sys.stdin).get('equity',0))" 2>/dev/null || echo "?")
        SIGNALS=$(echo "$STATUS" | python3 -c "import sys,json; print(json.load(sys.stdin).get('signals',0))" 2>/dev/null || echo "?")

        PNL_COLOR=$GREEN
        if [ "${PNL:0:1}" = "-" ]; then
            PNL_COLOR=$RED
        fi

        printf "\r${CYAN}Equity: \$${EQUITY}  ${PNL_COLOR}Daily: \$${PNL}${RESET}  ${CYAN}Trades: ${TRADES}  Signals: ${SIGNALS}  ${RESET}"
    else
        printf "\r${RED}Engine not responding...${RESET}"
    fi
    sleep 5
done
