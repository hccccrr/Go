package handlers

import (
	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/core"
	"shizumusic/utils"
	"shizumusic/helpers"
)

// TEXTS templates from utils
var TEXTS = helpers.TextTemplates

func RegisterUserHandlers(client *core.Client, db *core.Database) {
	// Profile/Me
	client.BotClient.AddMessageHandler("/me", handleProfile(db))
	client.BotClient.AddMessageHandler("/profile", handleProfile(db))
	
	// Stats
	client.BotClient.AddMessageHandler("/stats", handleStats(db))
	
	// Leaderboard
	client.BotClient.AddMessageHandler("/leaderboard", handleLeaderboard(db))
	client.BotClient.AddMessageHandler("/topusers", handleLeaderboard(db))
}

func handleProfile(db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		m.Reply("**👤 Your Profile**\n\nFeature coming soon!")
		return nil
	}
}

func handleStats(db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		m.Reply("**📊 Bot Statistics**\n\nFeature coming soon!")
		return nil
	}
}

func handleLeaderboard(db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		m.Reply("**🏆 Leaderboard**\n\nFeature coming soon!")
		return nil
	}
}
