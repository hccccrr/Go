package handlers

import (
	"fmt"
	"os"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/utils"
	"shizumusic/helpers"
)

// TEXTS templates from utils
var TEXTS = helpers.TextTemplates

// RegisterSudoHandlers registers sudo command handlers
func RegisterSudoHandlers(client *core.Client, db *core.Database, calls *core.Calls) {
	// Autoend
	client.BotClient.AddMessageHandler("/autoend", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleAutoend(db))(m)
	})

	// Gban/Block
	client.BotClient.AddMessageHandler("/gban", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleGban(client, db))(m)
	})
	
	client.BotClient.AddMessageHandler("/block", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleGban(client, db))(m)
	})

	// Ungban/Unblock
	client.BotClient.AddMessageHandler("/ungban", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleUngban(client, db))(m)
	})
	
	client.BotClient.AddMessageHandler("/unblock", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleUngban(client, db))(m)
	})

	// Gbanlist/Blocklist
	client.BotClient.AddMessageHandler("/gbanlist", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleGbanlist(client, db, true))(m)
	})
	
	client.BotClient.AddMessageHandler("/blocklist", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleGbanlist(client, db, false))(m)
	})

	// Logs
	client.BotClient.AddMessageHandler("/logs", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleLogs(client))(m)
	})

	// Restart
	client.BotClient.AddMessageHandler("/restart", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleRestart(client, db, calls))(m)
	})

	// Sudolist
	client.BotClient.AddMessageHandler("/sudolist", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleSudolist(client))(m)
	})

	// Broadcast/Gcast
	client.BotClient.AddMessageHandler("/gcast", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleGcast(client, db))(m)
	})
}

func handleAutoend(db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		parts := strings.Fields(m.Text)
		if len(parts) != 2 {
			return m.Reply(
				"**Usage:**\n\n" +
					"__To turn off autoend:__ `/autoend off`\n" +
					"__To turn on autoend:__ `/autoend on`",
			)
		}

		cmd := strings.ToLower(parts[1])
		autoend, _ := db.GetAutoend()

		if cmd == "on" {
			if autoend {
				return m.Reply("AutoEnd is already enabled.")
			}
			// db.SetAutoend(true)
			return m.Reply(
				"AutoEnd Enabled! Now I will automatically end the stream after 5 minutes when the VC is empty.",
			)
		}

		if cmd == "off" {
			if !autoend {
				return m.Reply("AutoEnd is already disabled.")
			}
			// db.SetAutoend(false)
			return m.Reply("AutoEnd Disabled!")
		}

		return m.Reply(
			"**Usage:**\n\n" +
				"__To turn off autoend:__ `/autoend off`\n" +
				"__To turn on autoend:__ `/autoend on`",
		)
	}
}

func handleGban(client *core.Client, db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		var userID int64
		var userName string
		
		parts := strings.Fields(m.Text)
		cmd := strings.TrimPrefix(parts[0], "/")

		// Check if reply
		if m.ReplyToMsgID != 0 {
			replied, err := client.BotClient.GetMessages(m.Chat.ID, []int32{int32(m.ReplyToMsgID)})
			if err != nil || len(replied) == 0 {
				return m.Reply("Failed to get replied message!")
			}
			
			if replied[0].From != nil {
				userID = replied[0].From.ID
				userName = replied[0].From.FirstName
			}
		} else {
			if len(parts) != 2 {
				return m.Reply("Reply to a user's message or give their id.")
			}

			userStr := parts[1]
			user, err := client.BotClient.GetUser(userStr)
			if err != nil {
				return m.Reply("User not found!")
			}
			userID = user.ID
			userName = user.FirstName
		}

		mention := fmt.Sprintf("[%s](tg://user?id=%d)", userName, userID)

		// Validation checks
		if userID == m.From.ID {
			return m.Reply(fmt.Sprintf("You can't %s yourself.", cmd))
		}
		if userID == client.BotClient.Me.ID {
			return m.Reply(fmt.Sprintf("Yo! I'm not stupid to %s myself.", cmd))
		}
		if config.Cfg.IsSudo(userID) {
			return m.Reply(fmt.Sprintf("I can't %s my sudo users.", cmd))
		}

		// Check if already banned
		// isGbanned, _ := db.IsGbannedUser(userID)
		// if isGbanned {
		//     return m.Reply(fmt.Sprintf("%s is already in %s list.", mention, cmd))
		// }

		// Add to banned list
		config.Cfg.AddBanned(userID)

		if cmd == "gban" {
			// TODO: Ban from all chats
			// allChats := db.GetAllChats()
			// count := 0
			// for _, chat := range allChats {
			//     client.BotClient.BanChatMember(chat.ChatID, userID)
			//     count++
			// }
			
			// db.AddGbannedUser(userID)
			return m.Reply(fmt.Sprintf(
				"**Gbanned Successfully!**\n\n**User:** %s",
				mention,
			))
		}

		// db.AddBlockedUser(userID)
		return m.Reply(fmt.Sprintf(
			"**Blocked Successfully!**\n\n**User:** %s",
			mention,
		))
	}
}

