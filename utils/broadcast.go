package utils

import (
	"context"
	"fmt"
	"os"
	"time"
)

// TelegramError types
type TelegramError struct {
	Type    string
	Seconds int
	Message string
}

func (e *TelegramError) Error() string {
	return e.Message
}

// Error type constants
const (
	ErrFloodWait         = "FLOOD_WAIT"
	ErrUserDeactivated   = "USER_DEACTIVATED"
	ErrUserBlocked       = "USER_BLOCKED"
	ErrPeerIDInvalid     = "PEER_ID_INVALID"
	ErrGeneric           = "GENERIC"
)

// BroadcastTarget represents a broadcast recipient
type BroadcastTarget struct {
	UserID int64
	ChatID int64
}

// MessageSender interface for sending messages (to be implemented by Telegram client)
type MessageSender interface {
	ForwardMessage(ctx context.Context, targetID int64, message interface{}) error
	SendMessage(ctx context.Context, targetID int64, message interface{}) error
}

// Database interface for getting targets
type Database interface {
	GetAllChats(ctx context.Context) ([]BroadcastTarget, error)
	GetAllUsers(ctx context.Context) ([]BroadcastTarget, error)
	TotalChatsCount(ctx context.Context) (int, error)
	TotalUsersCount(ctx context.Context) (int, error)
}

// Broadcast handles message broadcasting
type Broadcast struct {
	fileName string
	sender   MessageSender
	db       Database
}

// NewBroadcast creates a new Broadcast instance
func NewBroadcast(sender MessageSender, db Database) *Broadcast {
	return &Broadcast{
		fileName: "broadcast_%d.txt",
		sender:   sender,
		db:       db,
	}
}

// SendMsgResult represents the result of sending a message
type SendMsgResult struct {
	StatusCode int
	ErrorMsg   string
}

// SendMsg sends a message to a user (forward or copy)
func (b *Broadcast) SendMsg(ctx context.Context, userID int64, message interface{}, copy bool) (SendMsgResult, error) {
	var err error

	if !copy {
		err = b.sender.ForwardMessage(ctx, userID, message)
	} else {
		err = b.sender.SendMessage(ctx, userID, message)
	}

	if err != nil {
		return b.handleError(userID, err)
	}

	return SendMsgResult{StatusCode: 200, ErrorMsg: ""}, nil
}

// handleError processes Telegram errors and returns appropriate result
func (b *Broadcast) handleError(userID int64, err error) (SendMsgResult, error) {
	if telegramErr, ok := err.(*TelegramError); ok {
		switch telegramErr.Type {
		case ErrFloodWait:
			// Sleep and retry
			time.Sleep(time.Duration(telegramErr.Seconds) * time.Second)
			return SendMsgResult{StatusCode: 429, ErrorMsg: ""}, err

		case ErrUserDeactivated:
			return SendMsgResult{
				StatusCode: 400,
				ErrorMsg:   fmt.Sprintf("%d -:- deactivated\n", userID),
			}, nil

		case ErrUserBlocked:
			return SendMsgResult{
				StatusCode: 400,
				ErrorMsg:   fmt.Sprintf("%d -:- blocked the bot\n", userID),
			}, nil

		case ErrPeerIDInvalid:
			return SendMsgResult{
				StatusCode: 400,
				ErrorMsg:   fmt.Sprintf("%d -:- user id invalid\n", userID),
			}, nil
		}
	}

	// Generic error
	return SendMsgResult{
		StatusCode: 500,
		ErrorMsg:   fmt.Sprintf("%d -:- %v\n", userID, err),
	}, nil
}

// BroadcastStats represents broadcast statistics
type BroadcastStats struct {
	TotalTargets int
	Done         int
	Success      int
	Failed       int
	CompletedIn  time.Duration
	LogFile      string
}

// BroadcastOptions contains broadcast configuration
type BroadcastOptions struct {
	Type    string // "chats", "users", "all"
	Copy    bool
	Message interface{}
}

// BroadcastMessage broadcasts a message to users/chats
func (b *Broadcast) BroadcastMessage(ctx context.Context, opts BroadcastOptions) (*BroadcastStats, error) {
	var targets [][]BroadcastTarget
	var count int

	// Get targets based on type
	switch opts.Type {
	case "chats":
		chats, err := b.db.GetAllChats(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get chats: %w", err)
		}
		targets = append(targets, chats)
		count, _ = b.db.TotalChatsCount(ctx)

	case "users":
		users, err := b.db.GetAllUsers(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get users: %w", err)
		}
		targets = append(targets, users)
		count, _ = b.db.TotalUsersCount(ctx)

	case "all":
		users, err := b.db.GetAllUsers(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get users: %w", err)
		}
		chats, err := b.db.GetAllChats(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get chats: %w", err)
		}
		targets = append(targets, users, chats)
		
		userCount, _ := b.db.TotalUsersCount(ctx)
		chatCount, _ := b.db.TotalChatsCount(ctx)
		count = userCount + chatCount

	default:
		return nil, fmt.Errorf("invalid broadcast type: %s", opts.Type)
	}

	// Create log file
	fileName := fmt.Sprintf(b.fileName, time.Now().Unix())
	logFile, err := os.Create(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}
	defer logFile.Close()

	// Track statistics
	stats := &BroadcastStats{
		TotalTargets: count,
		LogFile:      fileName,
	}

	startTime := time.Now()

	// Send messages
	for _, targetList := range targets {
		for _, target := range targetList {
			// Determine target ID
			var targetID int64
			if target.UserID != 0 {
				targetID = target.UserID
			} else {
				targetID = target.ChatID
			}

			// Send message
			result, err := b.SendMsg(ctx, targetID, opts.Message, opts.Copy)
			
			// Handle flood wait - retry
			if err != nil && result.StatusCode == 429 {
				result, _ = b.SendMsg(ctx, targetID, opts.Message, opts.Copy)
			}

			// Write error to log if any
			if result.ErrorMsg != "" {
				logFile.WriteString(result.ErrorMsg)
			}

			// Update statistics
			if result.StatusCode == 200 {
				stats.Success++
			} else {
				stats.Failed++
			}
			stats.Done++

			// Check context cancellation
			select {
			case <-ctx.Done():
				stats.CompletedIn = time.Since(startTime)
				return stats, ctx.Err()
			default:
			}
		}
	}

	stats.CompletedIn = time.Since(startTime)
	return stats, nil
}

// FormatBroadcastResult formats broadcast statistics into a message
func FormatBroadcastResult(stats *BroadcastStats, pasteLink string) string {
	result := fmt.Sprintf(
		"__Broadcast completed successfully!__\n\n"+
			"**Chats in DB:** `%d chats` \n"+
			"**Gcast Iterations:** `%d loops` \n"+
			"**Gcasted in:** `%d chats` \n"+
			"**Failed in:** `%d chats` \n"+
			"**Completed in:** `%s`",
		stats.TotalTargets,
		stats.Done,
		stats.Success,
		stats.Failed,
		stats.CompletedIn,
	)

	if stats.Failed > 0 && pasteLink != "" {
		result += fmt.Sprintf("\n\n**Error log:** [here](%s)", pasteLink)
	}

	return result
}

// CleanupLogFile removes the log file
func CleanupLogFile(fileName string) error {
	return os.Remove(fileName)
}
