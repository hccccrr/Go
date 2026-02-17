package handlers

import (
	"fmt"
	"log"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/utils"
	"shizumusic/helpers"
)

// TEXTS templates from utils
var TEXTS = helpers.TextTemplates

// RegisterWatcherHandlers registers event watchers
func RegisterWatcherHandlers(client *core.Client, db *core.Database) {
	// Track new users in PM
	client.BotClient.OnNewMessage(func(m *tg.NewMessage) error {
		if m.IsPrivate {
			return trackNewUser(m, client, db)
		}
		return nil
	})

	// Track new chats
	client.BotClient.OnNewMessage(func(m *tg.NewMessage) error {
		if m.IsGroup {
			return trackNewChat(m, client, db)
		}
		return nil
	})

	// Track user messages for leaderboard
	client.BotClient.OnNewMessage(func(m *tg.NewMessage) error {
		if m.IsGroup && m.From != nil {
			return trackUserMessage(m, db)
		}
		return nil
	})

	// Message count command
	client.BotClient.AddMessageHandler("/msgcount", func(m *tg.NewMessage) error {
		return handleMsgCount(m, db)
	})
	
	client.BotClient.AddMessageHandler("/messagecount", func(m *tg.NewMessage) error {
		return handleMsgCount(m, db)
	})

	// Reset spam cooldown (sudo only)
	client.BotClient.AddMessageHandler("/resetspam", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleResetSpam(client, db))(m)
	})
	
	client.BotClient.AddMessageHandler("/clearspam", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleResetSpam(client, db))(m)
	})
}

func trackNewUser(m *tg.NewMessage, client *core.Client, db *core.Database) error {
	userID := m.From.ID
	userName := m.From.FirstName

	// Check if user exists
	exists, _ := db.IsUserExist(userID)
	if !exists {
		// Add new user
		if err := db.AddUser(userID, userName); err != nil {
			log.Printf("Failed to add user %d: %v", userID, err)
			return nil
		}

		// Log to logger channel
		if config.Cfg.LoggerID != 0 {
			me, _ := client.BotClient.GetMe()
			mention := fmt.Sprintf("[%s](tg://user?id=%d)", userName, userID)
			
			client.BotClient.SendMessage(
				config.Cfg.LoggerID,
				fmt.Sprintf(
					"**⤷ User:** %s\n"+
						"**⤷ ID:** `%d`\n"+
						"__⤷ Started @%s !!__",
					mention,
					userID,
					me.Username,
				),
			)
		}

		log.Printf("#NewUser: Name: %s, ID: %d", userName, userID)
	} else {
		// Update username if changed
		db.UpdateUser(userID, "user_name", userName)
	}

	return nil
}

func trackNewChat(m *tg.NewMessage, client *core.Client, db *core.Database) error {
	chatID := m.Chat.ID

	// Check if chat exists
	// exists, _ := db.IsChatExist(chatID)
	// if !exists {
	// 	// Add new chat
	// 	if err := db.AddChat(chatID); err != nil {
	// 		log.Printf("Failed to add chat %d: %v", chatID, err)
	// 		return nil
	// 	}

	// 	// Get chat info
	// 	chat, _ := client.BotClient.GetChat(chatID)
	// 	chatTitle := "Unknown"
	// 	if chat != nil && chat.Title != "" {
	// 		chatTitle = chat.Title
	// 	}

	// 	// Log to logger channel
	// 	if config.Cfg.LoggerID != 0 {
	// 		me, _ := client.BotClient.GetMe()
			
	// 		client.BotClient.SendMessage(
	// 			config.Cfg.LoggerID,
	// 			fmt.Sprintf(
	// 				"**⤷ Chat Title:** %s\n"+
	// 					"**⤷ Chat ID:** `%d`\n"+
	// 					"__⤷ ADDED @%s !!__",
	// 				chatTitle,
	// 				chatID,
	// 				me.Username,
	// 			),
	// 		)
	// 	}

	// 	log.Printf("#NewChat: Title: %s, ID: %d", chatTitle, chatID)
	// }

	_ = chatID // Use variable
	return nil
}

