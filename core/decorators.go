package core

import (
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
)

// Decorator function type
type HandlerFunc func(*tg.NewMessage) error

// CheckMode checks if private mode is enabled
func CheckMode(handler HandlerFunc) HandlerFunc {
	return func(m *tg.NewMessage) error {
		onModes := []string{"on", "enable", "yes", "true", "1"}
		privateMode := strings.ToLower(config.Cfg.PrivateMode)

		for _, mode := range onModes {
			if privateMode == mode {
				if !config.Cfg.IsSudo(m.From.ID) {
					m.Reply("**🔒 Private Mode Enabled**\n\n" +
						"This bot is in private mode and only authorized users can use it.")
					return nil
				}
				break
			}
		}

		return handler(m)
	}
}

// AdminOnly checks if user is admin with voice chat permissions
func AdminOnly(handler HandlerFunc) HandlerFunc {
	return func(m *tg.NewMessage) error {
		// Delete command message
		_ = m.Delete()

		// Check for anonymous admin
		if m.From.ID == m.Chat.ID {
			m.Reply("**❌ Anonymous Admin Detected**\n\n" +
				"You're an anonymous admin. Please revert to your personal account to use this command.")
			return nil
		}

		// Bypass for sudo users
		if config.Cfg.IsSudo(m.From.ID) {
			return handler(m)
		}

		// Check admin rights
		// TODO: Implement get_user_rights check
		// For now, allow all users (implement proper check later)

		return handler(m)
	}
}

// AuthOnly checks if user is authorized
func AuthOnly(db *Database) func(HandlerFunc) HandlerFunc {
	return func(handler HandlerFunc) HandlerFunc {
		return func(m *tg.NewMessage) error {
			// Delete command message
			_ = m.Delete()

			// Check for anonymous admin
			if m.From.ID == m.Chat.ID {
				m.Reply("**❌ Anonymous Admin Detected**\n\n" +
					"You're an anonymous admin. Please revert to your personal account to use this command.")
				return nil
			}

			// Check if VC is active
			active, _ := db.IsActiveVC(m.Chat.ID)
			if !active {
				m.Reply("**❌ No Active Stream**\n\n" +
					"Nothing is currently playing in the voice chat!")
				return nil
			}

			// Check if authchat is enabled
			isAuthChat, _ := db.IsAuthchat(m.Chat.ID)

			if !isAuthChat {
				// Bypass for sudo users
				if config.Cfg.IsSudo(m.From.ID) {
					return handler(m)
				}

				// TODO: Get authorized users list and check
				// For now, allow all (implement proper check later)
			}

			return handler(m)
		}
	}
}

// UserOnly allows all users except anonymous admins
func UserOnly(handler HandlerFunc) HandlerFunc {
	return func(m *tg.NewMessage) error {
		// Delete command message
		_ = m.Delete()

		// Check for anonymous admin
		if m.From.ID == m.Chat.ID {
			m.Reply("**❌ Anonymous Admin Detected**\n\n" +
				"You're an anonymous admin. Please revert to your personal account to use this command.")
			return nil
		}

		return handler(m)
	}
}

// SudoOnly checks if user is sudo
func SudoOnly(handler HandlerFunc) HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !config.Cfg.IsSudo(m.From.ID) {
			m.Reply("**❌ Unauthorized**\n\n" +
				"This command is only for sudo users!")
			return nil
		}

		return handler(m)
	}
}

// OwnerOnly checks if user is owner
func OwnerOnly(handler HandlerFunc) HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !config.Cfg.IsGod(m.From.ID) {
			m.Reply("**❌ Unauthorized**\n\n" +
				"This command is only for the bot owner!")
			return nil
		}

		return handler(m)
	}
}

// PlayContext holds play command context
type PlayContext struct {
	IsVideo   bool
	IsForce   bool
	IsURL     string
	IsTGAudio interface{}
	IsTGVideo interface{}
}

// PlayWrapper validates and prepares playback context
func PlayWrapper(handler func(*tg.NewMessage, *PlayContext) error) HandlerFunc {
	return func(m *tg.NewMessage) error {
		// Delete command message
		_ = m.Delete()

		// Check for anonymous admin
		if m.From.ID == m.Chat.ID {
			m.Reply("**❌ Anonymous Admin Detected**\n\n" +
				"You're an anonymous admin. Please revert to your personal account to use this command.")
			return nil
		}

		ctx := &PlayContext{}

		// Parse command
		parts := strings.Fields(m.Text)
		if len(parts) == 0 {
			return nil
		}

		command := strings.ToLower(strings.TrimPrefix(parts[0], "/"))

		// Check for video/force flags
		if strings.HasPrefix(command, "v") {
			ctx.IsVideo = true
		}
		if strings.HasPrefix(command, "f") {
			ctx.IsForce = true
			if len(command) > 1 && command[1] == 'v' {
				ctx.IsVideo = true
			}
		}

		// Check for URL in message
		for _, part := range parts[1:] {
			if strings.Contains(part, "youtube.com") || strings.Contains(part, "youtu.be") {
				ctx.IsURL = part
				break
			}
		}

		// Check for replied media
		if m.ReplyToMsgID != 0 {
			// Get replied message
			// TODO: Implement media detection from reply
			// ctx.IsTGAudio = ...
			// ctx.IsTGVideo = ...
		}

		// Validate input
		if ctx.IsTGAudio == nil && ctx.IsTGVideo == nil && ctx.IsURL == "" && len(parts) < 2 {
			m.Reply("**❌ Invalid Input**\n\n" +
				"**Usage:**\n" +
				"• Reply to an audio/video file\n" +
				"• Provide a YouTube link\n" +
				"• Search with a query\n\n" +
				"**Examples:**\n" +
				"`/play Faded`\n" +
				"`/vplay Despacito`")
			return nil
		}

		return handler(m, ctx)
	}
}
