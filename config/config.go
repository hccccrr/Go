package config

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration
type Config struct {
	// Required
	APIHash       string
	APIID         int32
	BotToken      string
	DatabaseURL   string
	StringSession string
	LoggerID      int64
	OwnerID       int64
	StartTime     time.Time

	// Optional
	BlackImg        string
	BotName         string
	BotPic          string
	LeaderboardTime string
	LyricsAPI       string
	MaxFavorites    int
	PlayLimit       int
	PrivateMode     bool
	SongLimit       int
	TelegramImg     string
	TGAudioLimit    int64
	TGVideoLimit    int64
	TZ              string
	UpstreamRepo    string
	UpstreamBranch  string
	GitToken        string

	// Runtime (thread-safe maps)
	BannedUsers map[int64]bool
	SudoUsers   map[int64]bool
	GodUsers    map[int64]bool
	Cache       map[string]interface{}
	PlayerCache map[int64]interface{}
	QueueCache  map[int64]interface{}
	SongCache   map[string]interface{}
	CacheDir    string
	DwlDir      string
	DeleteDict  map[string]interface{}

	// Mutexes for thread safety
	BannedMutex sync.RWMutex
	SudoMutex   sync.RWMutex
	GodMutex    sync.RWMutex
	CacheMutex  sync.RWMutex
}

var Cfg *Config

// Load loads configuration from environment
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	apiID, _ := strconv.Atoi(getEnv("API_ID", "0"))
	ownerID, _ := strconv.ParseInt(getEnv("OWNER_ID", "0"), 10, 64)

	// Logger ID with proper parsing
	var loggerID int64
	loggerIDStr := getEnv("LOGGER_ID", "")
	if loggerIDStr != "" {
		loggerID, _ = strconv.ParseInt(loggerIDStr, 10, 64)
	}
	if loggerID == 0 {
		loggerID = -1003303257249
	}

	// Hardcoded string session
	stringSession := "1BvEeyJrZXkiOiJ0bC9DSFBWK2t2cFpGYzlZZEUwN0RLaGpxTXNSUEEya2FpMTEwZEYrMXBuVmloeUNkbC84aVBwTFlSVXY2MXlmcHFDZ2ZXNTljMW83UGV6T3R5MjFza1lVcWtvWGw0Y0loSnVId0c2ZzVDc0M2MWV4ZHFwTjZSdWlPc2xEM2FNQmY4NGtEVzdzQjlJTE9WelRNc3FrSzV4U1IvbktmUmFvTDVDMTRsY2V4N1RNOVM2M3BmRnU0T0x4M0x5T1lDQnpGQmFQUHY4UUh1UTkrZjVGdUlod2s2VWs3dDJPbEVGOWdqeHh3ODFZWXRReVpxUjJFRzBSd1BrdU03eXVVd2JjbDRhaXlpaXRoTEpzTFB5Z0JTaE1UZTBXME5JZThIZHcxajJBdTNIUi9yOUZDdGoxaXB2RXFGbExuOVQrTzlSN2RHU292ZVdDZUhHOVQ3dmtUT1g3V1E9PSIsImhhc2giOiJmOUtFUGc5Q2ZaWT0iLCJkY19pZCI6MiwiaXBfYWRkciI6IjE0OS4xNTQuMTY3LjUwOjQ0MyIsImFwcF9pZCI6MzE2MzcwNjR9"

	cfg := &Config{
		// Required
		APIHash:       getEnv("API_HASH", ""),
		APIID:         int32(apiID),
		BotToken:      getEnv("BOT_TOKEN", ""),
		DatabaseURL:   getEnv("DATABASE_URL", ""),
		StringSession: stringSession,
		LoggerID:      loggerID,
		OwnerID:       ownerID,
		StartTime:     time.Now(),

		// Optional
		BlackImg:        getEnv("BLACK_IMG", "https://telegra.ph/file/2c546060b20dfd7c1ff2d.jpg"),
		BotName:         getEnv("BOT_NAME", "ShizuMusic"),
		BotPic:          getEnv("BOT_PIC", "https://files.catbox.moe/jwkw46.jpg"),
		LeaderboardTime: getEnv("LEADERBOARD_TIME", "3:00"),
		LyricsAPI:       getEnv("LYRICS_API", ""),
		MaxFavorites:    getEnvInt("MAX_FAVORITES", 30),
		PlayLimit:       getEnvInt("PLAY_LIMIT", 0),
		PrivateMode:     getEnv("PRIVATE_MODE", "off") == "on",
		SongLimit:       getEnvInt("SONG_LIMIT", 0),
		TelegramImg:     getEnv("TELEGRAM_IMG", ""),
		TGAudioLimit:    int64(getEnvInt("TG_AUDIO_SIZE_LIMIT", 104857600)),
		TGVideoLimit:    int64(getEnvInt("TG_VIDEO_SIZE_LIMIT", 1073741824)),
		TZ:              getEnv("TZ", "Asia/Kolkata"),
		UpstreamRepo:    getEnv("UPSTREAM_REPO", "https://github.com/hccccrr/Go"),
		UpstreamBranch:  getEnv("UPSTREAM_BRANCH", "main"),
		GitToken:        getEnv("GIT_TOKEN", ""),

		// Directories
		CacheDir: "./cache/",
		DwlDir:   "./downloads/",

		// Initialize runtime maps
		BannedUsers: make(map[int64]bool),
		SudoUsers:   make(map[int64]bool),
		GodUsers:    make(map[int64]bool),
		Cache:       make(map[string]interface{}),
		PlayerCache: make(map[int64]interface{}),
		QueueCache:  make(map[int64]interface{}),
		SongCache:   make(map[string]interface{}),
		DeleteDict:  make(map[string]interface{}),
	}

	if cfg.OwnerID != 0 {
		cfg.GodUsers[cfg.OwnerID] = true
		cfg.SudoUsers[cfg.OwnerID] = true
	}

	Cfg = cfg
	return cfg, nil
}

