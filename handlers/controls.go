package handlers

import (
	"fmt"
	"strconv"
	"strings"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/utils"
	"shizumusic/helpers"
)

// TEXTS templates from utils
var TEXTS = helpers.TextTemplates

// RegisterControlHandlers registers playback control handlers
func RegisterControlHandlers(client *core.Client, db *core.Database, calls *core.Calls) {
	// Mute/Unmute
	client.BotClient.AddMessageHandler("/mute", func(m *tg.NewMessage) error {
		return core.AuthOnly(db)(handleMute(db, calls))(m)
	})
	
	client.BotClient.AddMessageHandler("/unmute", func(m *tg.NewMessage) error {
		return core.AuthOnly(db)(handleUnmute(db, calls))(m)
	})

	// Pause/Resume
	client.BotClient.AddMessageHandler("/pause", func(m *tg.NewMessage) error {
		return core.AuthOnly(db)(handlePause(db, calls))(m)
	})
	
	client.BotClient.AddMessageHandler("/resume", func(m *tg.NewMessage) error {
		return core.AuthOnly(db)(handleResume(db, calls))(m)
	})

	// Stop/End
	client.BotClient.AddMessageHandler("/stop", func(m *tg.NewMessage) error {
		return core.AuthOnly(db)(handleStop(db, calls))(m)
	})
	
	client.BotClient.AddMessageHandler("/end", func(m *tg.NewMessage) error {
		return core.AuthOnly(db)(handleStop(db, calls))(m)
	})

	// Loop
	client.BotClient.AddMessageHandler("/loop", func(m *tg.NewMessage) error {
		return core.AuthOnly(db)(handleLoop(db))(m)
	})

	// Replay
	client.BotClient.AddMessageHandler("/replay", func(m *tg.NewMessage) error {
		return core.AuthOnly(db)(handleReplay(db, calls))(m)
	})

	// Skip
	client.BotClient.AddMessageHandler("/skip", func(m *tg.NewMessage) error {
		return core.AuthOnly(db)(handleSkip(db, calls))(m)
	})

	// Seek
	client.BotClient.AddMessageHandler("/seek", func(m *tg.NewMessage) error {
		return core.AuthOnly(db)(handleSeek(db, calls))(m)
	})
}

func handleMute(db *core.Database, calls *core.Calls) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !m.IsGroup || config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		// Check if already muted
		// is_muted := db.GetWatcher(m.Chat.ID, "mute")
		
		// Mute VC
		if err := calls.MuteVC(m.Chat.ID); err != nil {
			m.Reply("Failed to mute voice chat!")
			return err
		}

		mention := fmt.Sprintf("[%s](tg://user?id=%d)", m.From.FirstName, m.From.ID)
		m.Reply(fmt.Sprintf("__VC Muted by:__ %s", mention))
		return nil
	}
}

func handleUnmute(db *core.Database, calls *core.Calls) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !m.IsGroup || config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		// Unmute VC
		if err := calls.UnmuteVC(m.Chat.ID); err != nil {
			m.Reply("Failed to unmute voice chat!")
			return err
		}

		mention := fmt.Sprintf("[%s](tg://user?id=%d)", m.From.FirstName, m.From.ID)
		m.Reply(fmt.Sprintf("__VC Unmuted by:__ %s", mention))
		return nil
	}
}

func handlePause(db *core.Database, calls *core.Calls) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !m.IsGroup || config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		// Pause VC
		if err := calls.PauseVC(m.Chat.ID); err != nil {
			m.Reply("Failed to pause voice chat!")
			return err
		}

		mention := fmt.Sprintf("[%s](tg://user?id=%d)", m.From.FirstName, m.From.ID)
		m.Reply(fmt.Sprintf("__VC Paused by:__ %s", mention))
		return nil
	}
}

func handleResume(db *core.Database, calls *core.Calls) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !m.IsGroup || config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		// Resume VC
		if err := calls.ResumeVC(m.Chat.ID); err != nil {
			m.Reply("Failed to resume voice chat!")
			return err
		}

		mention := fmt.Sprintf("[%s](tg://user?id=%d)", m.From.FirstName, m.From.ID)
		m.Reply(fmt.Sprintf("__VC Resumed by:__ %s", mention))
		return nil
	}
}

