package helpers

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"
)

// Button represents a Telegram inline button
type Button struct {
	Text string
	Data string
	URL  string
}

// MakeButtons handles creation of Telegram inline keyboards
type MakeButtons struct{}

// NewMakeButtons creates a new MakeButtons instance
func NewMakeButtons() *MakeButtons {
	return &MakeButtons{}
}

// CloseMarkup returns a close button
func (mb *MakeButtons) CloseMarkup() *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "🗑", Data: []byte("close")},
				},
			},
		},
	}
}

// QueueMarkup returns queue navigation buttons
func (mb *MakeButtons) QueueMarkup(count, page int) *tg.InlineKeyboardMarkup {
	if count != 1 {
		return &tg.InlineKeyboardMarkup{
			Rows: []tg.KeyboardButtonRow{
				{
					Buttons: []tg.KeyboardButton{
						{Text: "◂", Data: []byte(fmt.Sprintf("queue|prev|%d", page))},
						{Text: "🗑", Data: []byte("close")},
						{Text: "▸", Data: []byte(fmt.Sprintf("queue|next|%d", page))},
					},
				},
			},
		}
	}
	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "🗑", Data: []byte("close")},
				},
			},
		},
	}
}

// PlayFavsMarkup returns play favorites buttons
func (mb *MakeButtons) PlayFavsMarkup(userID int64) *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "Audio", Data: []byte(fmt.Sprintf("favsplay|audio|%d", userID))},
					{Text: "Video", Data: []byte(fmt.Sprintf("favsplay|video|%d", userID))},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "🗑", Data: []byte(fmt.Sprintf("favsplay|close|%d", userID))},
				},
			},
		},
	}
}

// FavoriteMarkup returns favorites list with navigation
func (mb *MakeButtons) FavoriteMarkup(count, userID int64, page int, hasMultiplePages, showDelete bool) *tg.InlineKeyboardMarkup {
	var rows []tg.KeyboardButtonRow

	// Play button
	rows = append(rows, tg.KeyboardButtonRow{
		Buttons: []tg.KeyboardButton{
			{Text: "Play Favorites ❤️", Data: []byte(fmt.Sprintf("myfavs|play|%d|0|0", userID))},
		},
	})

	// Delete buttons row (if enabled)
	if showDelete {
		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButton{
				{Text: "Delete All ❌", Data: []byte(fmt.Sprintf("delfavs|all|%d", userID))},
			},
		})
	}

	// Navigation row
	d := 0
	if showDelete {
		d = 1
	}

	if hasMultiplePages {
		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButton{
				{Text: "◂", Data: []byte(fmt.Sprintf("myfavs|prev|%d|%d|%d", userID, page, d))},
				{Text: "🗑", Data: []byte(fmt.Sprintf("myfavs|close|%d|%d|%d", userID, page, d))},
				{Text: "▸", Data: []byte(fmt.Sprintf("myfavs|next|%d|%d|%d", userID, page, d))},
			},
		})
	} else {
		rows = append(rows, tg.KeyboardButtonRow{
			Buttons: []tg.KeyboardButton{
				{Text: "🗑", Data: []byte(fmt.Sprintf("myfavs|close|%d|%d|%d", userID, page, d))},
			},
		})
	}

	return &tg.InlineKeyboardMarkup{Rows: rows}
}

// ActiveVCMarkup returns active voice chats navigation
func (mb *MakeButtons) ActiveVCMarkup(count, page int) *tg.InlineKeyboardMarkup {
	if count != 1 {
		return &tg.InlineKeyboardMarkup{
			Rows: []tg.KeyboardButtonRow{
				{
					Buttons: []tg.KeyboardButton{
						{Text: "◂", Data: []byte(fmt.Sprintf("activevc|prev|%d", page))},
						{Text: "🗑", Data: []byte("close")},
						{Text: "▸", Data: []byte(fmt.Sprintf("activevc|next|%d", page))},
					},
				},
			},
		}
	}
	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "🗑", Data: []byte("close")},
				},
			},
		},
	}
}

// AuthUsersMarkup returns authorized users navigation
func (mb *MakeButtons) AuthUsersMarkup(count, page int, randKey string) *tg.InlineKeyboardMarkup {
	if count != 1 {
		return &tg.InlineKeyboardMarkup{
			Rows: []tg.KeyboardButtonRow{
				{
					Buttons: []tg.KeyboardButton{
						{Text: "◂", Data: []byte(fmt.Sprintf("authus|prev|%d|%s", page, randKey))},
						{Text: "🗑", Data: []byte(fmt.Sprintf("authus|close|%d|%s", page, randKey))},
						{Text: "▸", Data: []byte(fmt.Sprintf("authus|next|%d|%s", page, randKey))},
					},
				},
			},
		}
	}
	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "🗑", Data: []byte(fmt.Sprintf("authus|close|%d|%s", page, randKey))},
				},
			},
		},
	}
}

// PlayerMarkup returns player control buttons
func (mb *MakeButtons) PlayerMarkup(chatID int64, videoID, username string) *tg.InlineKeyboardMarkup {
	if videoID == "telegram" {
		return &tg.InlineKeyboardMarkup{
			Rows: []tg.KeyboardButtonRow{
				{
					Buttons: []tg.KeyboardButton{
						{Text: "🎛️", Data: []byte(fmt.Sprintf("controls|%s|%d", videoID, chatID))},
						{Text: "🗑", Data: []byte("close")},
					},
				},
			},
		}
	}

	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "About Song", URL: fmt.Sprintf("https://t.me/%s?start=song_%s", username, videoID)},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "❤️", Data: []byte(fmt.Sprintf("add_favorite|%s", videoID))},
					{Text: "🎛️", Data: []byte(fmt.Sprintf("controls|%s|%d", videoID, chatID))},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "🗑", Data: []byte("close")},
				},
			},
		},
	}
}

