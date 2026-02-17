package handlers

import (
	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/core"
)

func RegisterPlayHandlers(client *core.Client, db *core.Database, calls *core.Calls) {
	// Play commands
	client.BotClient.AddMessageHandler("/play", func(m *tg.NewMessage) error {
		return handlePlay(client, db, calls)(m)
	})
	client.BotClient.AddMessageHandler("/vplay", func(m *tg.NewMessage) error {
		return handleVPlay(client, db, calls)(m)
	})

	// Queue
	client.BotClient.AddMessageHandler("/queue", func(m *tg.NewMessage) error {
		return handleQueue(db)(m)
	})
	client.BotClient.AddMessageHandler("/current", func(m *tg.NewMessage) error {
		return handleCurrent(db)(m)
	})
}

func handlePlay(client *core.Client, db *core.Database, calls *core.Calls) tg.MessageHandler {
	return func(m *tg.NewMessage) error {
		m.Reply("**🎵 Play Feature**\n\nComing soon!")
		return nil
	}
}

func handleVPlay(client *core.Client, db *core.Database, calls *core.Calls) tg.MessageHandler {
	return func(m *tg.NewMessage) error {
		m.Reply("**📹 Video Play Feature**\n\nComing soon!")
		return nil
	}
}

func handleQueue(db *core.Database) tg.MessageHandler {
	return func(m *tg.NewMessage) error {
		m.Reply("**📋 Queue**\n\nNo songs in queue!")
		return nil
	}
}

func handleCurrent(db *core.Database) tg.MessageHandler {
	return func(m *tg.NewMessage) error {
		m.Reply("**🎧 Current Song**\n\nNothing is playing!")
		return nil
	}
}
