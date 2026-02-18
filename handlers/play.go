package handlers

import (
	"fmt"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/helpers"
)

func init() {
	RegisterPlugin("play_commands", func(client *core.Client, db *core.Database) {
		calls := core.NewCalls(client.UserClient)

		client.BotClient.AddMessageHandler("/play", func(m *tg.NewMessage) error {
			return handlePlay(m, client, db, calls)
		})
		client.BotClient.AddMessageHandler("/vplay", func(m *tg.NewMessage) error {
			return handleVPlay(m, client, db, calls)
		})
		client.BotClient.AddMessageHandler("/queue", func(m *tg.NewMessage) error {
			return handleQueue(m, db)
		})
		client.BotClient.AddMessageHandler("/current", func(m *tg.NewMessage) error {
			return handleCurrent(m, db)
		})
	})
}

func handlePlay(m *tg.NewMessage, client *core.Client, db *core.Database, calls *core.Calls) error {
	sender, err := m.GetSender()
	if err != nil || sender == nil {
		return nil
	}
	if config.Cfg.IsBanned(sender.ID) {
		return nil
	}
	if !m.IsGroup() {
		return nil
	}

	me, _ := client.BotClient.GetMe()
	mention := fmt.Sprintf("[%s](tg://user?id=%d)", sender.FirstName, sender.ID)
	text := fmt.Sprintf(helpers.TextTemplates.PlayReply(), mention, fmt.Sprintf("@%s", me.Username))
	btns := helpers.Buttons.PlayerMarkup(m.ChatID(), "coming_soon", me.Username)
	m.Reply(text, tg.SendOptions{ReplyMarkup: btns})
	return nil
}

func handleVPlay(m *tg.NewMessage, client *core.Client, db *core.Database, calls *core.Calls) error {
	sender, err := m.GetSender()
	if err != nil || sender == nil {
		return nil
	}
	if config.Cfg.IsBanned(sender.ID) {
		return nil
	}
	if !m.IsGroup() {
		return nil
	}

	me, _ := client.BotClient.GetMe()
	mention := fmt.Sprintf("[%s](tg://user?id=%d)", sender.FirstName, sender.ID)
	text := fmt.Sprintf(helpers.TextTemplates.VPlayReply(), mention, fmt.Sprintf("@%s", me.Username))
	btns := helpers.Buttons.PlayerMarkup(m.ChatID(), "coming_soon", me.Username)
	m.Reply(text, tg.SendOptions{ReplyMarkup: btns})
	return nil
}

func handleQueue(m *tg.NewMessage, db *core.Database) error {
	sender, err := m.GetSender()
	if err != nil || sender == nil {
		return nil
	}
	if config.Cfg.IsBanned(sender.ID) {
		return nil
	}
	if !m.IsGroup() {
		return nil
	}

	btns := helpers.Buttons.QueueMarkup(1, 0)
	m.Reply(helpers.TextTemplates.QueueEmpty(), tg.SendOptions{ReplyMarkup: btns})
	return nil
}

func handleCurrent(m *tg.NewMessage, db *core.Database) error {
	sender, err := m.GetSender()
	if err != nil || sender == nil {
		return nil
	}
	if config.Cfg.IsBanned(sender.ID) {
		return nil
	}
	if !m.IsGroup() {
		return nil
	}

	btns := helpers.Buttons.CloseMarkup()
	m.Reply(helpers.TextTemplates.NothingPlaying(), tg.SendOptions{ReplyMarkup: btns})
	return nil
}
