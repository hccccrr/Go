package handlers

import (
	"context"
	"fmt"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/helpers"
	"shizumusic/utils"
)

func init() {
	RegisterPlugin("play_commands", func(client *core.Client, db *core.Database) {
		calls := core.NewCalls(client.UserClient)
		adapter := core.NewVCAdapter(calls)
		player := utils.NewPlayer(adapter, utils.YTube, utils.Thumb, db, nil, utils.Queue)

		client.BotClient.AddMessageHandler("/play", func(m *tg.NewMessage) error {
			return handlePlay(m, client, player, false, false)
		})
		client.BotClient.AddMessageHandler("/vplay", func(m *tg.NewMessage) error {
			return handlePlay(m, client, player, true, false)
		})
		client.BotClient.AddMessageHandler("/fplay", func(m *tg.NewMessage) error {
			return handlePlay(m, client, player, false, true)
		})
		client.BotClient.AddMessageHandler("/fvplay", func(m *tg.NewMessage) error {
			return handlePlay(m, client, player, true, true)
		})
		client.BotClient.AddMessageHandler("/queue", func(m *tg.NewMessage) error {
			return handleQueue(m)
		})
		client.BotClient.AddMessageHandler("/current", func(m *tg.NewMessage) error {
			return handleCurrent(m, client)
		})
	})
}

// tgMessage wraps *tg.NewMessage to satisfy utils.MessageEditable
type tgMessage struct {
	msg *tg.NewMessage
}

func (t *tgMessage) Edit(ctx context.Context, text string) error {
	_, err := t.msg.Edit(text)
	return err
}

func (t *tgMessage) Reply(ctx context.Context, text string) error {
	_, err := t.msg.Reply(text)
	return err
}

func (t *tgMessage) Delete(ctx context.Context) error {
	_, err := t.msg.Delete()
	return err
}

func handlePlay(m *tg.NewMessage, client *core.Client, player *utils.Player, video bool, force bool) error {
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

	parts := strings.SplitN(m.Text(), " ", 2)
	if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
		m.Reply("**Usage:** `/play <song name or YouTube URL>`")
		return nil
	}
	query := strings.TrimSpace(parts[1])

	hell, _ := m.Reply("🔍 Searching ...")
	hellMsg := &tgMessage{msg: m}
	_ = hell

	ctx := context.Background()
	mention := fmt.Sprintf("[%s](tg://user?id=%d)", sender.FirstName, sender.ID)

	vcType := "voice"
	if video {
		vcType = "video"
	}

	// YouTube URL
	if strings.Contains(query, "youtube.com") || strings.Contains(query, "youtu.be") {
		videoID := utils.ExtractVideoIDFromLink(query)
		if videoID == "" {
			hell.Edit("❌ Invalid YouTube URL.")
			return nil
		}

		info, err := utils.YTube.GetVideoInfo(ctx, videoID)
		if err != nil || info == nil {
			hell.Edit("❌ Could not fetch video info.")
			return nil
		}

		playCtx := utils.PlayContext{
			ChatID:   m.ChatID(),
			UserID:   sender.ID,
			Duration: info.Duration,
			File:     videoID,
			Title:    info.Title,
			User:     mention,
			VideoID:  videoID,
			VCType:   vcType,
			Force:    force,
		}
		return player.Play(ctx, hellMsg, playCtx, true)
	}

	// Search query
	results, err := utils.YTube.GetData(ctx, query, true, 1)
	if err != nil || len(results) == 0 {
		hell.Edit("❌ No results found. Try a different query.")
		return nil
	}

	result := results[0]
	playCtx := utils.PlayContext{
		ChatID:   m.ChatID(),
		UserID:   sender.ID,
		Duration: result.Duration,
		File:     result.ID,
		Title:    result.Title,
		User:     mention,
		VideoID:  result.ID,
		VCType:   vcType,
		Force:    force,
	}
	return player.Play(ctx, hellMsg, playCtx, true)
}

func handleQueue(m *tg.NewMessage) error {
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

	queue := utils.Queue.GetQueue(m.ChatID())
	if len(queue) == 0 {
		btns := helpers.Buttons.CloseMarkup()
		m.Reply(helpers.TextTemplates.QueueEmpty(), &tg.SendOptions{ReplyMarkup: btns})
		return nil
	}

	text := "╭─────────────────────╮\n│  **📋 Queue**\n╰─────────────────────╯\n\n"
	for i, item := range queue {
		if i == 0 {
			text += fmt.Sprintf("**▶️ Now:** `%s` | `%s`\n\n**📌 Up Next:**\n", item.Title, item.Duration)
		} else {
			text += fmt.Sprintf("`%d.` `%s` | `%s` | %s\n", i, item.Title, item.Duration, item.User)
		}
	}

	btns := helpers.Buttons.QueueMarkup(len(queue), 0)
	m.Reply(text, &tg.SendOptions{ReplyMarkup: btns})
	return nil
}

func handleCurrent(m *tg.NewMessage, client *core.Client) error {
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

	que := utils.Queue.GetCurrent(m.ChatID())
	if que == nil {
		btns := helpers.Buttons.CloseMarkup()
		m.Reply(helpers.TextTemplates.NothingPlaying(), &tg.SendOptions{ReplyMarkup: btns})
		return nil
	}

	me, _ := client.BotClient.GetMe()
	btns := helpers.Buttons.PlayerMarkup(m.ChatID(), que.VideoID, me.Username)

	text := fmt.Sprintf(helpers.TextTemplates.Playing(),
		fmt.Sprintf("@%s", me.Username),
		que.Title,
		que.Duration,
		que.User,
	)

	photo := utils.Thumb.Generate(359, 297, que.VideoID)
	m.Reply(text, &tg.SendOptions{
		ReplyMarkup: btns,
		Media:       photo,
	})
	return nil
}
