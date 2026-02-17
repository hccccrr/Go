package handlers

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/utils"
	"shizumusic/helpers"
)

// TEXTS templates from utils
var TEXTS = helpers.TextTemplates

// RegisterFavoritesHandlers registers favorites command handlers
func RegisterFavoritesHandlers(client *core.Client, db *core.Database) {
	// Favorites list
	client.BotClient.AddMessageHandler("/favs", func(m *tg.NewMessage) error {
		return core.UserOnly(handleFavorites(db))(m)
	})
	
	client.BotClient.AddMessageHandler("/myfavs", func(m *tg.NewMessage) error {
		return core.UserOnly(handleFavorites(db))(m)
	})
	
	client.BotClient.AddMessageHandler("/favorites", func(m *tg.NewMessage) error {
		return core.UserOnly(handleFavorites(db))(m)
	})

	// Delete favorites
	client.BotClient.AddMessageHandler("/delfavs", func(m *tg.NewMessage) error {
		return core.UserOnly(handleDeleteFavorites(db))(m)
	})
}

func handleFavorites(db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		msg, _ := m.Reply("Fetching favorites...")

		// TODO: Get favorites from database
		// favs := db.GetAllFavorites(m.From.ID)
		
		// For now, placeholder
		text := fmt.Sprintf(
			"**❤️ Your Favorites**\n\n"+
				"**User:** [%s](tg://user?id=%d)\n"+
				"**Total:** `0`\n\n"+
				"You don't have any favorite tracks added yet!\n\n"+
				"Use the ⭐ button on songs to add them to favorites.",
			m.From.FirstName,
			m.From.ID,
		)

		msg.Edit(text)
		return nil
	}
}

func handleDeleteFavorites(db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		msg, _ := m.Reply("Fetching favorites...")

		// TODO: Get favorites and show delete options
		text := fmt.Sprintf(
			"**🗑️ Delete Favorites**\n\n"+
				"**User:** [%s](tg://user?id=%d)\n\n"+
				"You don't have any favorites to delete!",
			m.From.FirstName,
			m.From.ID,
		)

		msg.Edit(text)
		return nil
	}
}
