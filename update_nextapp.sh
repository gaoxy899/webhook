#!/bin/bash

# ========= 配置 =========
APP_DIR="/home/ubuntu/farmnext-app"     # Next.js 项目根目录
LOG_FILE="$APP_DIR/next.log"
PID_FILE="$APP_DIR/next.pid"
BRANCH="main"
REMOTE="origin"
PORT=7080
RUN_CMD="npm run start"
BUILD_CMD="npm run build"

cd "$APP_DIR" || exit 1

# ========= 工具函数 =========
is_running() {
    [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null
}

stop_app() {
    if is_running; then
        echo "🛑 停止应用，PID: $(cat "$PID_FILE")"
        kill "$(cat "$PID_FILE")"
        rm -f "$PID_FILE"
        sleep 1
    else
        echo "ℹ️ 应用未在运行"
        rm -f "$PID_FILE"
    fi
}

start_app() {
    echo "🚀 启动 Next.js 应用..."
    nohup $RUN_CMD > "$LOG_FILE" 2>&1 &
    echo $! > "$PID_FILE"
    echo "✅ 应用已启动，PID: $(cat "$PID_FILE")"
}

restart_app() {
    stop_app
    start_app
}

check_and_update() {
    echo "📦 正在检查远程仓库更新..."
    git fetch $REMOTE

    LOCAL=$(git rev-parse $BRANCH)
    REMOTE_HASH=$(git rev-parse $REMOTE/$BRANCH)

    if [ "$LOCAL" != "$REMOTE_HASH" ]; then
        echo "🔄 检测到更新，开始拉取并重启服务"

        stop_app

        git pull --rebase

        #echo "🔧 执行构建..."
        #$BUILD_CMD

        start_app
    else
        echo "✅ 无需更新，代码已是最新"
    fi
}

# ========= 主逻辑 =========
check_and_update
