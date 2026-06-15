package main

import (
	"context"
	"log"
	"os"

	"github.com/falconluca/go-harness/internal/engine"
	"github.com/falconluca/go-harness/internal/provider"
	"github.com/falconluca/go-harness/internal/tools"
)

func main() {
	// 确保已设置 ZHIPU_API_KEY
	if os.Getenv("ZHIPU_API_KEY") == "" {
		log.Fatal("请先导出 ZHIPU_API_KEY 环境变量")
	}
	workDir, _ := os.Getwd()

	p := provider.NewZhipuClaudeProvider("glm-5.1")

	r := &tools.MockRegistry{}

	// 组装并启动核心 Engine
	eng := engine.NewAgentEngine(p, r, workDir, true)

	// 设定测试任务
	prompt := "我想去北京跑步，帮我查查天气适合吗？"

	err := eng.Run(context.Background(), prompt)
	if err != nil {
		log.Fatalf("引擎崩溃: %v", err)
	}
}