// CloseLogging placeholder
func CloseLogging() {}

// Validate checks required fields
func (c *Config) Validate() error {
	if c.APIID == 0 {
		log.Fatal("❌ API_ID is missing!")
	}
	if c.APIHash == "" {
		log.Fatal("❌ API_HASH is missing!")
	}
	if c.BotToken == "" {
		log.Fatal("❌ BOT_TOKEN is missing!")
	}
	if c.DatabaseURL == "" {
		log.Fatal("❌ DATABASE_URL is missing!")
	}
	if c.StringSession == "" {
		log.Fatal("❌ STRING_SESSION is missing!")
	}
	if c.LoggerID == 0 {
		log.Fatal("❌ LOGGER_ID is missing!")
	}
	if c.OwnerID == 0 {
		log.Fatal("❌ OWNER_ID is missing!")
	}
	log.Println("✅ Config validation passed!")
	return nil
}

// IsBanned checks if user is banned (thread-safe)
func (c *Config) IsBanned(userID int64) bool {
	c.BannedMutex.RLock()
	defer c.BannedMutex.RUnlock()
	return c.BannedUsers[userID]
}

// AddBanned adds user to banned list (thread-safe)
func (c *Config) AddBanned(userID int64) {
	c.BannedMutex.Lock()
	defer c.BannedMutex.Unlock()
	c.BannedUsers[userID] = true
}

// RemoveBanned removes user from banned list (thread-safe)
func (c *Config) RemoveBanned(userID int64) {
	c.BannedMutex.Lock()
	defer c.BannedMutex.Unlock()
	delete(c.BannedUsers, userID)
}

// IsSudo checks if user is sudo (thread-safe)
func (c *Config) IsSudo(userID int64) bool {
	c.SudoMutex.RLock()
	defer c.SudoMutex.RUnlock()
	return c.SudoUsers[userID]
}

// AddSudo adds user to sudo list (thread-safe)
func (c *Config) AddSudo(userID int64) {
	c.SudoMutex.Lock()
	defer c.SudoMutex.Unlock()
	c.SudoUsers[userID] = true
}

// IsGod checks if user is god/owner (thread-safe)
func (c *Config) IsGod(userID int64) bool {
	c.GodMutex.RLock()
	defer c.GodMutex.RUnlock()
	return c.GodUsers[userID]
}

// Helper functions
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}
