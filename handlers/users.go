package handlers

import (
	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/core"
)

func RegisterUserHandlers(client *core.Client, db *core.Database) {
	// Profile/Me
	client.BotClient.AddMessageHandler("/me", handleProfile(db))
	client.BotClient.AddMessageHandler("/profile", handleProfile(db))

	// Stats
	client.BotClient.AddMessageHandler("/stats", handleStats(db))

	// Leaderboard - handled by leaderboard_handler.go
	client.BotClient.AddMessageHandler("/leaderboard", handleLeaderboard(client, db))
	client.BotClient.AddMessageHandler("/topusers", handleLeaderboard(client, db))
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
