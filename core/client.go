package core

import (
	"context"
	"fmt"
	"log"

	"shizumusic/config"

	tg "github.com/amarnathcjd/gogram/telegram"
)

// Client holds both bot and user clients
type Client struct {
	BotClient  *tg.Client
	UserClient *tg.Client
	Config     *config.Config
}

// NewClient creates a new client instance
func NewClient(cfg *config.Config) (*Client, error) {
	return &Client{
		Config: cfg,
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

	// Start as bot
	if err := client.Start(c.Config.BotToken); err != nil {
		return fmt.Errorf("failed to start bot: %w", err)
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

// StartUser starts the user client
func (c *Client) StartUser(ctx context.Context) error {
	if c.Config.StringSession == "" {
		log.Println("⚠️ No user session provided, skipping user client")
		return nil
	}

	log.Println(">> Booting up user client...")

	// Create user client with string session
	client, err := tg.NewClient(tg.ClientConfig{
		AppID:         c.Config.APIID,
		AppHash:       c.Config.APIHash,
		Session:       c.Config.StringSession,
		LogLevel:      tg.LogInfo,
		StringSession: c.Config.StringSession, // Pass the actual session string
	})
	if err != nil {
		return fmt.Errorf("failed to create user client: %w", err)
	}

	// Connect
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect user client: %w", err)
	}

	// Get user info
	me, err := client.GetMe()
	if err != nil {
		return fmt.Errorf("failed to get user info: %w", err)
	}

	c.UserClient = client
	log.Printf(">> User @%s is online now!", me.Username)

	// Join support channels
	go c.joinChannels()

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
		// Send with photo - removed File field
		_, err := c.BotClient.SendMedia(c.Config.LoggerID, photo, &tg.MediaOptions{
			Caption: text,
		})
		return err
	}

	// Send text only
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
		log.Println(">> Disconnecting user client...")
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
