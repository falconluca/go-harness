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
	if os.Getenv("ZHIPU_API_KEY") == "" {
		log.Fatal("请先导出 ZHIPU_API_KEY 环境变量")
	}
	workDir, _ := os.Getwd()

	p := provider.NewZhipuOpenAIProvider("deepseek-v4-flash")

	r := tools.NewRegistry()
	readFileTool := tools.NewReadFileTool(workDir)
	r.Register(readFileTool)

	eng := engine.NewAgentEngine(p, r, workDir, true)

	prompt := "请调用工具读取一下当前工作区目录下 hello.txt 文件的内容，并用一句话向我总结它说了什么。"

	err := eng.Run(context.Background(), prompt)
	if err != nil {
		log.Fatalf("引擎崩溃: %v", err)
	}
}
