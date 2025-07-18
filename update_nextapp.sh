#!/bin/bash

# ========= 配置 =========
APP_DIR="/home/ubuntu/farmweb-app"     # Next.js 项目根目录
LOG_FILE="$APP_DIR/next.log"
PID_FILE="$APP_DIR/next.pid"
BRANCH="main"
REMOTE="origin"
PORT=3000
RUN_CMD="npm run start"
BUILD_CMD="npm run build"
GIT_PROXY=http://127.0.0.1:1080

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
    cd "$APP_DIR" || exit 1
    nohup $RUN_CMD > "$LOG_FILE" 2>&1 &

    # 等待服务启动并监听端口
    echo "⏳ 等待服务监听端口 $PORT ..."
    for i in {1..10}; do
        PID=$(ss -lntp "sport = :$PORT" 2>/dev/null | awk -F 'pid=' '/pid=/ {split($2,a,","); print a[1]; exit}')
        if [ -n "$PID" ]; then
            echo "$PID" > "$PID_FILE"
            echo "✅ 启动成功，PID: $PID，日志: $LOG_FILE"
            return
        fi
        sleep 1
    done

    echo "❌ 启动失败：未检测到监听端口 $PORT 的进程"
}

restart_app() {
    stop_app
    start_app
}
send_ding_message() {
  local content="$1"
  DING_WEBHOOK="https://oapi.dingtalk.com/robot/send?access_token=9518954fbd8300e31e30b465a8eaf763cc443f2926ac4eaa4d5e5a0e0688efc6"
  curl -s "$DING_WEBHOOK" \
    -H 'Content-Type: application/json' \
    -d "{
          \"msgtype\": \"text\",
          \"text\": {
            \"content\": \"${content}\"
          }
        }" >/dev/null
}
check_and_update() {
    echo "📦 正在检查远程仓库更新..."
    git -c http.proxy=$GIT_PROXY fetch $REMOTE

    LOCAL=$(git rev-parse $BRANCH)
    REMOTE_HASH=$(git rev-parse $REMOTE/$BRANCH)

    if [ "$LOCAL" != "$REMOTE_HASH" ]; then
        echo "🔄 检测到更新，开始拉取并重启服务"
        send_ding_message "检测到farmweb更新:$REMOTE_HASH"
        stop_app

        git -c http.proxy=$GIT_PROXY pull --rebase
        if [ $? -ne 0 ]; then
           echo "❌ Git pull 执行失败"
           send_ding_message "❌ Git pull 执行失败，请检查网络或代理设置"
           exit 1
    fi

        #echo "🔧 执行构建..."
        #$BUILD_CMD

        start_app
            send_ding_message "farmweb更新完成"
    else
        echo "✅ 无需更新，代码已是最新"
    fi
}

# ========= 主逻辑 =========
check_and_update