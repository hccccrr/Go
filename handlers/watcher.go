package handlers

import (
	"fmt"
	"log"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
)

// RegisterWatcherHandlers registers event watchers
func RegisterWatcherHandlers(client *core.Client, db *core.Database) {
	// Track new users in PM
	client.BotClient.AddMessageHandler("", func(m *tg.NewMessage) error {
		if m.IsPrivate() {
			return trackNewUser(m, client, db)
		}
		return nil
	})

	// Track user messages for leaderboard
	client.BotClient.AddMessageHandler("", func(m *tg.NewMessage) error {
		sender, err := m.GetSender()
		if err != nil || sender == nil {
			return nil
		}
		if m.IsGroup() {
			return trackUserMessage(m, sender, db)
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
	sender, err := m.GetSender()
	if err != nil || sender == nil {
		return nil
	}

	userID := sender.ID
	userName := sender.FirstName

	exists, _ := db.IsUserExist(userID)
	if !exists {
		if err := db.AddUser(userID, userName); err != nil {
			log.Printf("Failed to add user %d: %v", userID, err)
			return nil
		}

		if config.Cfg.LoggerID != 0 {
			me, _ := client.BotClient.GetMe()
			mention := fmt.Sprintf("[%s](tg://user?id=%d)", userName, userID)
			client.BotClient.SendMessage(
				config.Cfg.LoggerID,
				fmt.Sprintf(
					"**↷ User:** %s\n**↷ ID:** `%d`\n__↷ Started @%s !!__",
					mention, userID, me.Username,
				),
			)
		}

		log.Printf("#NewUser: Name: %s, ID: %d", userName, userID)
	} else {
		db.UpdateUser(userID, "user_name", userName)
	}

	return nil
}

func trackUserMessage(m *tg.NewMessage, sender *tg.UserObj, db *core.Database) error {
	if config.Cfg.IsBanned(sender.ID) {
		return nil
	}

	userID := sender.ID
	userName := sender.FirstName

	exists, _ := db.IsUserExist(userID)
	if !exists {
		db.AddUser(userID, userName)
	}

	return nil
}

func handleMsgCount(m *tg.NewMessage, db *core.Database) error {
	if !m.IsGroup() {
		return nil
	}

	sender, err := m.GetSender()
	if err != nil || sender == nil {
		return nil
	}

	if config.Cfg.IsBanned(sender.ID) {
		return nil
	}

	user, err := db.GetUser(sender.ID)
	if err != nil || user == nil {
		_, _ = m.Reply("❌ You are not registered yet!\nSend some messages to get started.")
		return nil
	}

	mention := fmt.Sprintf("[%s](tg://user?id=%d)", sender.FirstName, sender.ID)
	_, _ = m.Reply(fmt.Sprintf(
		"📊 **Your Statistics**\n\n"+
			"👤 **User:** %s\n"+
			"💬 **Messages:** `%d`\n"+
			"🎵 **Songs Played:** `%d`\n"+
			"📅 **Joined:** `%s`",
		mention,
		user.MessagesCount,
		user.SongsPlayed,
		user.JoinDate,
	))

	return nil
}

func handleResetSpam(client *core.Client, db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if m.ReplyToMsgID == 0 {
			_, _ = m.Reply("❌ Reply to a user's message to reset their spam cooldown!")
			return nil
		}

		replied, err := client.BotClient.GetMessages(m.ChatID(), []int32{int32(m.ReplyToMsgID)})
		if err != nil || len(replied) == 0 {
			_, _ = m.Reply("Failed to get replied message!")
			return nil
		}

		replyMsg := replied[0]
		replyFrom := replyMsg.Sender()
		if replyFrom == nil {
			_, _ = m.Reply("Cannot reset spam for this user!")
			return nil
		}

		if user, ok := replyFrom.(*tg.UserObj); ok {
			mention := fmt.Sprintf("[%s](tg://user?id=%d)", user.FirstName, user.ID)
			_, _ = m.Reply(fmt.Sprintf("✅ Spam cooldown reset for %s!", mention))
		}

		return nil
	}
}

// StartBackgroundTasks starts background tasks
func StartBackgroundTasks(db *core.Database, calls *core.Calls) {
	go func() {
		for {
			time.Sleep(1 * time.Second)
			updatePlayedDuration(db)
		}
	}()

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
		// TODO: utils.Queue.UpdateDuration(vc.ChatID, 1, 1)
	}
}

func endInactiveVCs(db *core.Database, calls *core.Calls) {
	// TODO: Check for inactive VCs and end them
}
