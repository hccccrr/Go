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

	tg "github.com/amarnathcjd/gogram/telegram"
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
	fmt.Println(">> Starting assistant client...")
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
	// REGISTER HANDLERS HERE!
	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

	bot := client.BotClient

	// /start command
	bot.OnNewMessage(tg.OnNewMessage{Pattern: "^/start"}, func(m *tg.NewMessage) error {
		text := fmt.Sprintf(`
🎵 **Welcome to ShizuMusic!**

Hello %s! I'm alive and ready to play music!

**Version:** %s
**Status:** Online ✅

**Quick Commands:**
/help - Show all commands
/play - Play a song
/ping - Check bot status

**Support:** @Its_HellBot
`, m.Sender.FirstName, version.Version)

		_, err := m.Reply(text, &tg.SendOptions{ParseMode: "Markdown"})
		return err
	})

	// /help command
	bot.OnNewMessage(tg.OnNewMessage{Pattern: "^/help"}, func(m *tg.NewMessage) error {
		text := `
📚 **ShizuMusic Help**

**Music Commands:**
/play <song> - Play a song
/pause - Pause playback
/resume - Resume playback
/skip - Skip current song
/end - End playback

**Queue:**
/queue - Show queue
/shuffle - Shuffle queue

**Info:**
/ping - Check bot status
/stats - Bot statistics

More commands coming soon!
`
		_, err := m.Reply(text, &tg.SendOptions{ParseMode: "Markdown"})
		return err
	})

	// /ping command
	bot.OnNewMessage(tg.OnNewMessage{Pattern: "^/ping"}, func(m *tg.NewMessage) error {
		start := time.Now()
		msg, _ := m.Reply("⏳ Pinging...", nil)
		elapsed := time.Since(start).Milliseconds()

		uptime := time.Since(cfg.StartTime)
		hours := int(uptime.Hours())
		minutes := int(uptime.Minutes()) % 60

		text := fmt.Sprintf(`
🏓 **Pong!**

**Response Time:** %dms
**Uptime:** %dh %dm
**NTgCalls:** %dms
**Status:** Online ✅

**Version:** %s
`, elapsed, hours, minutes, calls.GetPing(), version.Version)

		msg.Edit(text, &tg.SendOptions{ParseMode: "Markdown"})
		return nil
	})

	// /stats command
	bot.OnNewMessage(tg.OnNewMessage{Pattern: "^/stats"}, func(m *tg.NewMessage) error {
		totalUsers, _ := db.TotalUsersCount()
		totalSongs, _ := db.TotalSongsCount()
		activeVCs := db.GetActiveVC()

		text := fmt.Sprintf(`
📊 **Bot Statistics**

**Users:** %d
**Songs Played:** %d
**Active VCs:** %d
**Version:** %s

**Uptime:** %s
**Status:** Online ✅
`, totalUsers, totalSongs, len(activeVCs), version.Version, 
		time.Since(cfg.StartTime).Round(time.Second))

		_, err := m.Reply(text, &tg.SendOptions{ParseMode: "Markdown"})
		return err
	})

	// Fallback for unknown commands
	bot.OnNewMessage(tg.OnNewMessage{}, func(m *tg.NewMessage) error {
		// Only respond to commands
		if len(m.Text()) > 0 && m.Text()[0] == '/' {
			text := "❌ Unknown command! Send /help for available commands."
			m.Reply(text, nil)
		}
		return nil
	})

	fmt.Println("✅  Handlers registered!")

	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
	// SEND BOOT MESSAGE
	// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

	if cfg.LoggerID != 0 {
		botMe, _ := bot.GetMe()
		userMe, _ := client.UserClient.GetMe()

		bootMsg := fmt.Sprintf(`
🎵 **ShizuMusic Started!**

✅ **Version:** %s
✅ **Bot:** @%s
✅ **Assistant:** @%s
✅ **Database:** Connected
✅ **NTgCalls:** Ready

**Status:** Bot is now online! ✅
**Time:** %s

Send /start to test!
`, version.Version, botMe.Username, userMe.Username, time.Now().Format("15:04:05"))

		_, err := bot.SendMessage(cfg.LoggerID, bootMsg, &tg.SendOptions{ParseMode: "Markdown"})
		if err != nil {
			fmt.Printf("⚠️  Failed to send boot message: %v\n", err)
		} else {
			fmt.Println("✅  Boot message sent to logger!")
		}
	}

	// Print status
	fmt.Printf("\n🎵 ShizuMusic [%s] is now online!\n", version.Version)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅  Bot Client:   READY")
	fmt.Println("✅  User Client:  READY")
	fmt.Println("✅  Database:     CONNECTED")
	fmt.Println("✅  NTgCalls:     READY")
	fmt.Println("✅  Handlers:     REGISTERED")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("\n📝  Bot is ready! Test with /start")
	fmt.Println("⏸️   Press Ctrl+C to stop\n")

	// Idle - wait for signals
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
