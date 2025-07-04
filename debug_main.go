package main

import (
	"fmt"
	"log"

	"ai-api-gateway/internal/infrastructure/config"
	"ai-api-gateway/internal/infrastructure/logger"
)

func main() {
	fmt.Println("Starting debug version...")

	// 加载配置
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Config loaded: %+v\n", cfg.Server)

	// 初始化日志记录器
	logger.InitGlobalLogger(&cfg.Logging)
	log := logger.GetLogger()

	log.Info("Logger initialized")
	log.WithField("port", cfg.Server.Port).Info("Server configuration")

	fmt.Println("Debug completed successfully!")
}
