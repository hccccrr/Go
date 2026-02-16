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
	"shizumusic/handlers"
	"shizumusic/version"
)

var (
	isShuttingDown bool
	globalClient   *core.Client
	globalDB       *core.Database
	globalCalls    *core.Calls
)

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
		log.Printf("🛑 Received signal: %v", sig)
		shutdownHandler(ctx, cancel)
	}()

	// Start bot
	if err := startBot(ctx, cfg); err != nil {
		log.Fatal("Failed to start bot:", err)
	}

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("✅ Shutdown complete. Goodbye!")
}

func startBot(ctx context.Context, cfg *config.Config) error {
	log.Println("✅ All checks completed! Let's start ShizuMusic...")

	// Initialize clients
	log.Println(">> Initializing Telegram clients...")
	client, err := core.NewClient(cfg)
	if err != nil {
		return err
	}
	globalClient = client

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
	globalDB = db

	// Load all plugins (handlers will be registered automatically)
	log.Println(">> Loading handler plugins...")
	handlers.LoadAllPlugins(client, db)

	// Initialize NTgCalls only if user client is available
	if client.UserClient != nil {
		log.Println(">> Booting NTgCalls...")
		calls := core.NewCalls(client.UserClient)
		if err := calls.Start(); err != nil {
			log.Printf("⚠️  Failed to start NTgCalls: %v", err)
			log.Println("   Voice chat features will not be available")
		} else {
			globalCalls = calls
			log.Println("✅ NTgCalls initialized successfully!")
		}
	} else {
		log.Println("⚠️  User client not available - NTgCalls disabled")
		log.Println("   Voice chat features will not work")
		log.Println("   Add STRING_SESSION to enable voice chat")
	}

	// Send boot message
	bootMsg := formatBootMessage()
	if err := client.SendToLogger(bootMsg, cfg.BotPic); err != nil {
		log.Printf("⚠️  Failed to send boot message: %v", err)
	}

	log.Printf("🎵 ShizuMusic [%s] is now online!", version.Info.ShizuMusic)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("✅ Bot Client:   READY")
	if client.UserClient != nil {
		log.Println("✅ User Client:  READY")
	} else {
		log.Println("⚠️  User Client:  NOT AVAILABLE")
	}
	log.Println("✅ Database:     CONNECTED")
	if globalCalls != nil {
		log.Println("✅ NTgCalls:     READY")
	} else {
		log.Println("⚠️  NTgCalls:     DISABLED")
	}
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Keep running
	for !isShuttingDown {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(1 * time.Second):
			// Heartbeat - check if still alive
		}
	}

	return nil
}

func shutdownHandler(ctx context.Context, cancel context.CancelFunc) {
	if isShuttingDown {
		return
	}
	isShuttingDown = true

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Println("🛑 Shutdown signal received. Stopping ShizuMusic...")
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Stop NTgCalls first (active voice chats)
	if globalCalls != nil {
		log.Println(">> Stopping NTgCalls...")
		globalCalls.Stop()
		log.Println("✅ NTgCalls stopped")
	}

	// Stop Telegram clients
	if globalClient != nil {
		log.Println(">> Disconnecting Telegram clients...")
		globalClient.Stop()
		log.Println("✅ Telegram clients disconnected")
	}

	// Close database connection
	if globalDB != nil {
		log.Println(">> Closing database connection...")
		globalDB.Close()
		log.Println("✅ Database connection closed")
	}

	// Send offline message
	if globalClient != nil && globalClient.BotClient != nil {
		offlineMsg := `#STOP

**ShizuMusic Bot is going offline**

**• Version:** ` + version.Info.ShizuMusic + `
**• Uptime:** ` + version.GetUptimeString()

		if err := globalClient.SendToLogger(offlineMsg, ""); err != nil {
			log.Printf("⚠️  Failed to send offline message: %v", err)
		}
	}

	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	log.Printf("👋 ShizuMusic [%s] is now offline!", version.Info.ShizuMusic)
	log.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Trigger context cancellation
	cancel()
}

func createDirectories(cfg *config.Config) error {
	dirs := []string{cfg.DwlDir, cfg.CacheDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	log.Printf("✅ Created directories: %v", dirs)
	return nil
}

func formatBootMessage() string {
	status := "✅ READY"
	if globalClient == nil || globalClient.UserClient == nil {
		status = "⚠️ LIMITED (No User Client)"
	}

	return `#START

**🎵 ShizuMusic Bot is now online!**

**System Information:**
• **Status:** ` + status + `
• **Version:** ` + version.Info.ShizuMusic + `
• **Go Version:** ` + version.Info.GoVersion + `
• **Gogram:** ` + version.Info.Gogram + `
• **NTgCalls:** ` + version.Info.NTgCalls + `
• **Uptime:** ` + version.GetUptimeString() + `

**Features:**
✅ Music Playback
✅ Queue Management
✅ Multi-platform Support
` + func() string {
		if globalClient != nil && globalClient.UserClient != nil {
			return "✅ Voice Chat Streaming"
		}
		return "⚠️ Voice Chat (Disabled - No User Client)"
	}()
}
