package main

import (
	"log"
	"net/http"
	"os"

	"github.com/falconluca/go-harness/internal/engine"
	"github.com/falconluca/go-harness/internal/feishu"
	"github.com/falconluca/go-harness/internal/provider"
	"github.com/falconluca/go-harness/internal/tools"
	"github.com/joho/godotenv"
	"github.com/larksuite/oapi-sdk-go/v3/core/httpserverext"
)

func main() {
	// 加载同目录下的 .env 文件；文件不存在时静默回退到系统环境变量。
	if err := godotenv.Load(); err != nil {
		log.Printf("未找到 .env 文件，将使用系统环境变量: %v", err)
	}

	if os.Getenv("ZHIPU_API_KEY") == "" {
		log.Fatal("请先配置 ZHIPU_API_KEY（见 .env.example）")
	}

	model := os.Getenv("ZHIPU_MODEL")
	if model == "" {
		log.Fatal("请先配置 ZHIPU_MODEL（见 .env.example）")
	}

	p := provider.NewZhipuOpenAIProvider(model)

	workDir, _ := os.Getwd()

	r := tools.NewRegistry()
	r.Register(tools.NewReadFileTool(workDir))
	r.Register(tools.NewWriteFileTool(workDir))
	r.Register(tools.NewBashTool(workDir))
	r.Register(tools.NewEditFileTool(workDir))

	eng := engine.NewAgentEngine(p, r, workDir, true)

	bot := feishu.NewFeishuBot(eng)
	handler := httpserverext.NewEventHandlerFunc(bot.GetEventDispatcher())

	http.HandleFunc("/webhook/event", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "48080"
	}
	log.Printf("服务端已启动，正在监听 :%s 端口\n", port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
