#!/bin/bash

# Configuration
REPO_ROOT="$(pwd)"
BIN_DIR="$REPO_ROOT/bin"
BINARY_NAME="yt"
FRONTEND_DIR="$REPO_ROOT/frontend"
DEFAULT_WEB_ADDR="127.0.0.1:8080"
alias yt="$BIN_DIR/yt"

# Colors and Formatting
GREEN='\033[1;32m'
RED='\033[1;31m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Usage function
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo "Options:"
    echo "  -w, --web    Automatically launch the UI after building"
    echo "  -h, --help   Show this help message"
    exit 1
}

# 1. Strict Verification Check
# Matches 'module github.com/lvcoi/ytdl-go' regardless of extra spaces or tabs
IS_YTDL=false
if [ -f "$REPO_ROOT/go.mod" ]; then
    if grep -qE "^module[[:space:]]+github\.com/lvcoi/ytdl-go([[:space:]]|$)" "$REPO_ROOT/go.mod"; then
        IS_YTDL=true
    fi
fi

if [ "$IS_YTDL" = false ]; then
    echo -e "${RED}Error: This script must be run from the root of the correct repository.${NC}"
    echo "Reason: Could not find 'go.mod' with 'module github.com/lvcoi/ytdl-go'."
    exit 1
fi

# Flag handling
AUTO_LAUNCH=false
WEB_HOST="127.0.0.1"
WEB_PORT="8080"

while [[ "$#" -gt 0 ]]; do
    case $1 in
        -w|--web) 
            AUTO_LAUNCH=true 
            ;;
        -p|--port)
            WEB_PORT="$2"
            shift
            ;;
        -H|--host)
            WEB_HOST="$2"
            shift
            ;;
        -h|--help)
            usage
            ;;
        *) 
            echo -e "${RED}Error: Unknown parameter: $1${NC}"
            usage
            ;;
    esac
    shift
done


# Spinner/Progress Function
run_with_feedback() {
    local task_name=$1
    shift
    if [ "$#" -eq 0 ]; then
        echo -e "${RED}Error: No command provided for task '${task_name}'.${NC}"
        exit 1
    fi

    local pid
    local dots=""
    local spinner_chars="/-\|"
    local i=0
    local start_time=$SECONDS
    local tmp_err

    tmp_err=$(mktemp)
    "$@" 2>"$tmp_err" &
    pid=$!

    tput civis # Hide cursor

    while kill -0 $pid 2>/dev/null; do
        i=$(( (i+1) % 4 ))
        local elapsed=$(( SECONDS - start_time ))
        if [ "$elapsed" -gt 0 ]; then
            dots=$(printf '%.0s.' $(seq 1 $elapsed))
        else
            dots=""
        fi
        
        printf "\r${spinner_chars:$i:1} %s {%s}" "$task_name" "$dots"
        sleep 0.2
    done

    wait $pid
    local exit_code=$?
    
    tput cnorm # Show cursor
    printf "\r\033[K" # Clear line

    if [ $exit_code -eq 0 ]; then
        echo -e "${GREEN}${BOLD}GO${NC}"
        rm -f "$tmp_err"
        return 0
    else
        echo -e "${RED}${BOLD}NO-GO${NC}"
        cat "$tmp_err"
        rm -f "$tmp_err"
        exit 1
    fi
}

run_frontend_build() {
    npm install && npm run build
}

# 2. Build Go Backend
mkdir -p "$BIN_DIR"
run_with_feedback "Building Go Backend" go build -o "./bin/$BINARY_NAME"

# 3. Transient Alias
# Active only for the duration of this script's process/subshell
shopt -s expand_aliases
alias yt="$BIN_DIR/$BINARY_NAME"

# 4. Build Frontend
if [ -d "$FRONTEND_DIR" ]; then
    cd "$FRONTEND_DIR" || exit
    run_with_feedback "Building NPM Frontend" run_frontend_build
else
    echo -e "${RED}Warning: Frontend directory not found.${NC}"
fi

# 5. Launch Logic
launch_ui() {
    local requested_web_addr="${WEB_HOST}:${WEB_PORT}"
    local backend_pid=""
    local ui_exit_code=0

    cleanup_backend() {
        if [ -n "$backend_pid" ] && kill -0 "$backend_pid" 2>/dev/null; then
            kill "$backend_pid" 2>/dev/null || true
            wait "$backend_pid" 2>/dev/null || true
        fi
    }

    echo -e "${GREEN}Launching backend (requested address: ${requested_web_addr})...${NC}"
    (
        cd "$REPO_ROOT" || exit 1
        "$BIN_DIR/$BINARY_NAME" -web --web-addr "$requested_web_addr"
    ) &
    backend_pid=$!

    trap cleanup_backend EXIT INT TERM

    sleep 0.5
    if ! kill -0 "$backend_pid" 2>/dev/null; then
        echo -e "${RED}Error: Backend failed to start. Check logs above.${NC}"
        wait "$backend_pid" 2>/dev/null || true
        trap - EXIT INT TERM
        return 1
    fi

    # Respect explicit override when provided; otherwise let Vite auto-detect
    # backend fallback ports from the default base target.
    # We always set VITE_API_PROXY_TARGET to match the requested backend address
    # unless specifically overridden.
    local target="http://${requested_web_addr}"
    if [ -n "${VITE_API_PROXY_TARGET:-}" ]; then
        target="$VITE_API_PROXY_TARGET"
    fi

    echo -e "${GREEN}Launching UI (API proxy: ${target})...${NC}"
    env VITE_API_PROXY_TARGET="$target" npm run dev
    ui_exit_code=$?

    cleanup_backend
    trap - EXIT INT TERM
    return "$ui_exit_code"
}


if [ "$AUTO_LAUNCH" = true ]; then
    launch_ui
    exit $?
else
    echo -n "Do you want to launch the UI? (y/n): "
    read -r confirm
    if [[ $confirm == [yY] || $confirm == [yY][eE][sS] ]]; then
        launch_ui
        exit $?
    else
        echo "Exiting. 'yt' binary is located at ./bin/$BINARY_NAME"
        exit 0
    fi
fi
