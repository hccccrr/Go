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

// helper: callback button
func cbBtn(text, data string) tg.KeyboardButtonObj {
	return tg.KeyboardButtonObj{
		Text: text,
		Data: []byte(data),
	}
}

// helper: url button
func urlBtn(text, url string) tg.KeyboardButtonUrl {
	return tg.KeyboardButtonUrl{
		Text: text,
		Url:  url,
	}
}

// helper: build inline markup from rows of buttons
func buildMarkup(rows ...[]tg.KeyboardButton) *tg.ReplyInlineMarkup {
	var keyRows []tg.KeyboardButtonRow
	for _, row := range rows {
		keyRows = append(keyRows, tg.KeyboardButtonRow{Buttons: row})
	}
	return &tg.ReplyInlineMarkup{Rows: keyRows}
}

// CloseMarkup returns a close button
func (mb *MakeButtons) CloseMarkup() *tg.ReplyInlineMarkup {
	return buildMarkup(
		[]tg.KeyboardButton{cbBtn("🗑", "close")},
	)
}

// QueueMarkup returns queue navigation buttons
func (mb *MakeButtons) QueueMarkup(count, page int) *tg.ReplyInlineMarkup {
	if count != 1 {
		return buildMarkup(
			[]tg.KeyboardButton{
				cbBtn("◂", fmt.Sprintf("queue|prev|%d", page)),
				cbBtn("🗑", "close"),
				cbBtn("▸", fmt.Sprintf("queue|next|%d", page)),
			},
		)
	}
	return buildMarkup(
		[]tg.KeyboardButton{cbBtn("🗑", "close")},
	)
}

// PlayFavsMarkup returns play favorites buttons
func (mb *MakeButtons) PlayFavsMarkup(userID int64) *tg.ReplyInlineMarkup {
	return buildMarkup(
		[]tg.KeyboardButton{
			cbBtn("Audio", fmt.Sprintf("favsplay|audio|%d", userID)),
			cbBtn("Video", fmt.Sprintf("favsplay|video|%d", userID)),
		},
		[]tg.KeyboardButton{
			cbBtn("🗑", fmt.Sprintf("favsplay|close|%d", userID)),
		},
	)
}

// FavoriteMarkup returns favorites list with navigation
func (mb *MakeButtons) FavoriteMarkup(count, userID int64, page int, hasMultiplePages, showDelete bool) *tg.ReplyInlineMarkup {
	d := 0
	if showDelete {
		d = 1
	}

	var rows [][]tg.KeyboardButton

	// Play button
	rows = append(rows, []tg.KeyboardButton{
		cbBtn("Play Favorites ❤️", fmt.Sprintf("myfavs|play|%d|0|0", userID)),
	})

	// Delete button
	if showDelete {
		rows = append(rows, []tg.KeyboardButton{
			cbBtn("Delete All ❌", fmt.Sprintf("delfavs|all|%d", userID)),
		})
	}

	// Navigation
	if hasMultiplePages {
		rows = append(rows, []tg.KeyboardButton{
			cbBtn("◂", fmt.Sprintf("myfavs|prev|%d|%d|%d", userID, page, d)),
			cbBtn("🗑", fmt.Sprintf("myfavs|close|%d|%d|%d", userID, page, d)),
			cbBtn("▸", fmt.Sprintf("myfavs|next|%d|%d|%d", userID, page, d)),
		})
	} else {
		rows = append(rows, []tg.KeyboardButton{
			cbBtn("🗑", fmt.Sprintf("myfavs|close|%d|%d|%d", userID, page, d)),
		})
	}

	return buildMarkup(rows...)
}

// ActiveVCMarkup returns active voice chats navigation
func (mb *MakeButtons) ActiveVCMarkup(count, page int) *tg.ReplyInlineMarkup {
	if count != 1 {
		return buildMarkup(
			[]tg.KeyboardButton{
				cbBtn("◂", fmt.Sprintf("activevc|prev|%d", page)),
				cbBtn("🗑", "close"),
				cbBtn("▸", fmt.Sprintf("activevc|next|%d", page)),
			},
		)
	}
	return buildMarkup(
		[]tg.KeyboardButton{cbBtn("🗑", "close")},
	)
}

