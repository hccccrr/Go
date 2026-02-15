package utils

import (
	"context"
	"fmt"
)

// PageMessage interface represents a message that can be edited
type PageMessage interface {
	Edit(ctx context.Context, text string, buttons interface{}) error
	Reply(ctx context.Context, text string, buttons interface{}) error
	Delete(ctx context.Context) error
	GetChatID() int64
}

// PageButtons interface for creating button markups
type PageButtons interface {
	SongMarkup(randKey, url string, key int) interface{}
	ActiveVCMarkup(count, page int) interface{}
	AuthUsersMarkup(count, page int, randKey string) interface{}
	FavoriteMarkup(collection [][]interface{}, userID int64, page, index int, delete bool) (interface{}, string, error)
	QueueMarkup(count, page int) interface{}
}

// PageDatabase interface for database operations
type PageDatabase interface {
	GetFavorite(ctx context.Context, userID int64, trackID string) (*FavoriteTrack, error)
}

// PageClient interface for Telegram client operations
type PageClient interface {
	GetEntity(ctx context.Context, chatID int64) (*ChatEntity, error)
	GetBotUsername() string
	GetBotMention() string
}

// FavoriteTrack represents a favorite track
type FavoriteTrack struct {
	Title    string
	Duration string
	AddDate  string
}

// ActiveVC represents an active voice chat
type ActiveVC struct {
	Title        string
	ChatID       int64
	Participants int
	Playing      string
	VCType       string
	ActiveSince  string
}

// AuthUser represents an authorized user
type AuthUser struct {
	AuthUser  string
	AdminName string
	AdminID   int64
	AuthDate  string
}

// SongCache represents cached song data
type SongCache struct {
	Title     string
	Link      string
	Thumbnail string
}

// Pages handles pagination for various lists
type Pages struct {
	buttons PageButtons
	db      PageDatabase
	client  PageClient
}

// NewPages creates a new Pages instance
func NewPages(buttons PageButtons, db PageDatabase, client PageClient) *Pages {
	return &Pages{
		buttons: buttons,
		db:      db,
		client:  client,
	}
}

// SongPage displays song page with navigation
func (p *Pages) SongPage(ctx context.Context, message PageMessage, songCache map[string][]SongCache, randKey string, key int) error {
	allTracks, exists := songCache[randKey]
	if !exists || key >= len(allTracks) {
		if err := message.Delete(ctx); err != nil {
			return err
		}
		return fmt.Errorf("query timed out")
	}

	track := allTracks[key]
	btns := p.buttons.SongMarkup(randKey, track.Link, key)
	
	caption := fmt.Sprintf(
		"__(%d/%d)__ **Song Downloader:**\n\n"+
			"**• Title:** `%s`\n\n"+
			"🎶 @%s",
		key+1,
		len(allTracks),
		track.Title,
		p.client.GetBotUsername(),
	)

	return message.Edit(ctx, caption, btns)
}

// ActiveVCPage displays active voice chats page
func (p *Pages) ActiveVCPage(ctx context.Context, message PageMessage, collection []ActiveVC, page, index int, edit bool) error {
	grouped := p.groupActiveVCs(collection, 5)
	total := len(collection)
	
	text := fmt.Sprintf(
		"__(%d/%d)__ **@%s Active Voice Chats:** __%d chats__\n\n",
		page+1,
		len(grouped),
		p.client.GetBotUsername(),
		total,
	)
	
	btns := p.buttons.ActiveVCMarkup(len(grouped), page)
	
	// Handle index out of range
	if page >= len(grouped) {
		page = 0
	}
	
	for _, active := range grouped[page] {
		index++
		prefix := "0" + fmt.Sprint(index)
		if index >= 10 {
			prefix = fmt.Sprint(index)
		}
		
		text += fmt.Sprintf(
			"**%s:** %s [`%d`]\n"+
				"    **Listeners:** __%d__\n"+
				"    **Playing:** __%s__\n"+
				"    **VC Type:** __%s__\n"+
				"    **Since:** __%s__\n\n",
			prefix,
			active.Title,
			active.ChatID,
			active.Participants,
			active.Playing,
			active.VCType,
			active.ActiveSince,
		)
	}
	
	if edit {
		return message.Edit(ctx, text, btns)
	}
	return message.Reply(ctx, text, btns)
}

// AuthUsersPage displays authorized users page
func (p *Pages) AuthUsersPage(ctx context.Context, message PageMessage, collection []AuthUser, randKey string, page, index int, edit bool) error {
	grouped := p.groupAuthUsers(collection, 6)
	total := len(collection)
	
	// Get chat title
	entity, err := p.client.GetEntity(ctx, message.GetChatID())
	chatTitle := "Unknown Chat"
	if err == nil {
		chatTitle = entity.Title
	}
	
	text := fmt.Sprintf(
		"__(%d/%d)__ **Authorized Users in %s:**\n    >> __%d users__\n\n",
		page+1,
		len(grouped),
		chatTitle,
		total,
	)
	
	btns := p.buttons.AuthUsersMarkup(len(grouped), page, randKey)
	
	// Handle index out of range
	if page >= len(grouped) {
		page = 0
	}
	
	for _, auth := range grouped[page] {
		index++
		prefix := "0" + fmt.Sprint(index)
		if index >= 10 {
			prefix = fmt.Sprint(index)
		}
		
		text += fmt.Sprintf(
			"**%s:** %s\n"+
				"    **Auth By:** %s (`%d`)\n"+
				"    **Since:** __%s__\n\n",
			prefix,
			auth.AuthUser,
			auth.AdminName,
			auth.AdminID,
			auth.AuthDate,
		)
	}
	
	if edit {
		return message.Edit(ctx, text, btns)
	}
	return message.Reply(ctx, text, btns)
}

