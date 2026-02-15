package utils

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// UserStats represents user statistics
type UserStats struct {
	ID       int64
	Username string
	Songs    int
	Messages int
}

// LeaderboardDatabase interface for database operations
type LeaderboardDatabase interface {
	GetAllUsers(ctx context.Context) ([]UserStats, error)
	GetAllChats(ctx context.Context) ([]int64, error)
}

// MessageSenderForLeaderboard interface for sending messages
type MessageSenderForLeaderboard interface {
	SendMessage(ctx context.Context, chatID int64, text string, buttons interface{}) error
}

// Leaderboard manages leaderboard generation and broadcasting
type Leaderboard struct {
	fileName string
	db       LeaderboardDatabase
	sender   MessageSenderForLeaderboard
}

// NewLeaderboard creates a new Leaderboard instance
func NewLeaderboard(db LeaderboardDatabase, sender MessageSenderForLeaderboard) *Leaderboard {
	return &Leaderboard{
		fileName: "leaderboard.txt",
		db:       db,
		sender:   sender,
	}
}

// BotDetails contains bot information for leaderboard
type BotDetails struct {
	Username string
	Mention  string
}

// GetHours parses hours from leaderboard time config (format: "HH:MM")
func (l *Leaderboard) GetHours(configTime string) int {
	parts := strings.Split(configTime, ":")
	if len(parts) > 0 {
		hrs, err := strconv.Atoi(parts[0])
		if err == nil {
			return hrs
		}
	}
	return 3 // Default
}

// GetMinutes parses minutes from leaderboard time config (format: "HH:MM")
func (l *Leaderboard) GetMinutes(configTime string) int {
	parts := strings.Split(configTime, ":")
	if len(parts) > 1 {
		mins, err := strconv.Atoi(parts[1])
		if err == nil {
			return mins
		}
	}
	return 0 // Default
}

// GetTop10Songs returns top 10 users by songs played
func (l *Leaderboard) GetTop10Songs(ctx context.Context) ([]UserStats, error) {
	users, err := l.db.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	// Sort by songs played (descending)
	sort.Slice(users, func(i, j int) bool {
		return users[i].Songs > users[j].Songs
	})

	// Return top 10
	if len(users) > 10 {
		return users[:10], nil
	}
	return users, nil
}

// GetTop10Messages returns top 10 users by message count
func (l *Leaderboard) GetTop10Messages(ctx context.Context) ([]UserStats, error) {
	users, err := l.db.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	// Sort by messages (descending)
	sort.Slice(users, func(i, j int) bool {
		return users[i].Messages > users[j].Messages
	})

	// Return top 10
	if len(users) > 10 {
		return users[:10], nil
	}
	return users, nil
}

// GenerateSongs generates leaderboard text for songs
func (l *Leaderboard) GenerateSongs(ctx context.Context, botDetails BotDetails) (string, error) {
	top10, err := l.GetTop10Songs(ctx)
	if err != nil {
		return "", err
	}

	text := fmt.Sprintf("**🎵 Top 10 Music Lovers of %s**\n\n", botDetails.Mention)

	for i, user := range top10 {
		index := i + 1
		link := fmt.Sprintf("https://t.me/%s?start=user_%d", botDetails.Username, user.ID)
		indexStr := fmt.Sprintf("%02d", index)
		text += fmt.Sprintf("**‣ %s:** [%s](%s) - **%d** songs\n", indexStr, user.Username, link, user.Songs)
	}

	text += "\n**🎧 Keep streaming! Enjoy the music!**"
	return text, nil
}