// AuthUsersMarkup returns authorized users navigation
func (mb *MakeButtons) AuthUsersMarkup(count, page int, randKey string) *tg.ReplyInlineMarkup {
	if count != 1 {
		return buildMarkup(
			[]tg.KeyboardButton{
				cbBtn("◂", fmt.Sprintf("authus|prev|%d|%s", page, randKey)),
				cbBtn("🗑", fmt.Sprintf("authus|close|%d|%s", page, randKey)),
				cbBtn("▸", fmt.Sprintf("authus|next|%d|%s", page, randKey)),
			},
		)
	}
	return buildMarkup(
		[]tg.KeyboardButton{
			cbBtn("🗑", fmt.Sprintf("authus|close|%d|%s", page, randKey)),
		},
	)
}

// PlayerMarkup returns player control buttons
func (mb *MakeButtons) PlayerMarkup(chatID int64, videoID, username string) *tg.ReplyInlineMarkup {
	if videoID == "telegram" {
		return buildMarkup(
			[]tg.KeyboardButton{
				cbBtn("🎛️", fmt.Sprintf("controls|%s|%d", videoID, chatID)),
				cbBtn("🗑", "close"),
			},
		)
	}

	return buildMarkup(
		[]tg.KeyboardButton{
			urlBtn("About Song", fmt.Sprintf("https://t.me/%s?start=song_%s", username, videoID)),
		},
		[]tg.KeyboardButton{
			cbBtn("❤️", fmt.Sprintf("add_favorite|%s", videoID)),
			cbBtn("🎛️", fmt.Sprintf("controls|%s|%d", videoID, chatID)),
		},
		[]tg.KeyboardButton{
			cbBtn("🗑", "close"),
		},
	)
}

// ControlsMarkup returns playback controls
func (mb *MakeButtons) ControlsMarkup(videoID string, chatID int64) *tg.ReplyInlineMarkup {
	return buildMarkup(
		[]tg.KeyboardButton{
			cbBtn("◂◂", fmt.Sprintf("ctrl|bseek|%d", chatID)),
			cbBtn("⸸", fmt.Sprintf("ctrl|play|%d", chatID)),
			cbBtn("▸▸", fmt.Sprintf("ctrl|fseek|%d", chatID)),
		},
		[]tg.KeyboardButton{
			cbBtn("⏹ End", fmt.Sprintf("ctrl|end|%d", chatID)),
			cbBtn("↻ Replay", fmt.Sprintf("ctrl|replay|%d", chatID)),
			cbBtn("∞ Loop", fmt.Sprintf("ctrl|loop|%d", chatID)),
		},
		[]tg.KeyboardButton{
			cbBtn("⸸ Mute", fmt.Sprintf("ctrl|mute|%d", chatID)),
			cbBtn("↵ Unmute", fmt.Sprintf("ctrl|unmute|%d", chatID)),
			cbBtn("⭭ Skip", fmt.Sprintf("ctrl|skip|%d", chatID)),
		},
		[]tg.KeyboardButton{
			cbBtn("🔙", fmt.Sprintf("player|%s|%d", videoID, chatID)),
			cbBtn("🗑", "close"),
		},
	)
}

// SongMarkup returns song download buttons
func (mb *MakeButtons) SongMarkup(randKey, url, key string) *tg.ReplyInlineMarkup {
	return buildMarkup(
		[]tg.KeyboardButton{
			urlBtn("Visit Youtube", url),
		},
		[]tg.KeyboardButton{
			cbBtn("Audio", fmt.Sprintf("song_dl|adl|%s|%s", key, randKey)),
			cbBtn("Video", fmt.Sprintf("song_dl|vdl|%s|%s", key, randKey)),
		},
		[]tg.KeyboardButton{
			cbBtn("◂", fmt.Sprintf("song_dl|prev|%s|%s", key, randKey)),
			cbBtn("▸", fmt.Sprintf("song_dl|next|%s|%s", key, randKey)),
		},
		[]tg.KeyboardButton{
			cbBtn("🗑", fmt.Sprintf("song_dl|close|%s|%s", key, randKey)),
		},
	)
}

