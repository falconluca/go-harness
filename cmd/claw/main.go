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
	r.Register(tools.NewWriteFileTool(workDir))
	r.Register(tools.NewBashTool(workDir))
	r.Register(tools.NewEditFileTool(workDir))

	eng := engine.NewAgentEngine(p, r, workDir, true)

	prompt := `
    我当前目录下有 a.txt, b.txt, c.txt 三个文件。
    为了节省时间，请你同时一次性读取这三个文件，并将它们的内容综合起来，告诉我它们分别记录了什么领域的信息。
    `

	err := eng.Run(context.Background(), prompt)
	if err != nil {
		log.Fatalf("引擎崩溃: %v", err)
	}
}