// ControlsMarkup returns playback controls
func (mb *MakeButtons) ControlsMarkup(videoID string, chatID int64) *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "◂◂", Data: []byte(fmt.Sprintf("ctrl|bseek|%d", chatID))},
					{Text: "⏸", Data: []byte(fmt.Sprintf("ctrl|play|%d", chatID))},
					{Text: "▸▸", Data: []byte(fmt.Sprintf("ctrl|fseek|%d", chatID))},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "⏹ End", Data: []byte(fmt.Sprintf("ctrl|end|%d", chatID))},
					{Text: "↻ Replay", Data: []byte(fmt.Sprintf("ctrl|replay|%d", chatID))},
					{Text: "∞ Loop", Data: []byte(fmt.Sprintf("ctrl|loop|%d", chatID))},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "⏸ Mute", Data: []byte(fmt.Sprintf("ctrl|mute|%d", chatID))},
					{Text: "⏵ Unmute", Data: []byte(fmt.Sprintf("ctrl|unmute|%d", chatID))},
					{Text: "⏭ Skip", Data: []byte(fmt.Sprintf("ctrl|skip|%d", chatID))},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "🔙", Data: []byte(fmt.Sprintf("player|%s|%d", videoID, chatID))},
					{Text: "🗑", Data: []byte("close")},
				},
			},
		},
	}
}

// SongMarkup returns song download buttons
func (mb *MakeButtons) SongMarkup(randKey, url, key string) *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "Visit Youtube", URL: url},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "Audio", Data: []byte(fmt.Sprintf("song_dl|adl|%s|%s", key, randKey))},
					{Text: "Video", Data: []byte(fmt.Sprintf("song_dl|vdl|%s|%s", key, randKey))},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "◂", Data: []byte(fmt.Sprintf("song_dl|prev|%s|%s", key, randKey))},
					{Text: "▸", Data: []byte(fmt.Sprintf("song_dl|next|%s|%s", key, randKey))},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "🗑", Data: []byte(fmt.Sprintf("song_dl|close|%s|%s", key, randKey))},
				},
			},
		},
	}
}

// SongDetailsMarkup returns song details buttons
func (mb *MakeButtons) SongDetailsMarkup(url, channelURL string) *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "🎥", URL: url},
					{Text: "📺", URL: channelURL},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "🗑", Data: []byte("close")},
				},
			},
		},
	}
}

// SourceMarkup returns source code and support buttons
func (mb *MakeButtons) SourceMarkup() *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "Github ❤️", URL: "https://github.com/The-HellBot"},
					{Text: "Repo 📦", URL: "https://github.com/The-HellBot/Music"},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "Under HellBot Network { 🇮🇳 }", URL: "https://t.me/HellBot_Networks"},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "Support 🎙️", URL: "https://t.me/HellBot_Chats"},
					{Text: "Updates 📣", URL: "https://t.me/Its_HellBot"},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "🔙", Data: []byte("help|start")},
					{Text: "🗑", Data: []byte("close")},
				},
			},
		},
	}
}

// StartMarkup returns start button for groups
func (mb *MakeButtons) StartMarkup(username string) *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "Start Me 🎵", URL: fmt.Sprintf("https://t.me/%s?start=start", username)},
					{Text: "🗑", Data: []byte("close")},
				},
			},
		},
	}
}

// StartPMMarkup returns start menu buttons for PM
func (mb *MakeButtons) StartPMMarkup(username string) *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "Help ⚙️", Data: []byte("help|back")},
					{Text: "Source 📦", Data: []byte("source")},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "Add Me To Group 👥", URL: fmt.Sprintf("https://t.me/%s?startgroup=true", username)},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "🗑", Data: []byte("close")},
				},
			},
		},
	}
}

// HelpGCMarkup returns help button for groups
func (mb *MakeButtons) HelpGCMarkup(username string) *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "Get Help ❓", URL: fmt.Sprintf("https://t.me/%s?start=help", username)},
					{Text: "🗑", Data: []byte("close")},
				},
			},
		},
	}
}

// HelpPMMarkup returns help menu buttons
func (mb *MakeButtons) HelpPMMarkup() *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "➊ Admins", Data: []byte("help|admin")},
					{Text: "➋ Users", Data: []byte("help|user")},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "➌ Sudos", Data: []byte("help|sudo")},
					{Text: "➍ Others", Data: []byte("help|others")},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "➎ Owner", Data: []byte("help|owner")},
				},
			},
			{
				Buttons: []tg.KeyboardButton{
					{Text: "🔙", Data: []byte("help|start")},
					{Text: "🗑", Data: []byte("close")},
				},
			},
		},
	}
}

// HelpBack returns back button for help
func (mb *MakeButtons) HelpBack() *tg.InlineKeyboardMarkup {
	return &tg.InlineKeyboardMarkup{
		Rows: []tg.KeyboardButtonRow{
			{
				Buttons: []tg.KeyboardButton{
					{Text: "🔙", Data: []byte("help|back")},
					{Text: "🗑", Data: []byte("close")},
				},
			},
		},
	}
}

// Global instance
var Buttons = NewMakeButtons()
