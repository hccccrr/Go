package core

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"shizumusic/config"

	tg "github.com/amarnathcjd/gogram/telegram"
)

// Client holds both bot and user clients
type Client struct {
	BotClient       *tg.Client
	UserClient      *tg.Client
	Config          *config.Config
	pluginsLoaded   bool  // Track if plugins are loaded
	handlersLoaded  bool  // Track if handlers are loaded
}

// NewClient creates a new client instance
func NewClient(cfg *config.Config) (*Client, error) {
	return &Client{
		Config:         cfg,
		pluginsLoaded:  false,
		handlersLoaded: false,
	}, nil
}

// StartBot starts the bot client
func (c *Client) StartBot(ctx context.Context) error {
	log.Println(">> Booting up bot client...")

	// Create bot client
	client, err := tg.NewClient(tg.ClientConfig{
		AppID:    c.Config.APIID,
		AppHash:  c.Config.APIHash,
		LogLevel: tg.LogInfo,
	})
	if err != nil {
		return fmt.Errorf("failed to create bot client: %w", err)
	}

	// Login as bot with proper error handling
	if err := client.LoginBot(c.Config.BotToken); err != nil {
		if strings.Contains(err.Error(), "ACCESS_TOKEN_EXPIRED") {
			log.Fatal("❌ Bot token has been revoked or expired.")
		} else {
			log.Fatal("❌ Failed to start the bot: " + err.Error())
		}
	}

	// Get bot info
	me, err := client.GetMe()
	if err != nil {
		return fmt.Errorf("failed to get bot info: %w", err)
	}

	c.BotClient = client
	log.Printf(">> Bot @%s is online now!", me.Username)

	return nil
}

// StartUser starts the assistant client
func (c *Client) StartUser(ctx context.Context) error {
	if c.Config.StringSession == "" {
		log.Fatal("❌ No STRING_SESSION provided for assistant client.")
	}

	log.Println(">> Starting assistant client...")

	// Create cache directory if not exists
	cacheDir := filepath.Join(".", "cache")
	sessionPath := filepath.Join(cacheDir, "assistant.session")

	// Create user client with string session
	client, err := tg.NewClient(tg.ClientConfig{
		AppID:         c.Config.APIID,
		AppHash:       c.Config.APIHash,
		Session:       sessionPath,
		LogLevel:      tg.LogInfo,
		StringSession: c.Config.StringSession,
	})
	if err != nil {
		log.Fatal("❌ Failed to create assistant client: " + err.Error())
	}

	// Connect
	if err := client.Connect(); err != nil {
		log.Fatal("❌ Failed to connect assistant client: " + err.Error())
	}

	// Get user info
	me, err := client.GetMe()
	if err != nil {
		log.Fatal("❌ Failed to get assistant info: " + err.Error())
	}

	c.UserClient = client
	log.Printf(">> Assistant @%s is online now!", me.Username)

	// Join support channels
	go c.joinChannels()

	return nil
}

// LoadHandlers loads all handlers automatically
// This is called from main.go after clients are started
func (c *Client) LoadHandlers(db *Database) error {
	if c.handlersLoaded {
		log.Println("⚠️  Handlers already loaded, skipping...")
		return nil
	}

	log.Println(">> Loading handlers...")

	// Import and load handlers package
	// This will automatically register all handlers
	if err := c.loadHandlerPlugins(db); err != nil {
		return fmt.Errorf("failed to load handlers: %w", err)
	}

	c.handlersLoaded = true
	log.Println("✅  All handlers loaded successfully!")
	
	return nil
}

// loadHandlerPlugins loads all handler modules
func (c *Client) loadHandlerPlugins(db *Database) error {
	// This function mimics Python's plugin loading
	// In Go, we'll explicitly import and register each handler
	
	log.Println("   → Loading bot handlers...")
	if err := c.registerBotHandlers(db); err != nil {
		return err
	}

	log.Println("   → Loading play handlers...")
	if err := c.registerPlayHandlers(db); err != nil {
		return err
	}

	log.Println("   → Loading control handlers...")
	if err := c.registerControlHandlers(db); err != nil {
		return err
	}

	log.Println("   → Loading admin handlers...")
	if err := c.registerAdminHandlers(db); err != nil {
		return err
	}

	log.Println("   → Loading callback handlers...")
	if err := c.registerCallbackHandlers(db); err != nil {
		return err
	}

	// Add more handler registrations here...

	return nil
}

