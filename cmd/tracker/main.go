package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"reddit-stats/internal/api"
	"reddit-stats/internal/config"
	"reddit-stats/internal/stats"
)

func main() {
	logFile, err := os.OpenFile("reddit-stats.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("Error opening log file:", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := api.NewRedditClient(cfg)
	if err != nil {
		log.Fatal("Error creating Reddit client:", err)
	}

	collector := stats.NewCollector(client)
	reporter := stats.NewReporter()

	go collector.Start(ctx)
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				reporter.Report(collector.Stats())
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down gracefully...")
	cancel()
	time.Sleep(time.Second)
}
