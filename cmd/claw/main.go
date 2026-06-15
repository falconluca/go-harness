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

	eng := engine.NewAgentEngine(p, r, workDir, true)

	prompt := `
    请帮我执行以下操作：
    1. 用 bash 查看一下我当前电脑的 Go 版本。
    2. 帮我写一个简单的 helloworld.go 文件，输出 "Hello, harness!"。
    3. 用 bash 编译并运行这个 go 文件，确认它能正常工作。
    `

	err := eng.Run(context.Background(), prompt)
	if err != nil {
		log.Fatalf("引擎崩溃: %v", err)
	}
}
