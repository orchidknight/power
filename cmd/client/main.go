package main

import (
	"github.com/power/internal/client"
	"github.com/power/internal/config"
	"github.com/power/pkg/logger"
)

func main() {
	log := logger.NewLog()
	cfg, err := config.NewClientConfig()
	if err != nil {
		log.Fatal("client", "can't read config: %v", err)
	}

	log.Debug("client", "config: %v", cfg)

	cli := client.NewClient(cfg.ServerAddress, cfg.MaxTries, log)

	err = cli.Connect()
	if err != nil {
		log.Fatal("client", "connect err: %v", err)
	}

	defer cli.Close()

	response, err := cli.GetWisdom()
	if err != nil {
		log.Error("client", "GetWisdom err: %v", err)
	} else {
		log.Info("client", "Wisdom received: %s", response)
	}
}
