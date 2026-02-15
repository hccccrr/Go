package helpers

import (
	"fmt"

	"github.com/celestix/gotgproto/ext"
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
func (mb *MakeButtons) CloseMarkup() [][]ext.InlineKeyboardButton {
	return [][]ext.InlineKeyboardButton{
		{
			{Text: "🗑", Data: "close"},
		},
	}
}

// QueueMarkup returns queue navigation buttons
func (mb *MakeButtons) QueueMarkup(count, page int) [][]ext.InlineKeyboardButton {
	if count != 1 {
		return [][]ext.InlineKeyboardButton{
			{
				{Text: "◂", Data: fmt.Sprintf("queue|prev|%d", page)},
				{Text: "🗑", Data: "close"},
				{Text: "▸", Data: fmt.Sprintf("queue|next|%d", page)},
			},
		}
	}
	return [][]ext.InlineKeyboardButton{
		{
			{Text: "🗑", Data: "close"},
		},
	}
}

// PlayFavsMarkup returns play favorites buttons
func (mb *MakeButtons) PlayFavsMarkup(userID int64) [][]ext.InlineKeyboardButton {
	return [][]ext.InlineKeyboardButton{
		{
			{Text: "Audio", Data: fmt.Sprintf("favsplay|audio|%d", userID)},
			{Text: "Video", Data: fmt.Sprintf("favsplay|video|%d", userID)},
		},
		{
			{Text: "🗑", Data: fmt.Sprintf("favsplay|close|%d", userID)},
		},
	}
}

// FavoriteMarkup returns favorites list with navigation
func (mb *MakeButtons) FavoriteMarkup(count, userID int64, page int, hasMultiplePages, showDelete bool) [][]ext.InlineKeyboardButton {
	var buttons [][]ext.InlineKeyboardButton

	// Play button
	playRow := []ext.InlineKeyboardButton{
		{Text: "Play Favorites ❤️", Data: fmt.Sprintf("myfavs|play|%d|0|0", userID)},
	}
	buttons = append(buttons, playRow)

	// Delete buttons row (if enabled)
	if showDelete {
		// This would be populated with numbered buttons based on favorites
		// For now, just add "Delete All" button
		deleteRow := []ext.InlineKeyboardButton{
			{Text: "Delete All ❌", Data: fmt.Sprintf("delfavs|all|%d", userID)},
		}
		buttons = append(buttons, deleteRow)
	}

	// Navigation row
	d := 0
	if showDelete {
		d = 1
	}

	if hasMultiplePages {
		navRow := []ext.InlineKeyboardButton{
			{Text: "◂", Data: fmt.Sprintf("myfavs|prev|%d|%d|%d", userID, page, d)},
			{Text: "🗑", Data: fmt.Sprintf("myfavs|close|%d|%d|%d", userID, page, d)},
			{Text: "▸", Data: fmt.Sprintf("myfavs|next|%d|%d|%d", userID, page, d)},
		}
		buttons = append(buttons, navRow)
	} else {
		navRow := []ext.InlineKeyboardButton{
			{Text: "🗑", Data: fmt.Sprintf("myfavs|close|%d|%d|%d", userID, page, d)},
		}
		buttons = append(buttons, navRow)
	}

	return buttons
}

// ActiveVCMarkup returns active voice chats navigation
func (mb *MakeButtons) ActiveVCMarkup(count, page int) [][]ext.InlineKeyboardButton {
	if count != 1 {
		return [][]ext.InlineKeyboardButton{
			{
				{Text: "◂", Data: fmt.Sprintf("activevc|prev|%d", page)},
				{Text: "🗑", Data: "close"},
				{Text: "▸", Data: fmt.Sprintf("activevc|next|%d", page)},
			},
		}
	}
	return [][]ext.InlineKeyboardButton{
		{
			{Text: "🗑", Data: "close"},
		},
	}
}

// AuthUsersMarkup returns authorized users navigation
func (mb *MakeButtons) AuthUsersMarkup(count, page int, randKey string) [][]ext.InlineKeyboardButton {
	if count != 1 {
		return [][]ext.InlineKeyboardButton{
			{
				{Text: "◂", Data: fmt.Sprintf("authus|prev|%d|%s", page, randKey)},
				{Text: "🗑", Data: fmt.Sprintf("authus|close|%d|%s", page, randKey)},
				{Text: "▸", Data: fmt.Sprintf("authus|next|%d|%s", page, randKey)},
			},
		}
	}
	return [][]ext.InlineKeyboardButton{
		{
			{Text: "🗑", Data: fmt.Sprintf("authus|close|%d|%s", page, randKey)},
		},
	}
}

// PlayerMarkup returns player control buttons
func (mb *MakeButtons) PlayerMarkup(chatID int64, videoID, username string) [][]ext.InlineKeyboardButton {
	if videoID == "telegram" {
		return [][]ext.InlineKeyboardButton{
			{
				{Text: "🎛️", Data: fmt.Sprintf("controls|%s|%d", videoID, chatID)},
				{Text: "🗑", Data: "close"},
			},
		}
	}

	return [][]ext.InlineKeyboardButton{
		{
			{Text: "About Song", URL: fmt.Sprintf("https://t.me/%s?start=song_%s", username, videoID)},
		},
		{
			{Text: "❤️", Data: fmt.Sprintf("add_favorite|%s", videoID)},
			{Text: "🎛️", Data: fmt.Sprintf("controls|%s|%d", videoID, chatID)},
		},
		{
			{Text: "🗑", Data: "close"},
		},
	}
}

// ControlsMarkup returns playback controls
func (mb *MakeButtons) ControlsMarkup(videoID string, chatID int64) [][]ext.InlineKeyboardButton {
	return [][]ext.InlineKeyboardButton{
		{
			{Text: "◂◂", Data: fmt.Sprintf("ctrl|bseek|%d", chatID)},
			{Text: "⏸", Data: fmt.Sprintf("ctrl|play|%d", chatID)},
			{Text: "▸▸", Data: fmt.Sprintf("ctrl|fseek|%d", chatID)},
		},
		{
			{Text: "⏹ End", Data: fmt.Sprintf("ctrl|end|%d", chatID)},
			{Text: "↻ Replay", Data: fmt.Sprintf("ctrl|replay|%d", chatID)},
			{Text: "∞ Loop", Data: fmt.Sprintf("ctrl|loop|%d", chatID)},
		},
		{
			{Text: "⏸ Mute", Data: fmt.Sprintf("ctrl|mute|%d", chatID)},
			{Text: "⏵ Unmute", Data: fmt.Sprintf("ctrl|unmute|%d", chatID)},
			{Text: "⏭ Skip", Data: fmt.Sprintf("ctrl|skip|%d", chatID)},
		},
		{
			{Text: "🔙", Data: fmt.Sprintf("player|%s|%d", videoID, chatID)},
			{Text: "🗑", Data: "close"},
		},
	}
}

// SongMarkup returns song download buttons
func (mb *MakeButtons) SongMarkup(randKey, url, key string) [][]ext.InlineKeyboardButton {
	return [][]ext.InlineKeyboardButton{
		{
			{Text: "Visit Youtube", URL: url},
		},
		{
			{Text: "Audio", Data: fmt.Sprintf("song_dl|adl|%s|%s", key, randKey)},
			{Text: "Video", Data: fmt.Sprintf("song_dl|vdl|%s|%s", key, randKey)},
		},
		{
			{Text: "◂", Data: fmt.Sprintf("song_dl|prev|%s|%s", key, randKey)},
			{Text: "▸", Data: fmt.Sprintf("song_dl|next|%s|%s", key, randKey)},
		},
		{
			{Text: "🗑", Data: fmt.Sprintf("song_dl|close|%s|%s", key, randKey)},
		},
	}
}

// SongDetailsMarkup returns song details buttons
func (mb *MakeButtons) SongDetailsMarkup(url, channelURL string) [][]ext.InlineKeyboardButton {
	return [][]ext.InlineKeyboardButton{
		{
			{Text: "🎥", URL: url},
			{Text: "📺", URL: channelURL},
		},
		{
			{Text: "🗑", Data: "close"},
		},
	}
}

// SourceMarkup returns source code and support buttons
func (mb *MakeButtons) SourceMarkup() [][]ext.InlineKeyboardButton {
	return [][]ext.InlineKeyboardButton{
		{
			{Text: "Github ❤️", URL: "https://github.com/The-HellBot"},
			{Text: "Repo 📦", URL: "https://github.com/The-HellBot/Music"},
		},
		{
			{Text: "Under HellBot Network { 🇮🇳 }", URL: "https://t.me/HellBot_Networks"},
		},
		{
			{Text: "Support 🎙️", URL: "https://t.me/HellBot_Chats"},
			{Text: "Updates 📣", URL: "https://t.me/Its_HellBot"},
		},
		{
			{Text: "🔙", Data: "help|start"},
			{Text: "🗑", Data: "close"},
		},
	}
}

// StartMarkup returns start button for groups
func (mb *MakeButtons) StartMarkup(username string) [][]ext.InlineKeyboardButton {
	return [][]ext.InlineKeyboardButton{
		{
			{Text: "Start Me 🎵", URL: fmt.Sprintf("https://t.me/%s?start=start", username)},
			{Text: "🗑", Data: "close"},
		},
	}
}

// StartPMMarkup returns start menu buttons for PM
func (mb *MakeButtons) StartPMMarkup(username string) [][]ext.InlineKeyboardButton {
	return [][]ext.InlineKeyboardButton{
		{
			{Text: "Help ⚙️", Data: "help|back"},
			{Text: "Source 📦", Data: "source"},
		},
		{
			{Text: "Add Me To Group 👥", URL: fmt.Sprintf("https://t.me/%s?startgroup=true", username)},
		},
		{
			{Text: "🗑", Data: "close"},
		},
	}
}

// HelpGCMarkup returns help button for groups
func (mb *MakeButtons) HelpGCMarkup(username string) [][]ext.InlineKeyboardButton {
	return [][]ext.InlineKeyboardButton{
		{
			{Text: "Get Help ❓", URL: fmt.Sprintf("https://t.me/%s?start=help", username)},
			{Text: "🗑", Data: "close"},
		},
	}
}

// HelpPMMarkup returns help menu buttons
func (mb *MakeButtons) HelpPMMarkup() [][]ext.InlineKeyboardButton {
	return [][]ext.InlineKeyboardButton{
		{
			{Text: "➊ Admins", Data: "help|admin"},
			{Text: "➋ Users", Data: "help|user"},
		},
		{
			{Text: "➌ Sudos", Data: "help|sudo"},
			{Text: "➍ Others", Data: "help|others"},
		},
		{
			{Text: "➎ Owner", Data: "help|owner"},
		},
		{
			{Text: "🔙", Data: "help|start"},
			{Text: "🗑", Data: "close"},
		},
	}
}

// HelpBack returns back button for help
func (mb *MakeButtons) HelpBack() [][]ext.InlineKeyboardButton {
	return [][]ext.InlineKeyboardButton{
		{
			{Text: "🔙", Data: "help|back"},
			{Text: "🗑", Data: "close"},
		},
	}
}

// Global instance
var Buttons = NewMakeButtons()
