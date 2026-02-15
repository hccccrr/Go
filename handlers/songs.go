package handlers

import (
	"fmt"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/utils"
	"shizumusic/helpers"
)

// TEXTS templates from utils
var TEXTS = helpers.TextTemplates

// RegisterSongsHandlers registers song command handlers
func RegisterSongsHandlers(client *core.Client) {
	// Song search
	client.BotClient.AddMessageHandler("/song", func(m *tg.NewMessage) error {
		return core.UserOnly(handleSong(client))(m)
	})

	// Lyrics
	client.BotClient.AddMessageHandler("/lyrics", func(m *tg.NewMessage) error {
		return core.UserOnly(handleLyrics(client))(m)
	})
}

func handleSong(client *core.Client) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		parts := strings.SplitN(m.Text, " ", 2)
		if len(parts) == 1 {
			return m.Reply("Nothing given to search.")
		}

		query := strings.TrimSpace(parts[1])
		msg, _ := m.Reply(fmt.Sprintf(
			"<b><i>Searching</i></b> \"%s\" ...",
			query,
		))

		// TODO: Search YouTube and return results
		// For now, placeholder response
		msg.Edit(fmt.Sprintf(
			"**🎵 Song Search**\n\n"+
				"**Query:** `%s`\n\n"+
				"Feature coming soon! Will search YouTube and provide download options.",
			query,
		))

		return nil
	}
}

func handleLyrics(client *core.Client) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		// Check if lyrics API is configured
		if config.Cfg.LyricsAPI == "" {
			return m.Reply("Lyrics module is disabled!")
		}

		parts := strings.SplitN(m.Text, " ", 2)
		if len(parts) != 2 {
			return m.Reply(
				"__Nothing given to search.__\n" +
					"Example: `/lyrics loose yourself - eminem`",
			)
		}

		input := strings.TrimSpace(parts[1])
		query := strings.Split(input, "-")

		var song, artist string
		if len(query) == 2 {
			song = strings.TrimSpace(query[0])
			artist = strings.TrimSpace(query[1])
		} else {
			song = strings.TrimSpace(query[0])
			artist = ""
		}

		text := fmt.Sprintf("**Searching lyrics...**\n\n__Song:__ `%s`", song)
		if artist != "" {
			text += fmt.Sprintf("\n__Artist:__ `%s`", artist)
		}

		msg, _ := m.Reply(text)

		// TODO: Fetch lyrics from API
		// For now, placeholder
		msg.Edit(fmt.Sprintf(
			"**🎤 Lyrics Search**\n\n"+
				"**Song:** `%s`\n"+
				"**Artist:** `%s`\n\n"+
				"Feature coming soon! Will fetch lyrics from API.",
			song,
			artist,
		))

		return nil
	}
}