// FavoritePage displays favorite tracks page
func (p *Pages) FavoritePage(ctx context.Context, message PageMessage, collection []string, userID int64, mention string, page, index int, edit, delete bool) error {
	grouped := p.groupStrings(collection, 5)
	total := len(collection)
	
	text := fmt.Sprintf(
		"__(%d/%d)__ %s **favorites:** __%d tracks__\n\n",
		page+1,
		len(grouped),
		mention,
		total,
	)
	
	// Handle index out of range
	if page >= len(grouped) {
		page = 0
	}
	
	// Build favorites text
	favText := ""
	currentIndex := index
	for _, trackID := range grouped[page] {
		currentIndex++
		
		fav, err := p.db.GetFavorite(ctx, userID, trackID)
		if err != nil {
			continue
		}
		
		prefix := "0" + fmt.Sprint(currentIndex)
		if currentIndex >= 10 {
			prefix = fmt.Sprint(currentIndex)
		}
		
		favText += fmt.Sprintf(
			"**%s:** %s\n"+
				"    **Duration:** %s\n"+
				"    **Since:** %s\n\n",
			prefix,
			fav.Title,
			fav.Duration,
			fav.AddDate,
		)
	}
	
	// Convert grouped to interface{} for buttons
	groupedInterface := make([][]interface{}, len(grouped))
	for i, group := range grouped {
		groupedInterface[i] = make([]interface{}, len(group))
		for j, item := range group {
			groupedInterface[i][j] = item
		}
	}
	
	btns, finalText, err := p.buttons.FavoriteMarkup(groupedInterface, userID, page, index, delete)
	if err != nil {
		return err
	}
	
	fullText := text + favText + finalText
	
	if edit {
		return message.Edit(ctx, fullText, btns)
	}
	return message.Reply(ctx, fullText, btns)
}

// QueuePage displays queue page
func (p *Pages) QueuePage(ctx context.Context, message PageMessage, collection []QueueItem, page, index int, edit bool) error {
	grouped := p.groupQueueItems(collection, 5)
	total := len(collection)
	
	text := fmt.Sprintf(
		"__(%d/%d)__ **In Queue:** __%d tracks__\n\n",
		page+1,
		len(grouped),
		total,
	)
	
	btns := p.buttons.QueueMarkup(len(grouped), page)
	
	// Handle index out of range
	if page >= len(grouped) {
		return message.Edit(ctx, "**No more tracks in queue!**", nil)
	}
	
	for _, que := range grouped[page] {
		index++
		prefix := "0" + fmt.Sprint(index)
		if index >= 10 {
			prefix = fmt.Sprint(index)
		}
		
		text += fmt.Sprintf(
			"**%s:** %s\n"+
				"    **VC Type:** %s\n"+
				"    **Requested By:** %s\n"+
				"    **Duration:** __%s__\n\n",
			prefix,
			que.Title,
			que.VCType,
			que.User,
			que.Duration,
		)
	}
	
	if edit {
		return message.Edit(ctx, text, btns)
	}
	return message.Reply(ctx, text, btns)
}

// Helper functions for grouping

func (p *Pages) groupActiveVCs(items []ActiveVC, groupSize int) [][]ActiveVC {
	var grouped [][]ActiveVC
	for i := 0; i < len(items); i += groupSize {
		end := i + groupSize
		if end > len(items) {
			end = len(items)
		}
		grouped = append(grouped, items[i:end])
	}
	return grouped
}

func (p *Pages) groupAuthUsers(items []AuthUser, groupSize int) [][]AuthUser {
	var grouped [][]AuthUser
	for i := 0; i < len(items); i += groupSize {
		end := i + groupSize
		if end > len(items) {
			end = len(items)
		}
		grouped = append(grouped, items[i:end])
	}
	return grouped
}

func (p *Pages) groupStrings(items []string, groupSize int) [][]string {
	var grouped [][]string
	for i := 0; i < len(items); i += groupSize {
		end := i + groupSize
		if end > len(items) {
			end = len(items)
		}
		grouped = append(grouped, items[i:end])
	}
	return grouped
}

func (p *Pages) groupQueueItems(items []QueueItem, groupSize int) [][]QueueItem {
	var grouped [][]QueueItem
	for i := 0; i < len(items); i += groupSize {
		end := i + groupSize
		if end > len(items) {
			end = len(items)
		}
		grouped = append(grouped, items[i:end])
	}
	return grouped
}