// GenerateMessages generates leaderboard text for messages
func (l *Leaderboard) GenerateMessages(ctx context.Context, botDetails BotDetails) (string, error) {
	top10, err := l.GetTop10Messages(ctx)
	if err != nil {
		return "", err
	}

	text := fmt.Sprintf("**💬 Top 10 Active Chatters of %s**\n\n", botDetails.Mention)

	for i, user := range top10 {
		index := i + 1
		link := fmt.Sprintf("https://t.me/%s?start=user_%d", botDetails.Username, user.ID)
		indexStr := fmt.Sprintf("%02d", index)
		text += fmt.Sprintf("**‣ %s:** [%s](%s) - **%d** messages\n", indexStr, user.Username, link, user.Messages)
	}

	text += "\n**💬 Keep chatting! Stay active!**"
	return text, nil
}

// Generate generates leaderboard text
// leaderboardType: "songs", "messages", or "both"
func (l *Leaderboard) Generate(ctx context.Context, botDetails BotDetails, leaderboardType string) (string, error) {
	switch leaderboardType {
	case "songs":
		return l.GenerateSongs(ctx, botDetails)
	case "messages":
		return l.GenerateMessages(ctx, botDetails)
	case "both":
		songsText, err := l.GenerateSongs(ctx, botDetails)
		if err != nil {
			return "", err
		}
		messagesText, err := l.GenerateMessages(ctx, botDetails)
		if err != nil {
			return "", err
		}
		separator := "\n\n" + strings.Repeat("─", 35) + "\n\n"
		return songsText + separator + messagesText, nil
	default:
		return "", fmt.Errorf("invalid leaderboard type: %s", leaderboardType)
	}
}

// BroadcastResult represents broadcast statistics
type BroadcastResult struct {
	Success   int
	Failed    int
	Total     int
	TimeTaken time.Duration
}

// Broadcast sends leaderboard to all chats
func (l *Leaderboard) Broadcast(ctx context.Context, text string, buttons interface{}) (*BroadcastResult, error) {
	startTime := time.Now()
	result := &BroadcastResult{}

	// Get all chats
	chats, err := l.db.GetAllChats(ctx)
	if err != nil {
		return result, err
	}

	result.Total = len(chats)

	// Create log file
	logFile, err := os.Create(l.fileName)
	if err != nil {
		return result, err
	}
	defer logFile.Close()

	// Send to each chat
	for _, chatID := range chats {
		success, errMsg := l.sendMessage(ctx, chatID, text, buttons)
		
		if errMsg != "" {
			logFile.WriteString(errMsg)
		}

		if success {
			result.Success++
		} else {
			result.Failed++
		}

		// Small delay to avoid flood
		time.Sleep(300 * time.Millisecond)

		// Check context cancellation
		select {
		case <-ctx.Done():
			result.TimeTaken = time.Since(startTime)
			return result, ctx.Err()
		default:
		}
	}

	result.TimeTaken = time.Since(startTime)
	return result, nil
}

// sendMessage sends message to a single chat with error handling
func (l *Leaderboard) sendMessage(ctx context.Context, chatID int64, text string, buttons interface{}) (bool, string) {
	err := l.sender.SendMessage(ctx, chatID, text, buttons)
	if err != nil {
		errMsg := fmt.Sprintf("%d -:- %v\n", chatID, err)
		return false, errMsg
	}
	return true, ""
}

// FormatBroadcastResult formats broadcast result into a message
func (l *Leaderboard) FormatBroadcastResult(result *BroadcastResult) string {
	return fmt.Sprintf(
		"**Leaderboard Auto Broadcast Completed in** `%s`\n\n"+
			"**Total Chats:** `%d`\n"+
			"**Success:** `%d`\n"+
			"**Failed:** `%d`\n\n"+
			"**🧡 Enjoy Streaming! Have Fun!**",
		result.TimeTaken,
		result.Total,
		result.Success,
		result.Failed,
	)
}

// GetLogFileName returns the log file name
func (l *Leaderboard) GetLogFileName() string {
	return l.fileName
}

// CleanupLogFile removes the log file
func (l *Leaderboard) CleanupLogFile() error {
	return os.Remove(l.fileName)
}

// SetLogFileName sets custom log file name
func (l *Leaderboard) SetLogFileName(fileName string) {
	l.fileName = fileName
}
