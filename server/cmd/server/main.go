package main

import (
	"fmt"
	"log"
	"os"

	"blogwitgoofy/server/internal/config"
	"blogwitgoofy/server/internal/handler"
	"blogwitgoofy/server/internal/model"
	"blogwitgoofy/server/internal/router"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll("uploads", 0o755); err != nil {
		log.Fatalf("创建上传目录失败: %v", err)
	}

	model.InitDB(cfg)
	handler.StartAnalyticsMaintenance()

	r := router.Setup(cfg)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Blog API 已启动: http://localhost%s (API 前缀 /api/v1)", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
