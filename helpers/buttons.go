package helpers

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"
)

// MakeButtons handles creation of Telegram inline keyboards
type MakeButtons struct{}

// NewMakeButtons creates a new MakeButtons instance
func NewMakeButtons() *MakeButtons {
	return &MakeButtons{}
}

// CloseMarkup returns a close button
func (mb *MakeButtons) CloseMarkup() *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.Data("🗑", "close"),
		).
		Build()
}

// QueueMarkup returns queue navigation buttons
func (mb *MakeButtons) QueueMarkup(count, page int) *tg.ReplyInlineMarkup {
	if count != 1 {
		return tg.NewKeyboard().
			AddRow(
				tg.Button.Data("◂", fmt.Sprintf("queue|prev|%d", page)),
				tg.Button.Data("🗑", "close"),
				tg.Button.Data("▸", fmt.Sprintf("queue|next|%d", page)),
			).
			Build()
	}
	return tg.NewKeyboard().
		AddRow(
			tg.Button.Data("🗑", "close"),
		).
		Build()
}

// PlayFavsMarkup returns play favorites buttons
func (mb *MakeButtons) PlayFavsMarkup(userID int64) *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.Data("Audio", fmt.Sprintf("favsplay|audio|%d", userID)),
			tg.Button.Data("Video", fmt.Sprintf("favsplay|video|%d", userID)),
		).
		AddRow(
			tg.Button.Data("🗑", fmt.Sprintf("favsplay|close|%d", userID)),
		).
		Build()
}

// FavoriteMarkup returns favorites list with navigation
func (mb *MakeButtons) FavoriteMarkup(count, userID int64, page int, hasMultiplePages, showDelete bool) *tg.ReplyInlineMarkup {
	d := 0
	if showDelete {
		d = 1
	}

	kb := tg.NewKeyboard()

	// Play button
	kb.AddRow(
		tg.Button.Data("Play Favorites ❤️", fmt.Sprintf("myfavs|play|%d|0|0", userID)),
	)

	// Delete button
	if showDelete {
		kb.AddRow(
			tg.Button.Data("Delete All ❌", fmt.Sprintf("delfavs|all|%d", userID)),
		)
	}

	// Navigation
	if hasMultiplePages {
		kb.AddRow(
			tg.Button.Data("◂", fmt.Sprintf("myfavs|prev|%d|%d|%d", userID, page, d)),
			tg.Button.Data("🗑", fmt.Sprintf("myfavs|close|%d|%d|%d", userID, page, d)),
			tg.Button.Data("▸", fmt.Sprintf("myfavs|next|%d|%d|%d", userID, page, d)),
		)
	} else {
		kb.AddRow(
			tg.Button.Data("🗑", fmt.Sprintf("myfavs|close|%d|%d|%d", userID, page, d)),
		)
	}

	return kb.Build()
}

// ActiveVCMarkup returns active voice chats navigation
func (mb *MakeButtons) ActiveVCMarkup(count, page int) *tg.ReplyInlineMarkup {
	if count != 1 {
		return tg.NewKeyboard().
			AddRow(
				tg.Button.Data("◂", fmt.Sprintf("activevc|prev|%d", page)),
				tg.Button.Data("🗑", "close"),
				tg.Button.Data("▸", fmt.Sprintf("activevc|next|%d", page)),
			).
			Build()
	}
	return tg.NewKeyboard().
		AddRow(
			tg.Button.Data("🗑", "close"),
		).
		Build()
}

// AuthUsersMarkup returns authorized users navigation
func (mb *MakeButtons) AuthUsersMarkup(count, page int, randKey string) *tg.ReplyInlineMarkup {
	if count != 1 {
		return tg.NewKeyboard().
			AddRow(
				tg.Button.Data("◂", fmt.Sprintf("authus|prev|%d|%s", page, randKey)),
				tg.Button.Data("🗑", fmt.Sprintf("authus|close|%d|%s", page, randKey)),
				tg.Button.Data("▸", fmt.Sprintf("authus|next|%d|%s", page, randKey)),
			).
			Build()
	}
	return tg.NewKeyboard().
		AddRow(
			tg.Button.Data("🗑", fmt.Sprintf("authus|close|%d|%s", page, randKey)),
		).
		Build()
}

// PlayerMarkup returns player control buttons
func (mb *MakeButtons) PlayerMarkup(chatID int64, videoID, username string) *tg.ReplyInlineMarkup {
	if videoID == "telegram" {
		return tg.NewKeyboard().
			AddRow(
				tg.Button.Data("🎛️", fmt.Sprintf("controls|%s|%d", videoID, chatID)),
				tg.Button.Data("🗑", "close"),
			).
			Build()
	}

	return tg.NewKeyboard().
		AddRow(
			tg.Button.URL("About Song", fmt.Sprintf("https://t.me/%s?start=song_%s", username, videoID)),
		).
		AddRow(
			tg.Button.Data("❤️", fmt.Sprintf("add_favorite|%s", videoID)),
			tg.Button.Data("🎛️", fmt.Sprintf("controls|%s|%d", videoID, chatID)),
		).
		AddRow(
			tg.Button.Data("🗑", "close"),
		).
		Build()
}