// registerBotHandlers registers basic bot commands
func (c *Client) registerBotHandlers(db *Database) error {
	bot := c.BotClient

	// /start command
	bot.AddMessageHandler("/start", func(m *tg.NewMessage) error {
		if c.Config.IsBanned(m.From.ID) {
			return nil
		}

		text := fmt.Sprintf(`
🎵 **Welcome to ShizuMusic!**

Hello %s! I'm your music bot.

**Commands:**
/play - Play a song
/help - Show help
/ping - Check status

**Support:** @Its_HellBot
`, m.From.FirstName)

		_, err := m.Reply(text, &tg.SendOptions{ParseMode: "Markdown"})
		return err
	})

	// /help command
	bot.AddMessageHandler("/help", func(m *tg.NewMessage) error {
		if c.Config.IsBanned(m.From.ID) {
			return nil
		}

		text := `
📚 **ShizuMusic Help**

**Music:**
/play <song> - Play music
/pause - Pause playback
/resume - Resume
/skip - Skip song
/end - End playback

**Queue:**
/queue - Show queue
/shuffle - Shuffle queue

Send /start for more info!
`
		_, err := m.Reply(text, &tg.SendOptions{ParseMode: "Markdown"})
		return err
	})

	// /ping command
	bot.AddMessageHandler("/ping", func(m *tg.NewMessage) error {
		text := "🏓 Pong! Bot is alive!"
		_, err := m.Reply(text, nil)
		return err
	})

	return nil
}

// registerPlayHandlers registers play-related commands
func (c *Client) registerPlayHandlers(db *Database) error {
	bot := c.BotClient

	bot.AddMessageHandler("/play", func(m *tg.NewMessage) error {
		text := "🎵 Play command - Coming soon!"
		m.Reply(text, nil)
		return nil
	})

	return nil
}

// registerControlHandlers registers playback control commands
func (c *Client) registerControlHandlers(db *Database) error {
	bot := c.BotClient

	bot.AddMessageHandler("/pause", func(m *tg.NewMessage) error {
		text := "⏸️ Pause command - Coming soon!"
		m.Reply(text, nil)
		return nil
	})

	bot.AddMessageHandler("/resume", func(m *tg.NewMessage) error {
		text := "▶️ Resume command - Coming soon!"
		m.Reply(text, nil)
		return nil
	})

	bot.AddMessageHandler("/skip", func(m *tg.NewMessage) error {
		text := "⏭️ Skip command - Coming soon!"
		m.Reply(text, nil)
		return nil
	})

	return nil
}

// registerAdminHandlers registers admin commands
func (c *Client) registerAdminHandlers(db *Database) error {
	bot := c.BotClient

	bot.AddMessageHandler("/stats", func(m *tg.NewMessage) error {
		text := "📊 Stats command - Coming soon!"
		m.Reply(text, nil)
		return nil
	})

	return nil
}

// registerCallbackHandlers registers callback query handlers
func (c *Client) registerCallbackHandlers(db *Database) error {
	// Callback handlers will be registered here
	// Example:
	// bot.AddCallbackHandler(pattern, handler)
	
	return nil
}

// joinChannels joins support channels
func (c *Client) joinChannels() {
	channels := []string{"Its_HellBot"}
	
	for _, channel := range channels {
		if _, err := c.UserClient.JoinChannel(channel); err != nil {
			log.Printf("Warning: Failed to join @%s: %v", channel, err)
		}
	}
}

// SendToLogger sends message to logger channel
func (c *Client) SendToLogger(text string, photo string) error {
	if c.Config.LoggerID == 0 {
		return fmt.Errorf("logger ID not configured")
	}

	if photo != "" {
		_, err := c.BotClient.SendMedia(c.Config.LoggerID, photo, &tg.MediaOptions{
			Caption: text,
		})
		return err
	}

	_, err := c.BotClient.SendMessage(c.Config.LoggerID, text, nil)
	return err
}

// Stop gracefully stops both clients
func (c *Client) Stop() {
	if c.BotClient != nil {
		log.Println(">> Disconnecting bot client...")
		c.BotClient.Stop()
	}

	if c.UserClient != nil {
		log.Println(">> Disconnecting assistant client...")
		c.UserClient.Stop()
	}
}

// IsBot checks if message is from bot
func (c *Client) IsBot(userID int64) bool {
	if c.BotClient == nil {
		return false
	}
	me, _ := c.BotClient.GetMe()
	return me != nil && me.ID == userID
}
