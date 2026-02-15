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

	// Create bot client with session file
	client, err := tg.NewClient(tg.ClientConfig{
		AppID:    c.Config.APIID,
		AppHash:  c.Config.APIHash,
		LogLevel: tg.LogInfo,
		Session:  "bot.session", // Bot session file
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

	// Create user client with converted session
	client, err := tg.NewClient(tg.ClientConfig{
		AppID:         c.Config.APIID,
		AppHash:       c.Config.APIHash,
		StringSession: stringSession,
		Session:       "user.session",
		LogLevel:      tg.LogInfo,
	})
	if err != nil {
		return fmt.Errorf("failed to create user client: %w", err)
	}

	// Connect
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect user client: %w", err)
	}

	// Verify authentication
	me, err := client.GetMe()
	if err != nil {
		log.Printf("❌ Failed to authenticate user: %v", err)
		log.Println("⚠️  Your STRING_SESSION might be invalid or expired")
		log.Println("   Generate new session using: ./session-gen")
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
		log.Println("   Detected: Telethon session format")
		sess, err := decodeTelethonSession(session)
		if err != nil {
			return "", fmt.Errorf("failed to decode Telethon session: %w", err)
		}
		return sess.Encode(), nil
	}
	
	// Try to decode as Pyrogram
	// Pyrogram sessions are base64 encoded and decode to 271 bytes
	if len(session) > 100 {
		sess, err := decodePyrogramSession(session)
		if err == nil {
			log.Println("   Detected: Pyrogram session format")
			return sess.Encode(), nil
		}
		// If decoding failed, log the error for debugging
		log.Printf("   Pyrogram decode attempt failed: %v", err)
	}
	
	// Assume it's already in Gogram format
	log.Println("   Using session as-is (Gogram format)")
	return session, nil
}

// decodePyrogramSession decodes a Pyrogram session string
func decodePyrogramSession(encodedString string) (*tg.Session, error) {
	// Clean the string
	encodedString = strings.TrimSpace(encodedString)
	
	// Pyrogram SESSION_STRING_FORMAT: 
	// Big-endian, uint8, uint32, bool, 256-byte array, uint64, bool
	const (
		dcIDSize     = 1   // uint8
		apiIDSize    = 4   // uint32
		testModeSize = 1   // bool
		authKeySize  = 256
		userIDSize   = 8   // uint64
		isBotSize    = 1   // bool
	)

	// Decode base64 (Pyrogram uses standard base64, not URL encoding)
	// First try standard base64
	packedData, err := base64.StdEncoding.DecodeString(encodedString)
	if err != nil {
		// Try URL encoding as fallback
		// Add padding if needed
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

	// Extract DC ID
	dcID := int(uint8(packedData[0]))
	
	// Extract test mode (at offset 5: after dcID(1) + apiID(4))
	testMode := packedData[5] != 0

	// Extract auth key (starts at offset 6)
	authKey := make([]byte, authKeySize)
	copy(authKey, packedData[6:6+authKeySize])

	// Resolve DC hostname
	hostname := tg.ResolveDC(dcID, testMode, false)
	
	log.Printf("   Pyrogram session decoded: DC=%d, TestMode=%v, Hostname=%s", dcID, testMode, hostname)

	return &tg.Session{
		Hostname: hostname,
		Key:      authKey,
	}, nil
}

// decodeTelethonSession decodes a Telethon session string
func decodeTelethonSession(sessionString string) (*tg.Session, error) {
	// Remove "1" prefix
	if !strings.HasPrefix(sessionString, "1") {
		return nil, fmt.Errorf("invalid Telethon session: must start with '1'")
	}

	// Decode the rest (URL-safe base64)
	data, err := base64.URLEncoding.DecodeString(sessionString[1:])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Determine IP length (IPv4 or IPv6)
	ipLen := 4
	if len(data) == 352 {
		ipLen = 16
	}

	expectedLen := 1 + ipLen + 2 + 256
	if len(data) != expectedLen {
		return nil, fmt.Errorf("invalid session string length: got %d, want %d", len(data), expectedLen)
	}

	offset := 1

	// Extract IP address
	ipData := data[offset : offset+ipLen]
	ip := net.IP(ipData)
	ipAddress := ip.String()
	offset += ipLen

	// Extract port (Big Endian)
	port := binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2

	// Extract auth key
	authKey := make([]byte, 256)
	copy(authKey, data[offset:offset+256])

	hostname := fmt.Sprintf("%s:%d", ipAddress, port)
	log.Printf("   Telethon session decoded: Hostname=%s", hostname)

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
