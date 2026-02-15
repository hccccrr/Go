package utils

import (
	"context"
	"fmt"
	"os"
)

// PlayContext contains information needed to play a track
type PlayContext struct {
	ChatID   int64
	UserID   int64
	Duration string
	File     string
	Title    string
	User     string
	VideoID  string
	VCType   string // "voice" or "video"
	Force    bool
}

// VoiceChatManager interface for voice chat operations
type VoiceChatManager interface {
	JoinVC(ctx context.Context, chatID int64, file string, video bool) error
	LeaveVC(ctx context.Context, chatID int64, force bool) error
	ChangeVC(ctx context.Context, chatID int64) error
	ReplayVC(ctx context.Context, chatID int64, file string, video bool) error
}

// YouTubeDownloader interface for downloading tracks
type YouTubeDownloader interface {
	Download(ctx context.Context, videoID string, audio, video bool) (string, error)
	GetData(ctx context.Context, query string, single bool, limit int) ([]VideoData, error)
}

// VideoData represents YouTube video data
type VideoData struct {
	ID       string
	Title    string
	Duration string
	Channel  string
	Views    string
}

// ThumbnailGenerator interface for generating thumbnails
type ThumbnailGenerator interface {
	Generate(width, height int, videoID string) string
}

// PlayDatabase interface for database operations
type PlayDatabase interface {
	UpdateSongsCount(ctx context.Context, count int) error
	UpdateUser(ctx context.Context, userID int64, field string, increment int) error
	IsActiveVC(ctx context.Context, chatID int64) (bool, error)
}

// PlayClient interface for client operations
type PlayClient interface {
	SendMessage(ctx context.Context, chatID int64, text string, buttons interface{}) error
	GetEntity(ctx context.Context, chatID int64) (*ChatEntity, error)
	GetBotUsername() string
	GetBotMention() string
	DeleteMessage(ctx context.Context, message interface{}) error
}

// Player handles music playback operations
type Player struct {
	vcManager VoiceChatManager
	ytube     YouTubeDownloader
	thumb     ThumbnailGenerator
	db        PlayDatabase
	client    PlayClient
	queue     *QueueDB
	buttons   PageButtons
}

// NewPlayer creates a new Player instance
func NewPlayer(
	vcManager VoiceChatManager,
	ytube YouTubeDownloader,
	thumb ThumbnailGenerator,
	db PlayDatabase,
	client PlayClient,
	queue *QueueDB,
	buttons PageButtons,
) *Player {
	return &Player{
		vcManager: vcManager,
		ytube:     ytube,
		thumb:     thumb,
		db:        db,
		client:    client,
		queue:     queue,
		buttons:   buttons,
	}
}

// MessageEditable interface for messages that can be edited/deleted
type MessageEditable interface {
	Edit(ctx context.Context, text string) error
	Reply(ctx context.Context, text string) error
	Delete(ctx context.Context) error
}

// Play plays a track in voice chat
func (p *Player) Play(ctx context.Context, message MessageEditable, playCtx PlayContext, edit bool) error {
	if playCtx.Force {
		if err := p.vcManager.LeaveVC(ctx, playCtx.ChatID, true); err != nil {
			return err
		}
	}

	var filePath string
	var err error

	if playCtx.VideoID == "telegram" {
		filePath = playCtx.File
	} else {
		// Notify download
		if edit {
			message.Edit(ctx, "Downloading ...")
		} else {
			message.Reply(ctx, "Downloading ...")
		}

		// Download from YouTube
		video := playCtx.VCType == "video"
		filePath, err = p.ytube.Download(ctx, playCtx.VideoID, true, video)
		if err != nil {
			errMsg := fmt.Sprintf("Download failed: %v", err)
			if edit {
				message.Edit(ctx, errMsg)
			} else {
				message.Reply(ctx, errMsg)
			}
			return err
		}
	}

	// Add to queue
	position := p.queue.PutQueue(
		playCtx.ChatID,
		playCtx.UserID,
		playCtx.Duration,
		filePath,
		playCtx.Title,
		playCtx.User,
		playCtx.VideoID,
		playCtx.VCType,
		playCtx.Force,
	)

	if position == 0 {
		// First in queue - start playing
		return p.playNow(ctx, message, playCtx, filePath)
	} else {
		// Added to queue
		return p.addToQueue(ctx, message, playCtx, position)
	}
}

