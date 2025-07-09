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

type WebhookConfig struct {
	Path   string `yaml:"path"`
	Script string `yaml:"script"`
}

type Config struct {
	ListenAddr  string          `yaml:"listen_addr"`
	SecretToken string          `yaml:"secret_token"`
	Webhooks    []WebhookConfig `yaml:"webhooks"`
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

	for _, wh := range config.Webhooks {
		path := wh.Path
		script := wh.Script

		http.HandleFunc(path, makeHandler(script))
		log.Printf("监听路径 %s 映射到脚本 %s", path, script)
	}

	log.Printf("Webhook 服务启动 %s", config.ListenAddr)
	log.Fatal(http.ListenAndServe(config.ListenAddr, nil))
}

func makeHandler(scriptPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "无法读取请求体", http.StatusBadRequest)
			return
		}

		if !verifySignature(r.Header.Get("X-Hub-Signature-256"), body) {
			http.Error(w, "签名验证失败", http.StatusForbidden)
			return
		}

		// 可选 JSON 校验
		var payload map[string]interface{}
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "无效 JSON", http.StatusBadRequest)
			return
		}

		// 执行对应脚本
		cmd := exec.Command("bash", scriptPath)
		var out, stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			log.Printf("[%s] 执行失败: %v\n%s", r.URL.Path, err, stderr.String())
			http.Error(w, "执行失败", http.StatusInternalServerError)
			return
		}

		log.Printf("[%s] 执行成功:\n%s", r.URL.Path, out.String())
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "执行成功\n")
	}
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