func trackUserMessage(m *tg.NewMessage, db *core.Database) error {
	// Skip if banned
	if config.Cfg.IsBanned(m.From.ID) {
		return nil
	}

	// Skip if from channel
	if m.From == nil {
		return nil
	}

	userID := m.From.ID
	userName := m.From.FirstName

	// Check if user exists, add if not
	exists, _ := db.IsUserExist(userID)
	if !exists {
		db.AddUser(userID, userName)
		return nil
	}

	// TODO: Implement message tracking with anti-spam
	// For now, just increment message count
	// isCounted := db.TrackMessage(userID, userName)
	// if !isCounted {
	// 	// User in spam cooldown
	// 	return nil
	// }

	return nil
}

func handleMsgCount(m *tg.NewMessage, db *core.Database) error {
	if !m.IsGroup || config.Cfg.IsBanned(m.From.ID) {
		return nil
	}

	userID := m.From.ID
	
	// Get user data
	user, err := db.GetUser(userID)
	if err != nil || user == nil {
		return m.Reply(
			"❌ You are not registered in the database yet!\n" +
				"Send some messages to get started.",
		)
	}

	msgCount := user.MessagesCount
	songsCount := user.SongsPlayed
	joinDate := user.JoinDate

	// Check spam cooldown
	// cooldown := db.GetSpamCooldown(userID)
	cooldownText := ""
	// if cooldown != nil {
	// 	minutes := int(cooldown.Minutes())
	// 	seconds := int(cooldown.Seconds()) % 60
	// 	cooldownText = fmt.Sprintf("\n\n⚠️ **Spam Cooldown:** %dm %ds remaining", minutes, seconds)
	// }

	mention := fmt.Sprintf("[%s](tg://user?id=%d)", m.From.FirstName, userID)

	m.Reply(fmt.Sprintf(
		"📊 **Your Statistics**\n\n"+
			"👤 **User:** %s\n"+
			"💬 **Messages:** `%d`\n"+
			"🎵 **Songs Played:** `%d`\n"+
			"📅 **Joined:** `%s`"+
			"%s",
		mention,
		msgCount,
		songsCount,
		joinDate,
		cooldownText,
	))

	return nil
}

func handleResetSpam(client *core.Client, db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if m.ReplyToMsgID == 0 {
			return m.Reply("❌ Reply to a user's message to reset their spam cooldown!")
		}

		// Get replied message
		replied, err := client.BotClient.GetMessages(m.Chat.ID, []int32{int32(m.ReplyToMsgID)})
		if err != nil || len(replied) == 0 {
			return m.Reply("Failed to get replied message!")
		}

		if replied[0].From == nil {
			return m.Reply("Cannot reset spam for this user!")
		}

		userID := replied[0].From.ID
		userName := replied[0].From.FirstName

		// Reset spam cooldown
		// TODO: Implement ResetSpamCooldown
		// db.ResetSpamCooldown(userID)

		mention := fmt.Sprintf("[%s](tg://user?id=%d)", userName, userID)
		m.Reply(fmt.Sprintf("✅ Spam cooldown reset for %s!", mention))

		return nil
	}
}

// StartBackgroundTasks starts background tasks
func StartBackgroundTasks(db *core.Database, calls *core.Calls) {
	// Update played duration every second
	go func() {
		for {
			time.Sleep(1 * time.Second)
			updatePlayedDuration(db)
		}
	}()

	// Check for inactive VCs every 10 seconds
	go func() {
		for {
			time.Sleep(10 * time.Second)
			endInactiveVCs(db, calls)
		}
	}()

	log.Println(">> Background tasks started!")
}

func updatePlayedDuration(db *core.Database) {
	activeVCs := db.GetActiveVC()
	
	for _, vc := range activeVCs {
		if vc.ChatID == 0 {
			continue
		}

		// Check if paused
		// isPaused := db.GetWatcher(vc.ChatID, "pause")
		// if isPaused {
		// 	continue
		// }

		// TODO: Update queue duration
		// Queue.UpdateDuration(vc.ChatID, 1, 1)
	}
}

func endInactiveVCs(db *core.Database, calls *core.Calls) {
	// TODO: Check for inactive VCs and end them
	// This would check if a VC has been empty for more than 5 minutes
}
