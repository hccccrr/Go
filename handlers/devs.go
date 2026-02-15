package handlers

import (
	"fmt"
	"os/exec"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/utils"
	"shizumusic/helpers"
)

// TEXTS templates from utils
var TEXTS = helpers.TextTemplates

// RegisterDevHandlers registers developer command handlers
func RegisterDevHandlers(client *core.Client, db *core.Database) {
	// Eval/Run command
	client.BotClient.AddMessageHandler("/eval", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleEval(client))(m)
	})
	
	client.BotClient.AddMessageHandler("/run", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleEval(client))(m)
	})

	// Exec/Term/Shell commands
	client.BotClient.AddMessageHandler("/exec", func(m *tg.NewMessage) error {
		return core.OwnerOnly(handleExec(client))(m)
	})
	
	client.BotClient.AddMessageHandler("/term", func(m *tg.NewMessage) error {
		return core.OwnerOnly(handleExec(client))(m)
	})
	
	client.BotClient.AddMessageHandler("/sh", func(m *tg.NewMessage) error {
		return core.OwnerOnly(handleExec(client))(m)
	})
	
	client.BotClient.AddMessageHandler("/shell", func(m *tg.NewMessage) error {
		return core.OwnerOnly(handleExec(client))(m)
	})

	// Get variable
	client.BotClient.AddMessageHandler("/getvar", func(m *tg.NewMessage) error {
		return core.OwnerOnly(handleGetvar)(m)
	})
	
	client.BotClient.AddMessageHandler("/gvar", func(m *tg.NewMessage) error {
		return core.OwnerOnly(handleGetvar)(m)
	})
	
	client.BotClient.AddMessageHandler("/var", func(m *tg.NewMessage) error {
		return core.OwnerOnly(handleGetvar)(m)
	})

	// Add/Remove sudo
	client.BotClient.AddMessageHandler("/addsudo", func(m *tg.NewMessage) error {
		return core.OwnerOnly(handleAddsudo(client, db))(m)
	})
	
	client.BotClient.AddMessageHandler("/delsudo", func(m *tg.NewMessage) error {
		return core.OwnerOnly(handleDelsudo(client, db))(m)
	})
	
	client.BotClient.AddMessageHandler("/rmsudo", func(m *tg.NewMessage) error {
		return core.OwnerOnly(handleDelsudo(client, db))(m)
	})

	// Update command
	client.BotClient.AddMessageHandler("/update", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleUpdate(client))(m)
	})
	
	client.BotClient.AddMessageHandler("/gitpull", func(m *tg.NewMessage) error {
		return core.SudoOnly(handleUpdate(client))(m)
	})
}

func handleEval(client *core.Client) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		msg, _ := m.Reply("**Processing...**")
		
		parts := strings.SplitN(m.Text, " ", 2)
		if len(parts) != 2 {
			return msg.Edit("**Received empty message!**")
		}

		code := strings.TrimSpace(parts[1])
		
		// Note: Go doesn't have runtime eval like Python
		// This is a placeholder - actual implementation would need
		// a Go interpreter or code execution sandbox
		
		output := fmt.Sprintf(
			"<b>EVAL</b>: <code>%s</code>\n\n<b>OUTPUT</b>:\n"+
				"<code>Note: Go doesn't support runtime code evaluation like Python.\n"+
				"This feature requires a Go interpreter.</code>",
			code,
		)

		msg.Edit(output)
		return nil
	}
}

func handleExec(client *core.Client) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		msg, _ := m.Reply("**Processing...**")
		
		parts := strings.SplitN(m.Text, " ", 2)
		if len(parts) != 2 {
			return msg.Edit("**Received empty message!**")
		}

		command := strings.TrimSpace(parts[1])
		
		// Execute shell command
		cmd := exec.Command("sh", "-c", command)
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			return msg.Edit(fmt.Sprintf("**Error:**\n`%s`", err.Error()))
		}

		outputStr := string(output)
		if outputStr == "" || outputStr == "\n" {
			outputStr = "No Output"
		}

		// Limit output length
		if len(outputStr) > 4000 {
			outputStr = outputStr[:4000] + "\n... (truncated)"
		}

		msg.Edit(fmt.Sprintf("**Output:**\n`%s`", outputStr))
		return nil
	}
}

