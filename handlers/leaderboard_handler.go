package handlers

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/utils"
	"shizumusic/helpers"
)

// RegisterLeaderboardHandlers registers leaderboard command handlers
func RegisterLeaderboardHandlers(client *core.Client, db *core.Database) {
	// Leaderboard command
	client.BotClient.AddMessageHandler("/leaderboard", func(m *tg.NewMessage) error {
		return handleLeaderboard(client, db)(m)
	})
	
	client.BotClient.AddMessageHandler("/topusers", func(m *tg.NewMessage) error {
		return handleLeaderboard(client, db)(m)
	})
}

func handleLeaderboard(client *core.Client, db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		msg, _ := m.Reply("Fetching top users...")

		// Get bot details
		me, _ := client.BotClient.GetMe()
		botMention := fmt.Sprintf("@%s", me.Username)

		// Get top 10 users by songs
		topSongs, err := getTopUsersBySongs(db, 10)
		if err != nil {
			return msg.Edit("Failed to fetch leaderboard data!")
		}

		// Get top 10 users by messages
		topMessages, err := getTopUsersByMessages(db, 10)
		if err != nil {
			return msg.Edit("Failed to fetch leaderboard data!")
		}

		// Build songs leaderboard text
		songsText := fmt.Sprintf("**🎵 Top 10 Music Lovers of %s**\n\n", botMention)
		for i, user := range topSongs {
			index := i + 1
			link := fmt.Sprintf("https://t.me/%s?start=user_%d", me.Username, user.ID)
			indexStr := fmt.Sprintf("%02d", index)
			songsText += fmt.Sprintf("**• %s:** [%s](%s) - **%d** songs\n", indexStr, user.Username, link, user.Songs)
		}
		songsText += "\n**🎧 Keep streaming! Enjoy the music!**"

		// Build messages leaderboard text
		messagesText := fmt.Sprintf("\n\n**💬 Top 10 Active Chatters of %s**\n\n", botMention)
		for i, user := range topMessages {
			index := i + 1
			link := fmt.Sprintf("https://t.me/%s?start=user_%d", me.Username, user.ID)
			indexStr := fmt.Sprintf("%02d", index)
			messagesText += fmt.Sprintf("**• %s:** [%s](%s) - **%d** messages\n", indexStr, user.Username, link, user.Messages)
		}
		messagesText += "\n**💬 Keep chatting! Stay active!**"

		// Combine both
		fullText := songsText + "\n\n" + "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" + messagesText

		msg.Edit(fullText)
		return nil
	}
}

// Helper functions

func getTopUsersBySongs(db *core.Database, limit int) ([]utils.UserStats, error) {
	// TODO: Implement database query
	// For now, return placeholder
	return []utils.UserStats{}, nil
}

func getTopUsersByMessages(db *core.Database, limit int) ([]utils.UserStats, error) {
	// TODO: Implement database query
	// For now, return placeholder
	return []utils.UserStats{}, nil
}