// playNow plays the track immediately
func (p *Player) playNow(ctx context.Context, message MessageEditable, playCtx PlayContext, filePath string) error {
	photo := p.thumb.Generate(359, 297, playCtx.VideoID)
	video := playCtx.VCType == "video"

	// Join voice chat
	if err := p.vcManager.JoinVC(ctx, playCtx.ChatID, filePath, video); err != nil {
		message.Delete(ctx)
		p.client.SendMessage(ctx, playCtx.ChatID, fmt.Sprintf("Failed to join VC: %v", err), nil)
		p.queue.ClearQueue(playCtx.ChatID)
		
		// Cleanup
		if filePath != "" && filePath != playCtx.File {
			os.Remove(filePath)
		}
		if photo != "" {
			os.Remove(photo)
		}
		return err
	}

	// Send now playing message
	text := fmt.Sprintf(
		"╭─────────────────────╮\n"+
			"│  **🎵 Now Playing**\n"+
			"╰─────────────────────╯\n\n"+
			"**🔗 Stream:** %s\n\n"+
			"**📝 Song:** `%s`\n"+
			"**⏱️ Duration:** `%s`\n"+
			"**👤 Requested By:** %s",
		p.client.GetBotMention(),
		playCtx.Title,
		playCtx.Duration,
		playCtx.User,
	)

	btns := p.buttons.SongMarkup("", "", 0) // Use appropriate button markup

	if photo != "" {
		// Send with photo
		p.client.SendMessage(ctx, playCtx.ChatID, text, btns)
		os.Remove(photo)
	} else {
		// Send without photo
		p.client.SendMessage(ctx, playCtx.ChatID, text, btns)
	}

	message.Delete(ctx)

	// Update statistics
	p.db.UpdateSongsCount(ctx, 1)
	p.db.UpdateUser(ctx, playCtx.UserID, "songs_played", 1)

	// Log
	entity, _ := p.client.GetEntity(ctx, playCtx.ChatID)
	chatName := "Unknown"
	if entity != nil {
		chatName = entity.Title
	}
	
	fmt.Printf("Playing: %s in %s (%d) by %s\n", playCtx.Title, chatName, playCtx.ChatID, playCtx.User)

	return nil
}

// addToQueue adds track to queue
func (p *Player) addToQueue(ctx context.Context, message MessageEditable, playCtx PlayContext, position int) error {
	text := fmt.Sprintf(
		"╭─────────────────────╮\n"+
			"│  **📋 Added to Queue**\n"+
			"╰─────────────────────╯\n\n"+
			"**🔢 Position:** `#%d`\n"+
			"**📝 Song:** `%s`\n"+
			"**⏱️ Duration:** `%s`\n"+
			"**👤 Queued By:** %s",
		position,
		playCtx.Title,
		playCtx.Duration,
		playCtx.User,
	)

	p.client.SendMessage(ctx, playCtx.ChatID, text, nil)
	message.Delete(ctx)

	return nil
}

// Skip skips the current track
func (p *Player) Skip(ctx context.Context, chatID int64, message MessageEditable) error {
	message.Edit(ctx, "Skipping ...")
	
	if err := p.vcManager.ChangeVC(ctx, chatID); err != nil {
		return err
	}
	
	message.Delete(ctx)
	return nil
}