func handleGetvar() core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		parts := strings.Fields(m.Text)
		if len(parts) != 2 {
			return m.Reply("**Give a variable name to get its value.**")
		}

		varName := strings.ToUpper(parts[1])
		
		// Map of available variables
		vars := map[string]interface{}{
			"API_ID":       config.Cfg.APIID,
			"BOT_TOKEN":    "***HIDDEN***",
			"DATABASE_URL": "***HIDDEN***",
			"OWNER_ID":     config.Cfg.OwnerID,
			"BOT_NAME":     config.Cfg.BotName,
			"PLAY_LIMIT":   config.Cfg.PlayLimit,
			"PRIVATE_MODE": config.Cfg.PrivateMode,
		}

		value, exists := vars[varName]
		if !exists {
			return m.Reply("**Give a valid variable name to get its value.**")
		}

		m.Reply(fmt.Sprintf("**%s:** `%v`", varName, value))
		return nil
	}
}

func handleAddsudo(client *core.Client, db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		var userID int64
		var userName string

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
			parts := strings.Fields(m.Text)
			if len(parts) != 2 {
				return m.Reply("**Reply to a user or give a user id to add them as sudo.**")
			}

			userStr := strings.TrimPrefix(parts[1], "@")
			user, err := client.BotClient.GetUser(userStr)
			if err != nil {
				return m.Reply("User not found!")
			}
			userID = user.ID
			userName = user.FirstName
		}

		mention := fmt.Sprintf("[%s](tg://user?id=%d)", userName, userID)

		if config.Cfg.IsSudo(userID) {
			return m.Reply(fmt.Sprintf("**%s is already a sudo user.**", mention))
		}

		// Add to sudo
		if err := db.AddSudo(userID); err != nil {
			return m.Reply("**Failed to add sudo user.**")
		}

		config.Cfg.AddSudo(userID)
		m.Reply(fmt.Sprintf("**%s is now a sudo user.**", mention))
		return nil
	}
}

func handleDelsudo(client *core.Client, db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		var userID int64
		var userName string

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
			parts := strings.Fields(m.Text)
			if len(parts) != 2 {
				return m.Reply("**Reply to a user or give a user id to remove them from sudo.**")
			}

			userStr := strings.TrimPrefix(parts[1], "@")
			user, err := client.BotClient.GetUser(userStr)
			if err != nil {
				return m.Reply("User not found!")
			}
			userID = user.ID
			userName = user.FirstName
		}

		mention := fmt.Sprintf("[%s](tg://user?id=%d)", userName, userID)

		if !config.Cfg.IsSudo(userID) {
			return m.Reply(fmt.Sprintf("**%s is not a sudo user.**", mention))
		}

		// Remove from sudo
		// TODO: db.RemoveSudo(userID)
		
		config.Cfg.SudoMutex.Lock()
		delete(config.Cfg.SudoUsers, userID)
		config.Cfg.SudoMutex.Unlock()

		m.Reply(fmt.Sprintf("**%s is no longer a sudo user.**", mention))
		return nil
	}
}

func handleUpdate(client *core.Client) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		// Check if on Heroku
		if _, exists := config.Cfg.Cache["DYNO"]; exists {
			return m.Reply(
				"**Heroku Update:**\n" +
					"Please use Heroku dashboard or CLI to update your bot on Heroku.",
			)
		}

		msg, _ := m.Reply("**Checking for updates...**")
		
		// Git pull
		cmd := exec.Command("git", "pull")
		output, err := cmd.CombinedOutput()
		
		if err != nil {
			return msg.Edit(fmt.Sprintf("**Git Error:**\n`%s`", err.Error()))
		}

		outputStr := string(output)
		
		if strings.Contains(outputStr, "Already up to date") {
			return msg.Edit("**Bot is up-to-date!**")
		}

		msg.Edit(fmt.Sprintf(
			"**Update pulled successfully!**\n\n"+
				"Restart the bot to apply changes.\n\n"+
				"Output: `%s`",
			outputStr,
		))
		
		return nil
	}
}
