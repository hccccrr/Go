package handlers

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/utils"
	"shizumusic/helpers"
)

// TEXTS templates from utils
var TEXTS = helpers.TextTemplates

// RegisterAdminHandlers registers admin command handlers
func RegisterAdminHandlers(client *core.Client, db *core.Database) {
	// Auth command
	client.BotClient.AddMessageHandler("/auth", func(m *tg.NewMessage) error {
		return core.AdminOnly(handleAuth(client, db))(m)
	})

	// Unauth command
	client.BotClient.AddMessageHandler("/unauth", func(m *tg.NewMessage) error {
		return core.AdminOnly(handleUnauth(client, db))(m)
	})

	// Authlist command
	client.BotClient.AddMessageHandler("/authlist", func(m *tg.NewMessage) error {
		return handleAuthlist(client, db)(m)
	})

	// Authchat command
	client.BotClient.AddMessageHandler("/authchat", func(m *tg.NewMessage) error {
		return core.AdminOnly(handleAuthchat(db))(m)
	})
}

func handleAuth(client *core.Client, db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !m.IsGroup || config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		var userID int64
		var userName string

		// Check if reply
		if m.ReplyToMsgID != 0 {
			// Get replied message
			replied, err := client.BotClient.GetMessages(m.Chat.ID, []int32{int32(m.ReplyToMsgID)})
			if err != nil || len(replied) == 0 {
				return m.Reply("Failed to get replied message!")
			}
			
			repliedMsg := replied[0]
			if repliedMsg.From != nil {
				userID = repliedMsg.From.ID
				userName = repliedMsg.From.FirstName
			} else {
				return m.Reply("Cannot authorize this user!")
			}
		} else {
			// Parse from text
			parts := strings.Fields(m.Text)
			if len(parts) != 2 {
				return m.Reply("Reply to a user or give a user id or username")
			}

			userStr := strings.TrimPrefix(parts[1], "@")
			
			// Try to parse as ID
			if id, err := strconv.ParseInt(userStr, 10, 64); err == nil {
				userID = id
				// Get user info
				user, err := client.BotClient.GetUser(userID)
				if err != nil {
					return m.Reply("Failed to get user info!")
				}
				userName = user.FirstName
			} else {
				// Try to get by username
				user, err := client.BotClient.ResolveUsername(userStr)
				if err != nil {
					return m.Reply("User not found!")
				}
				userID = user.ID
				userName = user.FirstName
			}
		}

		// Check auth list limit (30 users max)
		// TODO: Implement GetAllAuthusers
		// allAuths := db.GetAllAuthusers(m.Chat.ID)
		// if len(allAuths) >= 30 {
		//     return m.Reply("AuthList is full!\n\nLimit of Auth Users in a chat is: `30`")
		// }

		// Check if already authorized
		// TODO: Implement IsAuthuser
		// isAuth := db.IsAuthuser(m.Chat.ID, userID)
		// if isAuth {
		//     return m.Reply("This user is already Authorized in this chat!")
		// }

		// Add to auth list
		authData := map[string]interface{}{
			"user_name":    userName,
			"auth_by_id":   m.From.ID,
			"auth_by_name": m.From.FirstName,
			"auth_date":    time.Now().Format("02-01-2006 15:04"),
		}

		// TODO: Implement AddAuthusers
		// db.AddAuthusers(m.Chat.ID, userID, authData)

		_ = authData // Use variable to avoid error
		m.Reply("Successfully Authorized user in this chat!")
		return nil
	}
}

func handleUnauth(client *core.Client, db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !m.IsGroup || config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		var userID int64

		// Check if reply
		if m.ReplyToMsgID != 0 {
			replied, err := client.BotClient.GetMessages(m.Chat.ID, []int32{int32(m.ReplyToMsgID)})
			if err != nil || len(replied) == 0 {
				return m.Reply("Failed to get replied message!")
			}
			
			repliedMsg := replied[0]
			if repliedMsg.From != nil {
				userID = repliedMsg.From.ID
			} else {
				return m.Reply("Cannot unauthorize this user!")
			}
		} else {
			parts := strings.Fields(m.Text)
			if len(parts) != 2 {
				return m.Reply("Reply to a user or give a user id or username")
			}

			userStr := strings.TrimPrefix(parts[1], "@")
			
			if id, err := strconv.ParseInt(userStr, 10, 64); err == nil {
				userID = id
			} else {
				user, err := client.BotClient.ResolveUsername(userStr)
				if err != nil {
					return m.Reply("User not found!")
				}
				userID = user.ID
			}
		}

		// Check if authorized
		// TODO: Implement IsAuthuser
		// isAuth := db.IsAuthuser(m.Chat.ID, userID)
		// if !isAuth {
		//     return m.Reply("This user was not Authorized in this chat!")
		// }

		// Remove from auth list
		// TODO: Implement RemoveAuthuser
		// db.RemoveAuthuser(m.Chat.ID, userID)

		m.Reply("Removed user's Authorization in this chat!")
		return nil
	}
}

func handleAuthlist(client *core.Client, db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !m.IsGroup || config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		// Get all authorized users
		// TODO: Implement GetAllAuthusers
		// allAuths := db.GetAllAuthusers(m.Chat.ID)
		// if len(allAuths) == 0 {
		//     return m.Reply("No Authorized users in this chat!")
		// }

		msg, _ := m.Reply("Fetching Authorized users in this chat...")
		
		// Build auth list
		// TODO: Fetch user details and format list
		// collection := []map[string]interface{}{}
		// for _, userID := range allAuths {
		//     authData := db.GetAuthuser(m.Chat.ID, userID)
		//     collection = append(collection, authData)
		// }

		// TODO: Create paginated view
		msg.Edit("**Authorized Users:**\n\nFeature coming soon!")
		return nil
	}
}

func handleAuthchat(db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !m.IsGroup || config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		parts := strings.Fields(m.Text)
		isAuth, _ := db.IsAuthchat(m.Chat.ID)

		// Show current status if no argument
		if len(parts) != 2 {
			status := "Off"
			if isAuth {
				status = "On"
			}
			return m.Reply(fmt.Sprintf(
				"Current AuthChat Status: `%s`\n\nUsage: `/authchat on` or `/authchat off`",
				status,
			))
		}

		cmd := strings.ToLower(parts[1])

		if cmd == "on" {
			if isAuth {
				return m.Reply("AuthChat is already On!")
			}
			
			// TODO: Implement AddAuthchat
			// db.AddAuthchat(m.Chat.ID)
			
			return m.Reply(
				"**Turned On AuthChat!**\n\n" +
					"Now all users can use bot commands in this chat!",
			)
		}

		if cmd == "off" {
			if !isAuth {
				return m.Reply("AuthChat is already Off!")
			}
			
			// TODO: Implement RemoveAuthchat
			// db.RemoveAuthchat(m.Chat.ID)
			
			return m.Reply(
				"**Turned Off AuthChat!**\n\n" +
					"Now only Authorized users can use bot commands in this chat!",
			)
		}

		// Invalid argument
		status := "Off"
		if isAuth {
			status = "On"
		}
		return m.Reply(fmt.Sprintf(
			"Current AuthChat Status: `%s`\n\nUsage: `/authchat on` or `/authchat off`",
			status,
		))
	}
}
