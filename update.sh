#!/bin/bash
set -e  # 出错立即退出

# ========= 配置项 =========
REPO_DIR="/home/ubuntu/farm-replease"
APP_NAME="farmapp"
BRANCH="main"
REMOTE="origin"
LOG_FILE="$REPO_DIR/app.log"
PID_FILE="$REPO_DIR/app.pid"

cd "$REPO_DIR" || exit 1

# ========= 函数 =========

is_running() {
    [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null
}

stop_app() {
    if is_running; then
        echo "停止运行中的应用，PID: $(cat "$PID_FILE")"
        kill "$(cat "$PID_FILE")"
        sleep 2
        if is_running; then
            echo "应用未能正常停止，强制杀死"
            kill -9 "$(cat "$PID_FILE")"
        fi
        rm -f "$PID_FILE"
    else
        echo "应用未在运行，跳过停止"
        rm -f "$PID_FILE"
    fi
}

start_app() {
    echo "启动新版本应用..."
    nohup "$REPO_DIR/$APP_NAME" > "$LOG_FILE" 2>&1 &
    echo $! > "$PID_FILE"
    echo "新版本已启动，PID: $(cat "$PID_FILE")"
}

check_and_update() {
    echo "[$(date)] 正在检查远程分支更新..."

    git fetch $REMOTE || { echo "❌ git fetch 失败"; exit 1; }
    git checkout $BRANCH || { echo "❌ 分支切换失败"; exit 1; }

    LOCAL=$(git rev-parse HEAD)
    REMOTE_HASH=$(git rev-parse $REMOTE/$BRANCH)

    if [ "$LOCAL" != "$REMOTE_HASH" ]; then
        echo "检测到远程更新，准备更新"
        stop_app

        echo "执行 git reset --hard"
        git reset --hard $REMOTE/$BRANCH

        echo "重启应用"
        start_app
    else
        echo "无更新，跳过"
    fi
}

# ========= 主逻辑 =========
check_and_update
