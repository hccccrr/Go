package core

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"strings"

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

	// Create bot client - NO Session field, Gogram auto-handles it
	client, err := tg.NewClient(tg.ClientConfig{
		AppID:    c.Config.APIID,
		AppHash:  c.Config.APIHash,
		LogLevel: tg.LogInfo,
		// NO Session field - Gogram creates "gogram.session" automatically
	})
	if err != nil {
		return fmt.Errorf("failed to create bot client: %w", err)
	}

	// Login as bot
	if err := client.LoginBot(c.Config.BotToken); err != nil {
		if strings.Contains(err.Error(), "ACCESS_TOKEN_EXPIRED") {
			return fmt.Errorf("bot token has been revoked or expired")
		}
		return fmt.Errorf("failed to login bot: %w", err)
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

// StartUser starts the user client with automatic session type detection
func (c *Client) StartUser(ctx context.Context) error {
	if c.Config.StringSession == "" {
		log.Println("⚠️  No STRING_SESSION provided, skipping user client")
		log.Println("   User client needed for voice chat streaming")
		return nil
	}

	log.Println(">> Booting up user client...")

	// Detect and convert session type
	stringSession, err := c.convertSession(c.Config.StringSession)
	if err != nil {
		return fmt.Errorf("failed to process session: %w", err)
	}

	// Create user client - ONLY StringSession, NO Session field!
	client, err := tg.NewClient(tg.ClientConfig{
		AppID:         c.Config.APIID,
		AppHash:       c.Config.APIHash,
		StringSession: stringSession,  // ONLY THIS - imports session directly!
		LogLevel:      tg.LogInfo,
		// NO Session field - this was causing "is a directory" error!
	})
	if err != nil {
		return fmt.Errorf("failed to create user client: %w", err)
	}

	// Connect (auto-authenticates with string session)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect user client: %w", err)
	}

	// Verify authentication
	me, err := client.GetMe()
	if err != nil {
		log.Printf("❌ Failed to authenticate user: %v", err)
		log.Println("⚠️  Your STRING_SESSION is invalid or expired")
		log.Println("   Generate new session:")
		log.Println("   pip3 install telethon --break-system-packages")
		log.Println("   python3 -c \"from telethon.sync import TelegramClient; from telethon.sessions import StringSession; c=TelegramClient(StringSession(),25742938,'b35b715fe8dc0a58e8048988286fc5b6'); c.start(); print(c.session.save())\"")
		return fmt.Errorf("failed to authenticate user (invalid STRING_SESSION): %w", err)
	}

	c.UserClient = client
	log.Printf(">> User @%s is online now!", me.Username)

	// Join support channels
	go c.joinChannels()

	return nil
}

// convertSession automatically detects and converts session format
func (c *Client) convertSession(session string) (string, error) {
	// Clean the session string
	session = strings.TrimSpace(session)
	
	// Telethon sessions start with "1"
	if strings.HasPrefix(session, "1") {
		log.Println("   📡 Detected: Telethon session format")
		sess, err := decodeTelethonSession(session)
		if err != nil {
			return "", fmt.Errorf("failed to decode Telethon session: %w", err)
		}
		log.Println("   ✅ Converted Telethon → Gogram")
		return sess.Encode(), nil
	}
	
	// Try to decode as Pyrogram (base64 format, 271 bytes)
	if len(session) > 100 && !strings.HasPrefix(session, "1") {
		sess, err := decodePyrogramSession(session)
		if err == nil {
			log.Println("   📡 Detected: Pyrogram session format")
			log.Println("   ✅ Converted Pyrogram → Gogram")
			return sess.Encode(), nil
		}
		// Not Pyrogram, continue
		log.Printf("   ⚠️  Not Pyrogram format: %v", err)
	}
	
	// Assume it's already in Gogram format
	log.Println("   📡 Using session as-is (Gogram native format)")
	return session, nil
}

// decodePyrogramSession decodes a Pyrogram session string
func decodePyrogramSession(encodedString string) (*tg.Session, error) {
	encodedString = strings.TrimSpace(encodedString)
	
	const (
		dcIDSize     = 1
		apiIDSize    = 4
		testModeSize = 1
		authKeySize  = 256
		userIDSize   = 8
		isBotSize    = 1
	)

	// Try standard base64 first (Pyrogram uses this)
	packedData, err := base64.StdEncoding.DecodeString(encodedString)
	if err != nil {
		// Try URL encoding as fallback
		for len(encodedString)%4 != 0 {
			encodedString += "="
		}
		packedData, err = base64.URLEncoding.DecodeString(encodedString)
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %w", err)
		}
	}

	expectedSize := dcIDSize + apiIDSize + testModeSize + authKeySize + userIDSize + isBotSize
	if len(packedData) != expectedSize {
		return nil, fmt.Errorf("unexpected data length: got %d, want %d", len(packedData), expectedSize)
	}

	dcID := int(uint8(packedData[0]))
	testMode := packedData[5] != 0
	authKey := make([]byte, authKeySize)
	copy(authKey, packedData[6:6+authKeySize])

	hostname := tg.ResolveDC(dcID, testMode, false)
	log.Printf("      → DC=%d, TestMode=%v", dcID, testMode)

	return &tg.Session{
		Hostname: hostname,
		Key:      authKey,
	}, nil
}

// decodeTelethonSession decodes a Telethon session string
func decodeTelethonSession(sessionString string) (*tg.Session, error) {
	if !strings.HasPrefix(sessionString, "1") {
		return nil, fmt.Errorf("invalid Telethon session: must start with '1'")
	}

	data, err := base64.URLEncoding.DecodeString(sessionString[1:])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	ipLen := 4
	if len(data) == 352 {
		ipLen = 16
	}

	expectedLen := 1 + ipLen + 2 + 256
	if len(data) != expectedLen {
		return nil, fmt.Errorf("invalid session length: got %d, want %d", len(data), expectedLen)
	}

	offset := 1
	ipData := data[offset : offset+ipLen]
	ip := net.IP(ipData)
	ipAddress := ip.String()
	offset += ipLen

	port := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	authKey := make([]byte, 256)
	copy(authKey, data[offset:offset+256])

	hostname := fmt.Sprintf("%s:%d", ipAddress, port)
	log.Printf("      → IP=%s, Port=%d", ipAddress, port)

	return &tg.Session{
		Hostname: hostname,
		Key:      authKey,
	}, nil
}

// joinChannels joins support channels
func (c *Client) joinChannels() {
	if c.UserClient == nil {
		return
	}

	channels := []string{"Its_HellBot"}
	
	for _, channel := range channels {
		if _, err := c.UserClient.JoinChannel(channel); err != nil {
			log.Printf("⚠️  Failed to join @%s: %v", channel, err)
		} else {
			log.Printf("✅ Joined @%s", channel)
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
