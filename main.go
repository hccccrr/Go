package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/version"
)

var isShuttingDown bool

func main() {
	log.Println("🎵 Starting ShizuMusic Bot...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatal("Config validation failed:", err)
	}

	// Create directories
	if err := createDirectories(cfg); err != nil {
		log.Fatal("Failed to create directories:", err)
	}

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v", sig)
		shutdownHandler(ctx, cancel)
	}()

	// Start bot
	if err := startBot(ctx, cfg); err != nil {
		log.Fatal("Failed to start bot:", err)
	}

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutdown complete. Goodbye!")
}

func startBot(ctx context.Context, cfg *config.Config) error {
	log.Println("✅ All checks completed! Let's start ShizuMusic...")

	// Initialize clients
	log.Println(">> Initializing Telegram clients...")
	client, err := core.NewClient(cfg)
	if err != nil {
		return err
	}

	// Start bot client
	if err := client.StartBot(ctx); err != nil {
		return err
	}

	// Start user client
	if err := client.StartUser(ctx); err != nil {
		return err
	}

	// Initialize database
	log.Println(">> Connecting to database...")
	db, err := core.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	// Initialize NTgCalls
	log.Println(">> Booting NTgCalls...")
	calls := core.NewCalls(client.UserClient)
	if err := calls.Start(); err != nil {
		return err
	}

	// Send boot message
	bootMsg := formatBootMessage()
	if err := client.SendToLogger(bootMsg, cfg.BotPic); err != nil {
		log.Printf("Warning: Failed to send boot message: %v", err)
	}

	log.Printf(">> ShizuMusic [%s] is now online! 🎵", version.Info.ShizuMusic)

	// Keep running
	for !isShuttingDown {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(1 * time.Second):
			// Check if still alive
		}
	}

	return nil
}

func shutdownHandler(ctx context.Context, cancel context.CancelFunc) {
	if isShuttingDown {
		return
	}
	isShuttingDown = true

	log.Println("Shutdown signal received. Stopping ShizuMusic...")

	// Add shutdown logic here
	// - Stop NTgCalls
	// - Disconnect clients
	// - Close database

	cancel()
	log.Printf("ShizuMusic [%s] is now offline!", version.Info.ShizuMusic)
}

func createDirectories(cfg *config.Config) error {
	dirs := []string{cfg.DwlDir, cfg.CacheDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func formatBootMessage() string {
	return `#START

**ShizuMusic Bot is now online!**

**• Version:** ` + version.Info.ShizuMusic + `
**• Go Version:** ` + version.Info.GoVersion + `
**• Gogram:** ` + version.Info.Gogram + `
**• NTgCalls:** ` + version.Info.NTgCalls + `
**• Uptime:** ` + version.GetUptimeString()
}
