package main

import (
	"client/internal/chat"
	"client/internal/config"
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"
)

func main() {
	data, err := os.ReadFile("./configs/config.json")
	if err != nil {
		log.Fatal(err)
	}
	cfg := new(config.Config)
	if err := json.Unmarshal(data, cfg); err != nil {
		log.Fatal(err)
	}
	if err := env.Parse(cfg); err != nil {
		log.Fatal(err)
	}

	logger := log.New()
	loglevel, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	logger.SetLevel(loglevel)

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	for i := 0; i < 10000; i++ {
		c := chat.NewChat(cfg)
		go c.ClientHeandler(ctx, i)
	}

	select {
	case sig := <-sigChan:
		log.Info("Get signal: ", sig)
		cancel()
	case err := <-errChan:
		log.Warn(err)
	}
}
