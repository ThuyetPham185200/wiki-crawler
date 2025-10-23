package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"wikicrawler/internal/app"
)

func main() {
	// 1. Create the WikiCrawlerApp
	app := app.NewWikiCrawlerApp()

	// 2. Start the app (starts workers + HTTP server)
	go app.Start()
	log.Println("🚀 WikiCrawlerApp is running...")

	// 3. Graceful shutdown on SIGINT/SIGTERM
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("⚠️ Shutting down WikiCrawlerApp...")
	app.Stop()
}
