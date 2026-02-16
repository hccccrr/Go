package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/version"
)

func main() {
	fmt.Println("🎵 Starting ShizuMusic Bot...")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal("Config validation failed:", err)
	}

	// Create directories
	os.MkdirAll(cfg.DwlDir, 0755)
	os.MkdirAll(cfg.CacheDir, 0755)
	fmt.Printf("✅  Created directories: [%s %s]\n", cfg.DwlDir, cfg.CacheDir)
	fmt.Println("✅  All checks completed! Let's start ShizuMusic...")

	// Initialize clients
	fmt.Println(">> Initializing Telegram clients...")
	client, err := core.NewClient(cfg)
	if err != nil {
		log.Fatal("Failed to create client:", err)
	}

	ctx := context.Background()

	// Start bot
	if err := client.StartBot(ctx); err != nil {
		log.Fatal("Failed to start bot:", err)
	}

	// Start assistant
	if err := client.StartUser(ctx); err != nil {
		log.Fatal("Failed to start assistant:", err)
	}

	// Database
	fmt.Println(">> Connecting to database...")
	db, err := core.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	fmt.Println(">> Database connection successful!")

	// NTgCalls
	fmt.Println(">> Booting NTgCalls...")
	calls := core.NewCalls(client.UserClient)
	if err := calls.Start(); err != nil {
		log.Fatal("Failed to start NTgCalls:", err)
	}
	fmt.Println("✅  NTgCalls initialized successfully!")

	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	// AUTO-LOAD HANDLERS (Python style!)
	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

	if err := client.LoadHandlers(db); err != nil {
		log.Fatal("Failed to load handlers:", err)
	}

	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	// SEND BOOT MESSAGE
	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

	if cfg.LoggerID != 0 {
		botMe, _ := client.BotClient.GetMe()
		userMe, _ := client.UserClient.GetMe()

		bootMsg := fmt.Sprintf(`
🎵 **ShizuMusic Started!**

✅ **Version:** %s
✅ **Bot:** @%s
✅ **Assistant:** @%s
✅ **Database:** Connected
✅ **NTgCalls:** Ready
✅ **Handlers:** Loaded

**Status:** Bot is now online! ✅
**Time:** %s

Send /start to test!
`, version.Version, botMe.Username, userMe.Username, time.Now().Format("15:04:05"))

		_, err := client.BotClient.SendMessage(cfg.LoggerID, bootMsg, nil)
		if err != nil {
			fmt.Printf("⚠️  Failed to send boot message: %v\n", err)
		}
	}

	// Print status
	fmt.Printf("\n🎵 ShizuMusic [%s] is now online!\n", version.Version)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅  Bot Client:   READY")
	fmt.Println("✅  User Client:  READY")
	fmt.Println("✅  Database:     CONNECTED")
	fmt.Println("✅  NTgCalls:     READY")
	fmt.Println("✅  Handlers:     AUTO-LOADED")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\n📝  Bot is ready! Test with /start")
	fmt.Println("⏸️   Press Ctrl+C to stop\n")

	// Idle
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Cleanup
	fmt.Println("\n⏹️  Shutting down...")
	calls.Stop()
	db.Close()
	client.Stop()
	fmt.Println("✅  Shutdown complete!")
}