// Replay replays the current track
func (p *Player) Replay(ctx context.Context, chatID int64, message MessageEditable) error {
	que := p.queue.GetCurrent(chatID)
	if que == nil {
		return message.Edit(ctx, "Nothing is playing to replay")
	}

	video := que.VCType == "video"
	photo := p.thumb.Generate(359, 297, que.VideoID)

	var filePath string
	var err error

	if que.File == que.VideoID {
		filePath, err = p.ytube.Download(ctx, que.VideoID, true, video)
		if err != nil {
			return err
		}
	} else {
		filePath = que.File
	}

	// Replay in VC
	if err := p.vcManager.ReplayVC(ctx, chatID, filePath, video); err != nil {
		message.Delete(ctx)
		p.client.SendMessage(ctx, chatID, fmt.Sprintf("Replay failed: %v", err), nil)
		p.queue.ClearQueue(chatID)
		
		if filePath != que.File {
			os.Remove(filePath)
		}
		if photo != "" {
			os.Remove(photo)
		}
		return err
	}

	// Send now playing message
	text := fmt.Sprintf(
		"╭─────────────────────╮\n"+
			"│  **🎵 Now Playing**\n"+
			"╰─────────────────────╯\n\n"+
			"**🔗 Stream:** %s\n\n"+
			"**📝 Song:** `%s`\n"+
			"**⏱️ Duration:** `%s`\n"+
			"**👤 Requested By:** %s",
		p.client.GetBotMention(),
		que.Title,
		que.Duration,
		que.User,
	)

	if photo != "" {
		p.client.SendMessage(ctx, chatID, text, nil)
		os.Remove(photo)
	} else {
		p.client.SendMessage(ctx, chatID, text, nil)
	}

	message.Delete(ctx)
	return nil
}

// Playlist plays multiple tracks from a playlist
func (p *Player) Playlist(ctx context.Context, message MessageEditable, userID int64, userMention string, collection []string, video bool) error {
	vcType := "voice"
	if video {
		vcType = "video"
	}

	chatID := int64(0) // Get from message context
	count := 0
	failed := 0

	// Check if VC is active
	isActive, _ := p.db.IsActiveVC(ctx, chatID)
	if isActive {
		message.Edit(ctx, "This chat has an active VC. Adding songs from playlist to queue...\n\n__This might take some time!__")
	}

	previously := p.queue.GetQueueLength(chatID)

	for _, item := range collection {
		dataList, err := p.ytube.GetData(ctx, item, true, 1)
		if err != nil || len(dataList) == 0 {
			failed++
			continue
		}

		data := dataList[0]

		if count == 0 && previously == 0 {
			// First track - play immediately
			filePath, err := p.ytube.Download(ctx, data.ID, true, video)
			if err != nil {
				failed++
				continue
			}

			p.queue.PutQueue(chatID, userID, data.Duration, filePath, data.Title, userMention, data.ID, vcType, false)

			photo := p.thumb.Generate(359, 297, data.ID)
			if err := p.vcManager.JoinVC(ctx, chatID, filePath, video); err != nil {
				message.Edit(ctx, fmt.Sprintf("Failed to join VC: %v", err))
				p.queue.ClearQueue(chatID)
				os.Remove(filePath)
				if photo != "" {
					os.Remove(photo)
				}
				return err
			}

			// Send now playing
			text := fmt.Sprintf(
				"**🎵 Now Playing**\n\n"+
					"**📝 Song:** `%s`\n"+
					"**⏱️ Duration:** `%s`\n"+
					"**👤 Requested By:** %s",
				data.Title,
				data.Duration,
				userMention,
			)
			p.client.SendMessage(ctx, chatID, text, nil)
			
			if photo != "" {
				os.Remove(photo)
			}
		} else {
			// Add to queue
			p.queue.PutQueue(chatID, userID, data.Duration, data.ID, data.Title, userMention, data.ID, vcType, false)
		}

		count++
	}

	message.Edit(ctx, fmt.Sprintf(
		"**Added all tracks to queue!**\n\n"+
			"**Total tracks: `%d`**\n"+
			"**Failed: `%d`**",
		count,
		failed,
	))

	return nil
}