func handleUngban(client *core.Client, db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		var userID int64
		var userName string

		parts := strings.Fields(m.Text)
		cmd := strings.TrimPrefix(parts[0], "/")

		if m.ReplyToMsgID != 0 {
			replied, err := client.BotClient.GetMessages(m.Chat.ID, []int32{int32(m.ReplyToMsgID)})
			if err != nil || len(replied) == 0 {
				return m.Reply("Failed to get replied message!")
			}
			
			if replied[0].From != nil {
				userID = replied[0].From.ID
				userName = replied[0].From.FirstName
			}
		} else {
			if len(parts) != 2 {
				return m.Reply("Reply to a user's message or give their id.")
			}

			userStr := parts[1]
			user, err := client.BotClient.GetUser(userStr)
			if err != nil {
				return m.Reply("User not found!")
			}
			userID = user.ID
			userName = user.FirstName
		}

		mention := fmt.Sprintf("[%s](tg://user?id=%d)", userName, userID)

		// Remove from banned list
		config.Cfg.RemoveBanned(userID)

		if cmd == "ungban" {
			// TODO: Unban from all chats
			// db.RemoveGbannedUser(userID)
			return m.Reply(fmt.Sprintf(
				"**Ungbanned Successfully!**\n\n**User:** %s",
				mention,
			))
		}

		// db.RemoveBlockedUser(userID)
		return m.Reply(fmt.Sprintf(
			"**Unblocked Successfully!**\n\n**User:** %s",
			mention,
		))
	}
}

func handleGbanlist(client *core.Client, db *core.Database, isGban bool) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		var users []int64
		var title string

		if isGban {
			users, _ = db.GetGbannedUsers()
			title = "**Gbanned Users:**\n\n"
		} else {
			users, _ = db.GetBlockedUsers()
			title = "**Blocked Users:**\n\n"
		}

		if len(users) == 0 {
			if isGban {
				return m.Reply("No Gbanned Users Found!")
			}
			return m.Reply("No Blocked Users Found!")
		}

		msg, _ := m.Reply("Fetching list...")
		text := title
		
		for i, userID := range users {
			user, err := client.BotClient.GetUser(userID)
			if err != nil {
				text += fmt.Sprintf("%02d: [User] `%d`\n", i+1, userID)
				continue
			}
			
			userName := fmt.Sprintf("[%s](tg://user?id=%d)", user.FirstName, userID)
			text += fmt.Sprintf("%02d: %s `%d`\n", i+1, userName, userID)
		}

		msg.Edit(text)
		return nil
	}
}

func handleLogs(client *core.Client) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		logFile := "ShizuMusic.log"
		
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			return m.Reply("No Logs Found!")
		}

		// Send log file
		m.Reply("**Logs:**", &tg.MediaOptions{
			File: logFile,
		})
		
		return nil
	}
}

func handleRestart(client *core.Client, db *core.Database, calls *core.Calls) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		msg, _ := m.Reply("Notifying chats about restart...")

		// Get active VCs
		activeVCs := db.GetActiveVC()
		count := 0

		for _, vc := range activeVCs {
			if vc.ChatID == 0 {
				continue
			}

			// Notify and leave
			client.BotClient.SendMessage(
				vc.ChatID,
				"**Bot is restarting in a minute or two.**\n\n"+
					"Please wait for a minute before using me again.",
			)
			
			calls.LeaveVC(vc.ChatID)
			count++
		}

		msg.Edit(fmt.Sprintf(
			"Notified **%d** chat(s) about the restart.\n\nRestarting now...",
			count,
		))

		// Clean up
		os.RemoveAll("cache")
		os.RemoveAll("downloads")

		// Restart
		os.Exit(0)
		return nil
	}
}

func handleSudolist(client *core.Client) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		text := "**⟢ God Users:**\n"
		gods := 0

		config.Cfg.GodMutex.RLock()
		for userID := range config.Cfg.GodUsers {
			user, err := client.BotClient.GetUser(userID)
			if err != nil {
				continue
			}
			
			gods++
			userName := fmt.Sprintf("[%s](tg://user?id=%d)", user.FirstName, userID)
			text += fmt.Sprintf("%02d: %s\n", gods, userName)
		}
		config.Cfg.GodMutex.RUnlock()

		text += "\n**⟢ Sudo Users:**\n"
		sudos := 0

		config.Cfg.SudoMutex.RLock()
		for userID := range config.Cfg.SudoUsers {
			if config.Cfg.IsGod(userID) {
				continue
			}

			user, err := client.BotClient.GetUser(userID)
			if err != nil {
				continue
			}
			
			sudos++
			userName := fmt.Sprintf("[%s](tg://user?id=%d)", user.FirstName, userID)
			text += fmt.Sprintf("%02d: %s\n", gods+sudos, userName)
		}
		config.Cfg.SudoMutex.RUnlock()

		if gods == 0 && sudos == 0 {
			return m.Reply("No sudo users found.")
		}

		m.Reply(text)
		return nil
	}
}

func handleGcast(client *core.Client, db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if m.ReplyToMsgID == 0 {
			return m.Reply("Reply to a message to broadcast it.")
		}

		parts := strings.Fields(m.Text)
		if len(parts) == 1 {
			return m.Reply(
				"Where to gcast?\n\n" +
					"**With Forward Tag:** `/gcast chats`\n" +
					"- `/gcast users`\n" +
					"- `/gcast all`\n\n" +
					"**Without Forward Tag:** `/gcast chats copy`\n" +
					"- `/gcast users copy`\n" +
					"- `/gcast all copy`",
			)
		}

		broadcastType := strings.ToLower(parts[1])
		copy := len(parts) >= 3 && strings.ToLower(parts[2]) == "copy"

		// TODO: Implement broadcast logic
		_ = broadcastType
		_ = copy

		m.Reply("Gcast feature coming soon!")
		return nil
	}
}
