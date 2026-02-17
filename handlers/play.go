package handlers

import (
	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/core"
)

func RegisterPlayHandlers(client *core.Client, db *core.Database, calls *core.Calls) {
	// Play commands
	client.BotClient.AddMessageHandler("/play", handlePlay(client, db, calls))
	client.BotClient.AddMessageHandler("/vplay", handleVPlay(client, db, calls))
	
	// Queue
	client.BotClient.AddMessageHandler("/queue", handleQueue(db))
	client.BotClient.AddMessageHandler("/current", handleCurrent(db))
}

func handlePlay(client *core.Client, db *core.Database, calls *core.Calls) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		m.Reply("**🎵 Play Feature**\n\nComing soon!")
		return nil
	}
}

func handleVPlay(client *core.Client, db *core.Database, calls *core.Calls) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		m.Reply("**📹 Video Play Feature**\n\nComing soon!")
		return nil
	}
}

func handleQueue(db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		m.Reply("**📝 Queue**\n\nNo songs in queue!")
		return nil
	}
}

func handleCurrent(db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		m.Reply("**🎧 Current Song**\n\nNothing is playing!")
		return nil
	}
}