// SongDetailsMarkup returns song details buttons
func (mb *MakeButtons) SongDetailsMarkup(url, channelURL string) *tg.ReplyInlineMarkup {
	return buildMarkup(
		[]tg.KeyboardButton{
			urlBtn("🎥", url),
			urlBtn("📺", channelURL),
		},
		[]tg.KeyboardButton{
			cbBtn("🗑", "close"),
		},
	)
}

// SourceMarkup returns source code and support buttons
func (mb *MakeButtons) SourceMarkup() *tg.ReplyInlineMarkup {
	return buildMarkup(
		[]tg.KeyboardButton{
			urlBtn("Github ❤️", "https://github.com/The-HellBot"),
			urlBtn("Repo 📦", "https://github.com/The-HellBot/Music"),
		},
		[]tg.KeyboardButton{
			urlBtn("Under HellBot Network { 🇮🇳 }", "https://t.me/HellBot_Networks"),
		},
		[]tg.KeyboardButton{
			urlBtn("Support 🎙️", "https://t.me/HellBot_Chats"),
			urlBtn("Updates 📣", "https://t.me/Its_HellBot"),
		},
		[]tg.KeyboardButton{
			cbBtn("🔙", "help|start"),
			cbBtn("🗑", "close"),
		},
	)
}

// StartMarkup returns start button for groups
func (mb *MakeButtons) StartMarkup(username string) *tg.ReplyInlineMarkup {
	return buildMarkup(
		[]tg.KeyboardButton{
			urlBtn("Start Me 🎵", fmt.Sprintf("https://t.me/%s?start=start", username)),
			cbBtn("🗑", "close"),
		},
	)
}

// StartPMMarkup returns start menu buttons for PM
func (mb *MakeButtons) StartPMMarkup(username string) *tg.ReplyInlineMarkup {
	return buildMarkup(
		[]tg.KeyboardButton{
			cbBtn("Help ⚙️", "help|back"),
			cbBtn("Source 📦", "source"),
		},
		[]tg.KeyboardButton{
			urlBtn("Add Me To Group 👥", fmt.Sprintf("https://t.me/%s?startgroup=true", username)),
		},
		[]tg.KeyboardButton{
			cbBtn("🗑", "close"),
		},
	)
}

// HelpGCMarkup returns help button for groups
func (mb *MakeButtons) HelpGCMarkup(username string) *tg.ReplyInlineMarkup {
	return buildMarkup(
		[]tg.KeyboardButton{
			urlBtn("Get Help ❓", fmt.Sprintf("https://t.me/%s?start=help", username)),
			cbBtn("🗑", "close"),
		},
	)
}

// HelpPMMarkup returns help menu buttons
func (mb *MakeButtons) HelpPMMarkup() *tg.ReplyInlineMarkup {
	return buildMarkup(
		[]tg.KeyboardButton{
			cbBtn("➊ Admins", "help|admin"),
			cbBtn("➋ Users", "help|user"),
		},
		[]tg.KeyboardButton{
			cbBtn("➌ Sudos", "help|sudo"),
			cbBtn("➍ Others", "help|others"),
		},
		[]tg.KeyboardButton{
			cbBtn("➎ Owner", "help|owner"),
		},
		[]tg.KeyboardButton{
			cbBtn("🔙", "help|start"),
			cbBtn("🗑", "close"),
		},
	)
}

// HelpBack returns back button for help
func (mb *MakeButtons) HelpBack() *tg.ReplyInlineMarkup {
	return buildMarkup(
		[]tg.KeyboardButton{
			cbBtn("🔙", "help|back"),
			cbBtn("🗑", "close"),
		},
	)
}

// Global instance
var Buttons = NewMakeButtons()
