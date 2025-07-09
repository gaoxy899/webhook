package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ListenAddr   string `yaml:"listen_addr"`
	WebhookPath  string `yaml:"webhook_path"`
	SecretToken  string `yaml:"secret_token"`
	UpdateScript string `yaml:"update_script"`
}

var config Config

func loadConfig(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}
	log.Println("配置加载成功")
}

func main() {
	loadConfig("config.yaml")

	http.HandleFunc(config.WebhookPath, handleWebhook)

	log.Printf("Webhook 服务启动于 %s%s", config.ListenAddr, config.WebhookPath)
	log.Fatal(http.ListenAndServe(config.ListenAddr, nil))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "无法读取请求体", http.StatusBadRequest)
		return
	}

	log.Println("收到 Webhook 请求")

	// 校验 GitHub 签名
	if !verifySignature(r.Header.Get("X-Hub-Signature-256"), body) {
		http.Error(w, "签名无效", http.StatusForbidden)
		return
	}

	// 可选：解析 JSON 验证
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Println("无效 JSON:", err)
		http.Error(w, "无效 JSON", http.StatusBadRequest)
		return
	}

	// 执行更新脚本
	cmd := exec.Command("bash", config.UpdateScript)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Printf("执行更新失败: %s\n%s", err, stderr.String())
		http.Error(w, "更新失败", http.StatusInternalServerError)
		return
	}

	log.Println("更新脚本执行成功")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "更新成功")
}

func verifySignature(signatureHeader string, body []byte) bool {
	if signatureHeader == "" || !strings.HasPrefix(signatureHeader, "sha256=") {
		log.Println("缺少或格式错误的 X-Hub-Signature-256 头")
		return false
	}

	sigHex := strings.TrimPrefix(signatureHeader, "sha256=")
	actualSig, err := hex.DecodeString(sigHex)
	if err != nil {
		log.Println("签名解码失败:", err)
		return false
	}

	mac := hmac.New(sha256.New, []byte(config.SecretToken))
	mac.Write(body)
	expectedSig := mac.Sum(nil)

	// 安全比对
	valid := hmac.Equal(expectedSig, actualSig)
	if !valid {
		log.Println("签名不匹配")
	}
	return valid
}
