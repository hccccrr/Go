package handlers

import (
	"fmt"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
)

// RegisterCallbackHandlers registers callback query handlers
func RegisterCallbackHandlers(client *core.Client, db *core.Database, calls *core.Calls) {
	// Close button
	client.BotClient.OnCallbackQuery(func(cb *tg.CallbackQuery) error {
		if strings.HasPrefix(string(cb.Data), "close") {
			return handleClose(cb)
		}
		
		if strings.HasPrefix(string(cb.Data), "controls") {
			return handleControls(cb, client, db, calls)
		}
		
		if strings.HasPrefix(string(cb.Data), "player") {
			return handlePlayer(cb)
		}
		
		if strings.HasPrefix(string(cb.Data), "ctrl") {
			return handleCtrl(cb, db, calls)
		}
		
		if strings.HasPrefix(string(cb.Data), "help") {
			return handleHelpCallback(cb, client)
		}
		
		return nil
	})
}

func handleClose(cb *tg.CallbackQuery) error {
	if config.Cfg.IsBanned(cb.From.ID) {
		return nil
	}

	cb.Answer("Closed!", true)
	cb.Message.Delete()
	return nil
}

func handleControls(cb *tg.CallbackQuery, client *core.Client, db *core.Database, calls *core.Calls) error {
	if config.Cfg.IsBanned(cb.From.ID) {
		return nil
	}

	// Parse data: controls|video_id|chat_id
	parts := strings.Split(string(cb.Data), "|")
	if len(parts) < 3 {
		return nil
	}

	// TODO: Show control buttons
	cb.Answer("Controls", false)
	return nil
}

func handlePlayer(cb *tg.CallbackQuery) error {
	if config.Cfg.IsBanned(cb.From.ID) {
		return nil
	}

	// TODO: Show player info
	cb.Answer("Player", false)
	return nil
}

func handleCtrl(cb *tg.CallbackQuery, db *core.Database, calls *core.Calls) error {
	if config.Cfg.IsBanned(cb.From.ID) {
		return nil
	}

	// Parse data: ctrl|action|chat_id
	parts := strings.Split(string(cb.Data), "|")
	if len(parts) < 3 {
		return cb.Answer("Invalid callback data!", true)
	}

	action := parts[1]
	// chatIDStr := parts[2]

	mention := fmt.Sprintf("[%s](tg://user?id=%d)", cb.From.FirstName, cb.From.ID)

	switch action {
	case "play":
		// Toggle pause/resume
		cb.Answer("Play/Pause!", false)
		cb.Message.Reply(fmt.Sprintf("__VC Toggled by:__ %s", mention))

	case "mute":
		cb.Answer("Muted!", false)
		cb.Message.Reply(fmt.Sprintf("__VC Muted by:__ %s", mention))

	case "unmute":
		cb.Answer("Unmuted!", false)
		cb.Message.Reply(fmt.Sprintf("__VC Unmuted by:__ %s", mention))

	case "end":
		cb.Answer("Left VC!", false)
		cb.Message.Reply(fmt.Sprintf("__VC Stopped by:__ %s", mention))

	case "skip":
		cb.Answer("Skipped!", false)

	case "replay":
		cb.Answer("Replaying!", false)

	case "loop":
		cb.Answer("Loop updated!", false)

	case "bass":
		cb.Answer("Bass boost toggled!", false)

	case "speed":
		cb.Answer("Speed changed!", false)

	default:
		cb.Answer("Unknown action!", true)
	}

	return nil
}

func handleHelpCallback(cb *tg.CallbackQuery, client *core.Client) error {
	if config.Cfg.IsBanned(cb.From.ID) {
		return nil
	}

	// Parse data: help|category
	parts := strings.Split(string(cb.Data), "|")
	if len(parts) < 2 {
		return nil
	}

	category := parts[1]
	var text string

	switch category {
	case "admin":
		text = "**⚙️ Admin Commands**\n\n" +
			"/pause - Pause playback\n" +
			"/resume - Resume playback\n" +
			"/skip - Skip current song\n" +
			"/end - Stop VC\n" +
			"/loop - Set loop count"

	case "user":
		text = "**👤 User Commands**\n\n" +
			"/play - Play a song\n" +
			"/vplay - Play video\n" +
			"/queue - Show queue\n" +
			"/current - Current song"

	case "sudo":
		text = "**👑 Sudo Commands**\n\n" +
			"/stats - Bot statistics\n" +
			"/gban - Ban user globally\n" +
			"/restart - Restart bot"

	case "owner":
		text = "**🔧 Owner Commands**\n\n" +
			"/eval - Execute code\n" +
			"/exec - Shell command\n" +
			"/addsudo - Add sudo user"

	case "back":
		me, _ := client.BotClient.GetMe()
		text = fmt.Sprintf("📚 **Help Menu**\n\n@%s", me.Username)

	default:
		text = "Unknown category!"
	}

	cb.Answer("", false)
	cb.Message.Edit(text)
	return nil
}
