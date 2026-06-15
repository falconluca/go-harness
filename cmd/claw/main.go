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
	workDir, _ := os.Getwd()

	p := &provider.MockProvider{}
	r := &tools.MockRegistry{}

	// 4. 组装并启动核心 Engine
	eng := engine.NewAgentEngine(p, r, workDir, true)
	err := eng.Run(context.Background(), "帮我检查当前目录的文件")
	if err != nil {
		log.Fatalf("引擎崩溃: %v", err)
	}
}
