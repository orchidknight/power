package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/power/internal/cache"
	"github.com/power/internal/config"
	"github.com/power/internal/server"
	"github.com/power/pkg/citates"
	"github.com/power/pkg/logger"
)

func main() {
	log := logger.NewLog()

	cfg, err := config.NewServerConfig()
	if err != nil {
		log.Fatal("server", "Can't read config: %v", err)
	}

	log.Debug("server", "Config: %v", cfg)

	resourceCache := cache.NewCache(citates.SuccessfulSuccess)

	srv := server.NewServer(cfg, resourceCache, log)

	err = srv.Listen()
	if err != nil {
		log.Fatal("server", "Listen err: %v", err)
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	<-signalChannel
	srv.Shutdown()
}