// ControlsMarkup returns playback controls
func (mb *MakeButtons) ControlsMarkup(videoID string, chatID int64) *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.Data("◂◂", fmt.Sprintf("ctrl|bseek|%d", chatID)),
			tg.Button.Data("⏸", fmt.Sprintf("ctrl|play|%d", chatID)),
			tg.Button.Data("▸▸", fmt.Sprintf("ctrl|fseek|%d", chatID)),
		).
		AddRow(
			tg.Button.Data("⏹ End", fmt.Sprintf("ctrl|end|%d", chatID)),
			tg.Button.Data("↻ Replay", fmt.Sprintf("ctrl|replay|%d", chatID)),
			tg.Button.Data("∞ Loop", fmt.Sprintf("ctrl|loop|%d", chatID)),
		).
		AddRow(
			tg.Button.Data("🔇 Mute", fmt.Sprintf("ctrl|mute|%d", chatID)),
			tg.Button.Data("🔊 Unmute", fmt.Sprintf("ctrl|unmute|%d", chatID)),
			tg.Button.Data("⏭ Skip", fmt.Sprintf("ctrl|skip|%d", chatID)),
		).
		AddRow(
			tg.Button.Data("🔙", fmt.Sprintf("player|%s|%d", videoID, chatID)),
			tg.Button.Data("🗑", "close"),
		).
		Build()
}

// SongMarkup returns song download buttons
func (mb *MakeButtons) SongMarkup(randKey, url, key string) *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.URL("Visit Youtube", url),
		).
		AddRow(
			tg.Button.Data("Audio", fmt.Sprintf("song_dl|adl|%s|%s", key, randKey)),
			tg.Button.Data("Video", fmt.Sprintf("song_dl|vdl|%s|%s", key, randKey)),
		).
		AddRow(
			tg.Button.Data("◂", fmt.Sprintf("song_dl|prev|%s|%s", key, randKey)),
			tg.Button.Data("▸", fmt.Sprintf("song_dl|next|%s|%s", key, randKey)),
		).
		AddRow(
			tg.Button.Data("🗑", fmt.Sprintf("song_dl|close|%s|%s", key, randKey)),
		).
		Build()
}

// SongDetailsMarkup returns song details buttons
func (mb *MakeButtons) SongDetailsMarkup(url, channelURL string) *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.URL("🎥", url),
			tg.Button.URL("📺", channelURL),
		).
		AddRow(
			tg.Button.Data("🗑", "close"),
		).
		Build()
}

// SourceMarkup returns source code and support buttons
func (mb *MakeButtons) SourceMarkup() *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.URL("Github ❤️", "https://github.com/The-HellBot"),
			tg.Button.URL("Repo 📦", "https://github.com/The-HellBot/Music"),
		).
		AddRow(
			tg.Button.URL("Under HellBot Network { 🇮🇳 }", "https://t.me/HellBot_Networks"),
		).
		AddRow(
			tg.Button.URL("Support 🎙️", "https://t.me/HellBot_Chats"),
			tg.Button.URL("Updates 📣", "https://t.me/Its_HellBot"),
		).
		AddRow(
			tg.Button.Data("🔙", "help|start"),
			tg.Button.Data("🗑", "close"),
		).
		Build()
}

// StartMarkup returns start button for groups
func (mb *MakeButtons) StartMarkup(username string) *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.URL("Start Me 🎵", fmt.Sprintf("https://t.me/%s?start=start", username)),
			tg.Button.Data("🗑", "close"),
		).
		Build()
}

// StartPMMarkup returns start menu buttons for PM
func (mb *MakeButtons) StartPMMarkup(username string) *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.Data("Help ⚙️", "help|back"),
			tg.Button.Data("Source 📦", "source"),
		).
		AddRow(
			tg.Button.URL("Add Me To Group 👥", fmt.Sprintf("https://t.me/%s?startgroup=true", username)),
		).
		AddRow(
			tg.Button.Data("🗑", "close"),
		).
		Build()
}

// HelpGCMarkup returns help button for groups
func (mb *MakeButtons) HelpGCMarkup(username string) *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.URL("Get Help ❓", fmt.Sprintf("https://t.me/%s?start=help", username)),
			tg.Button.Data("🗑", "close"),
		).
		Build()
}

// HelpPMMarkup returns help menu buttons
func (mb *MakeButtons) HelpPMMarkup() *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.Data("➊ Admins", "help|admin"),
			tg.Button.Data("➋ Users", "help|user"),
		).
		AddRow(
			tg.Button.Data("➌ Sudos", "help|sudo"),
			tg.Button.Data("➍ Others", "help|others"),
		).
		AddRow(
			tg.Button.Data("➎ Owner", "help|owner"),
		).
		AddRow(
			tg.Button.Data("🔙", "help|start"),
			tg.Button.Data("🗑", "close"),
		).
		Build()
}

// HelpBack returns back button for help
func (mb *MakeButtons) HelpBack() *tg.ReplyInlineMarkup {
	return tg.NewKeyboard().
		AddRow(
			tg.Button.Data("🔙", "help|back"),
			tg.Button.Data("🗑", "close"),
		).
		Build()
}

// Global instance
var Buttons = NewMakeButtons()
