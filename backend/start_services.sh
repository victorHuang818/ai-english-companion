#!/bin/bash

# ==============================================================================
#  Starts all Go backend services for the AI Interview project.
#  Supports both running pre-compiled binaries (for server) and go run (local).
# ==============================================================================

# 1. Ensure we are in the backend directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd "$SCRIPT_DIR"

# Set file descriptor limit to maximum allowed for high concurrency load testing
ulimit -n 65535 2>/dev/null || echo -e "\e[33m[Warning] Failed to set ulimit to 65535. If you are running high concurrency tests, please run 'ulimit -n 65535' as root first.\e[0m"

# 2. Check for .env file and load it robustly (handling inline comments and CRLF line endings)
ENV_FILE="$SCRIPT_DIR/.env"
if [ -f "$ENV_FILE" ]; then
    echo -e "\e[36m[Info] Loading environment variables from $ENV_FILE...\e[0m"
    set -a
    source <(sed -e 's/\r$//' -e 's/[[:space:]]*#.*$//' "$ENV_FILE")
    set +a
else
    echo -e "\e[31m[Warning] .env file not found at $ENV_FILE. Services might fail to start if env vars are missing.\e[0m"
fi

# 3. Check for dependencies (using nc or bash /dev/tcp fallback)
check_port() {
    local port=$1
    local name=$2
    if command -v nc &>/dev/null; then
        if nc -z 127.0.0.1 $port &>/dev/null; then
            echo -e "\e[32m[OK] Dependency $name is running on port $port.\e[0m"
            return 0
        fi
    else
        # Fallback to bash tcp redirection if nc is not installed
        if timeout 1 bash -c "cat < /dev/null > /dev/tcp/127.0.0.1/$port" &>/dev/null; then
            echo -e "\e[32m[OK] Dependency $name is running on port $port.\e[0m"
            return 0
        fi
    fi
    echo -e "\e[33m[WARN] Dependency $name (port $port) seems offline. Please make sure it is running.\e[0m"
    return 1
}

echo -e "\n\e[33m[1/3] Checking Infrastructure Dependencies...\e[0m"
check_port 3306 "MySQL"
check_port 6379 "Redis"
check_port 9000 "Minio"

echo -e "\n\e[33m[2/3] Starting Backend Services...\e[0m"

# Array of services in execution order
# Format: Name | ServiceDir | BinaryName | GoEntryFile | ConfigName
SERVICES=(
    "ai-rpc|rpc/ai|ai-rpc|ai.go|ai.yaml"
    "user-rpc|rpc/user|user-rpc|user.go|user.yaml"
    "core-rpc|rpc/core|core-rpc|core.go|core.yaml"
    "interview-api|interview_api|interview_api|interview.go|interview-api.yaml"
    "gateway|gateway|gateway|main.go|gateway.yaml"
)

# Trap Ctrl+C and kill all background processes
PID_LIST=()
cleanup() {
    echo -e "\n\e[31m[Stopping] Terminating all background services...\e[0m"
    for pid in "${PID_LIST[@]}"; do
        if kill -0 $pid 2>/dev/null; then
            kill $pid
        fi
    done
    echo -e "\e[32m[Cleanup] All services stopped.\e[0m"
    exit 0
}
trap cleanup SIGINT SIGTERM

for item in "${SERVICES[@]}"; do
    IFS="|" read -r name path binary file configName <<< "$item"
    
    # Locate configuration file
    # Check etc/<config> first, then fall back to path/etc/<config>
    if [ -f "$SCRIPT_DIR/etc/$configName" ]; then
        configPath="$SCRIPT_DIR/etc/$configName"
    else
        configPath="$SCRIPT_DIR/$path/etc/$configName"
    fi

    # Check if a precompiled binary exists in build/
    binaryPath="$SCRIPT_DIR/build/$binary"
    
    if [ -f "$binaryPath" ]; then
        echo -e "       \e[32m[Binary] Launching precompiled $name...\e[0m"
        # Run binary directly from project root
        "$binaryPath" -f "$configPath" > "${SCRIPT_DIR}/${name}.log" 2>&1 &
    else
        echo -e "       \e[36m[Go Run] Launching service: $name from source...\e[0m"
        # Fall back to running from source if binary is missing
        (
            cd "$SCRIPT_DIR/$path"
            go run "$file" -f "etc/$configName"
        ) > "${SCRIPT_DIR}/${name}.log" 2>&1 &
    fi
    
    pid=$!
    PID_LIST+=($pid)
    echo "       Service $name started with PID $pid (Logs -> ${name}.log)"
    sleep 1.5
done

echo -e "\n\e[32m[3/3] Backend Services Started!\e[0m"
echo "--------------------------------------------------------"
echo "Press Ctrl+C to stop all services."
echo "--------------------------------------------------------"

# Keep script running to wait for SIGINT/SIGTERM
wait