func handleStop(db *core.Database, calls *core.Calls) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !m.IsGroup || config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		// Leave VC
		if err := calls.LeaveVC(m.Chat.ID); err != nil {
			m.Reply("Failed to stop voice chat!")
			return err
		}

		// Reset loop
		db.SetLoop(m.Chat.ID, 0)

		mention := fmt.Sprintf("[%s](tg://user?id=%d)", m.From.FirstName, m.From.ID)
		m.Reply(fmt.Sprintf("__VC Stopped by:__ %s", mention))
		return nil
	}
}

func handleLoop(db *core.Database) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !m.IsGroup || config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		parts := strings.Fields(m.Text)
		if len(parts) < 2 {
			return m.Reply(
				"Please specify the number of times to loop the song!\n\n" +
					"Maximum loop range is **10**. Give **0** to disable loop.",
			)
		}

		loopCount, err := strconv.Atoi(parts[1])
		if err != nil {
			return m.Reply(
				"Please enter a valid number!\n\n" +
					"Maximum loop range is **10**. Give **0** to disable loop.",
			)
		}

		currentLoop, _ := db.GetLoop(m.Chat.ID)
		mention := fmt.Sprintf("[%s](tg://user?id=%d)", m.From.FirstName, m.From.ID)

		// Disable loop
		if loopCount == 0 {
			if currentLoop == 0 {
				return m.Reply("There is no active loop in this chat!")
			}
			db.SetLoop(m.Chat.ID, 0)
			return m.Reply(fmt.Sprintf(
				"__Loop disabled by:__ %s\n\nPrevious loop was: `%d`",
				mention, currentLoop,
			))
		}

		// Set loop (1-10)
		if loopCount >= 1 && loopCount <= 10 {
			finalLoop := currentLoop + loopCount
			if finalLoop > 10 {
				finalLoop = 10
			}
			db.SetLoop(m.Chat.ID, finalLoop)
			return m.Reply(fmt.Sprintf(
				"__Loop set to:__ `%d`\n__By:__ %s\n\nPrevious loop was: `%d`",
				finalLoop, mention, currentLoop,
			))
		}

		return m.Reply(
			"Please enter a valid number!\n\n" +
				"Maximum loop range is **10**. Give **0** to disable loop.",
		)
	}
}

func handleReplay(db *core.Database, calls *core.Calls) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !m.IsGroup || config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		active, _ := db.IsActiveVC(m.Chat.ID)
		if !active {
			return m.Reply("No active Voice Chat found here!")
		}

		msg, _ := m.Reply("Replaying...")
		
		// TODO: Implement replay logic
		// queue := Queue.GetQueue(m.Chat.ID)
		// if len(queue) == 0 {
		//     msg.Edit("No songs in the queue to replay!")
		//     return nil
		// }
		// player.Replay(m.Chat.ID, msg)

		msg.Edit("Replay feature coming soon!")
		return nil
	}
}

func handleSkip(db *core.Database, calls *core.Calls) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !m.IsGroup || config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		active, _ := db.IsActiveVC(m.Chat.ID)
		if !active {
			return m.Reply("No active Voice Chat found here!")
		}

		msg, _ := m.Reply("Processing...")
		
		// TODO: Implement skip logic
		// queue := Queue.GetQueue(m.Chat.ID)
		// if len(queue) == 0 {
		//     msg.Edit("No songs in the queue to skip!")
		//     return nil
		// }
		// if len(queue) == 1 {
		//     msg.Edit("No more songs in queue to skip! Use /end or /stop to stop the VC.")
		//     return nil
		// }
		
		// Check and disable loop if needed
		currentLoop, _ := db.GetLoop(m.Chat.ID)
		if currentLoop != 0 {
			msg.Edit("Disabled Loop to skip the current song!")
			db.SetLoop(m.Chat.ID, 0)
		}

		// player.Skip(m.Chat.ID, msg)
		msg.Edit("Skip feature coming soon!")
		return nil
	}
}

func handleSeek(db *core.Database, calls *core.Calls) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if !m.IsGroup || config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		active, _ := db.IsActiveVC(m.Chat.ID)
		if !active {
			return m.Reply("No active Voice Chat found here!")
		}

		parts := strings.Fields(m.Text)
		if len(parts) < 2 {
			return m.Reply(
				"Please specify the time to seek!\n\n" +
					"**Example:**\n" +
					"__- Seek 10 secs forward >__ `/seek 10`\n" +
					"__- Seek 10 secs backward >__ `/seek -10`",
			)
		}

		msg, _ := m.Reply("Seeking...")
		
		// TODO: Implement seek logic
		// Parse seek time
		// Check if forward or backward
		// Update queue duration
		// Seek in NTgCalls

		msg.Edit("Seek feature coming soon!")
		return nil
	}
}
