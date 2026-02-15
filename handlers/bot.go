package handlers

import (
	"fmt"
	"strings"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/helpers"
)

// TEXTS templates from helpers
var TEXTS = helpers.TextTemplates

func RegisterBotHandlers(client *core.Client, db *core.Database) {
	client.BotClient.AddMessageHandler("/start", func(m *tg.NewMessage) error {
		return handleStart(m, client, db)
	})
	client.BotClient.AddMessageHandler("/help", func(m *tg.NewMessage) error {
		return handleHelp(m, client)
	})
	client.BotClient.AddMessageHandler("/ping", func(m *tg.NewMessage) error {
		return handlePing(m, client)
	})
	client.BotClient.AddMessageHandler("/sysinfo", func(m *tg.NewMessage) error {
		return core.UserOnly(handleSysinfo)(m)
	})
}

func handleStart(m *tg.NewMessage, client *core.Client, db *core.Database) error {
	if config.Cfg.IsBanned(m.From.ID) {
		return nil
	}
	if m.IsPrivate {
		return sendStartPM(m, client)
	}
	if m.IsGroup {
		return sendStartGC(m, client)
	}
	return nil
}

func handleHelp(m *tg.NewMessage, client *core.Client) error {
	if config.Cfg.IsBanned(m.From.ID) {
		return nil
	}
	if m.IsPrivate {
		return sendHelpPM(m, client)
	}
	if m.IsGroup {
		return sendHelpGC(m, client)
	}
	return nil
}

func handlePing(m *tg.NewMessage, client *core.Client) error {
	if config.Cfg.IsBanned(m.From.ID) {
		return nil
	}
	start := time.Now()
	msg, _ := m.Reply("Pong!")
	elapsed := time.Since(start).Milliseconds()
	uptime := formatUptime(time.Since(config.Cfg.StartTime))
	pingText := fmt.Sprintf(TEXTS.PingReply(), elapsed, uptime, "50")
	msg.Edit(pingText)
	return nil
}

func handleSysinfo(m *tg.NewMessage) error {
	if config.Cfg.IsBanned(m.From.ID) {
		return nil
	}
	uptime := formatUptime(time.Since(config.Cfg.StartTime))
	me, _ := m.Client.GetMe()
	text := fmt.Sprintf(TEXTS.System(), 4, "25%", "35%", "45%", uptime, fmt.Sprintf("@%s", me.Username))
	m.Reply(text)
	return nil
}

func sendStartPM(m *tg.NewMessage, client *core.Client) error {
	me, _ := client.BotClient.GetMe()
	text := fmt.Sprintf(TEXTS.StartPM(), m.From.FirstName, me.FirstName, me.Username)
	m.Reply(text)
	return nil
}

func sendStartGC(m *tg.NewMessage, client *core.Client) error {
	m.Reply(TEXTS.StartGC())
	return nil
}

func sendHelpPM(m *tg.NewMessage, client *core.Client) error {
	me, _ := client.BotClient.GetMe()
	text := fmt.Sprintf(TEXTS.HelpPM(), fmt.Sprintf("@%s", me.Username))
	m.Reply(text)
	return nil
}

func sendHelpGC(m *tg.NewMessage, client *core.Client) error {
	m.Reply(TEXTS.HelpGC())
	return nil
}

func formatUptime(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	if hours > 24 {
		days := hours / 24
		hours = hours % 24
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
