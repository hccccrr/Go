package handlers

import (
	"fmt"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/utils"
	"shizumusic/helpers"
)

// TEXTS templates from utils
var TEXTS = helpers.TextTemplates

// RegisterActiveHandlers registers active VCs command handlers
func RegisterActiveHandlers(client *core.Client, db *core.Database) {
	// Active voice chats
	client.BotClient.AddMessageHandler("/active", func(m *tg.NewMessage) error {
		return handleActive(client, db)(m)
	})
}

func handleActive(client *core.Client, db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		// Check if sudo
		if !config.Cfg.IsSudo(m.From.ID) {
			return nil
		}

		msg, _ := m.Reply("Getting active voice chats...")

		// Get active VCs
		activeVCs := db.GetActiveVC()
		
		if len(activeVCs) == 0 || (len(activeVCs) == 1 && activeVCs[0].ChatID == 0) {
			return msg.Edit("No active voice chats found!")
		}

		// Build list
		text := "**🎙️ Active Voice Chats**\n\n"
		count := 0

		for _, vc := range activeVCs {
			if vc.ChatID == 0 {
				continue
			}

			count++
			
			// Get chat info
			chat, err := client.BotClient.GetChat(vc.ChatID)
			chatTitle := "Private Group"
			if err == nil && chat.Title != "" {
				chatTitle = chat.Title
			}

			// Calculate duration
			duration := time.Since(vc.JoinTime)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60

			// Get VC type
			vcType := "🎵 Audio"
			if vc.VCType == "video" {
				vcType = "📹 Video"
			}

			text += fmt.Sprintf(
				"**%d.** %s\n"+
					"├ **Type:** %s\n"+
					"├ **Active:** %dh %dm\n"+
					"└ **ID:** `%d`\n\n",
				count,
				chatTitle,
				vcType,
				hours,
				minutes,
				vc.ChatID,
			)
		}

		if count == 0 {
			return msg.Edit("No active voice chats found!")
		}

		text += fmt.Sprintf("**Total Active:** `%d`", count)
		msg.Edit(text)
		return nil
	}
}
