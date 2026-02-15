package utils

import "sync"

// QueueItem represents a track in the queue
type QueueItem struct {
	ChatID   int64  `json:"chat_id"`
	UserID   int64  `json:"user_id"`
	Duration string `json:"duration"`
	File     string `json:"file"`
	Title    string `json:"title"`
	User     string `json:"user"`
	VideoID  string `json:"video_id"`
	VCType   string `json:"vc_type"` // "voice" or "video"
	Played   int    `json:"played"`  // Seconds already played
}

// QueueDB manages music queues for all chats
type QueueDB struct {
	queue map[int64][]QueueItem
	cache map[int64][]string // Cache for file paths
	mu    sync.RWMutex
}

// NewQueueDB creates a new queue database
func NewQueueDB() *QueueDB {
	return &QueueDB{
		queue: make(map[int64][]QueueItem),
		cache: make(map[int64][]string),
	}
}

// PutQueue adds a track to the queue
// Returns position in queue (0-indexed)
func (q *QueueDB) PutQueue(
	chatID int64,
	userID int64,
	duration string,
	file string,
	title string,
	user string,
	videoID string,
	vcType string,
	forceplay bool,
) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	item := QueueItem{
		ChatID:   chatID,
		UserID:   userID,
		Duration: duration,
		File:     file,
		Title:    title,
		User:     user,
		VideoID:  videoID,
		VCType:   vcType,
		Played:   0,
	}

	// Initialize queue if not exists
	if q.queue[chatID] == nil {
		q.queue[chatID] = []QueueItem{}
	}

	if forceplay {
		// Insert at beginning (force play)
		q.queue[chatID] = append([]QueueItem{item}, q.queue[chatID]...)
	} else {
		// Append to end
		q.queue[chatID] = append(q.queue[chatID], item)
	}

	// Add to cache
	if q.cache[chatID] == nil {
		q.cache[chatID] = []string{}
	}
	q.cache[chatID] = append(q.cache[chatID], file)

	// Return position (0-indexed)
	position := len(q.queue[chatID]) - 1
	if forceplay {
		position = 0
	}

	return position
}

// GetQueue returns the entire queue for a chat
func (q *QueueDB) GetQueue(chatID int64) []QueueItem {
	q.mu.RLock()
	defer q.mu.RUnlock()

	queue := q.queue[chatID]
	if queue == nil {
		return []QueueItem{}
	}

	// Return a copy to prevent external modification
	result := make([]QueueItem, len(queue))
	copy(result, queue)
	return result
}

// GetQueueLength returns the number of tracks in queue
func (q *QueueDB) GetQueueLength(chatID int64) int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.queue[chatID] == nil {
		return 0
	}
	return len(q.queue[chatID])
}

// RmQueue removes a track from queue by index
// Returns the file path of removed track
func (q *QueueDB) RmQueue(chatID int64, index int) string {
	q.mu.Lock()
	defer q.mu.Unlock()

	queue := q.queue[chatID]
	if queue == nil || index < 0 || index >= len(queue) {
		return ""
	}

	file := queue[index].File

	// Remove from queue
	q.queue[chatID] = append(queue[:index], queue[index+1:]...)

	return file
}

// ClearQueue clears all tracks from queue
func (q *QueueDB) ClearQueue(chatID int64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.queue[chatID] = []QueueItem{}
	q.cache[chatID] = []string{}
}

// GetCurrent returns the currently playing track (first in queue)
func (q *QueueDB) GetCurrent(chatID int64) *QueueItem {
	q.mu.RLock()
	defer q.mu.RUnlock()

	queue := q.queue[chatID]
	if queue == nil || len(queue) == 0 {
		return nil
	}

	// Return a copy
	current := queue[0]
	return &current
}

// PopCurrent removes and returns the current track
func (q *QueueDB) PopCurrent(chatID int64) *QueueItem {
	q.mu.Lock()
	defer q.mu.Unlock()

	queue := q.queue[chatID]
	if queue == nil || len(queue) == 0 {
		return nil
	}

	current := queue[0]
	q.queue[chatID] = queue[1:]

	return &current
}

// UpdateDuration updates the played duration of current track
// seekType: 0 for rewind (subtract), 1 for forward (add)
func (q *QueueDB) UpdateDuration(chatID int64, seekType int, time int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	queue := q.queue[chatID]
	if queue == nil || len(queue) == 0 {
		return
	}

	if seekType == 0 {
		// Rewind
		q.queue[chatID][0].Played -= time
		if q.queue[chatID][0].Played < 0 {
			q.queue[chatID][0].Played = 0
		}
	} else {
		// Forward
		q.queue[chatID][0].Played += time
	}
}

// GetPlayed returns seconds already played for current track
func (q *QueueDB) GetPlayed(chatID int64) int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	queue := q.queue[chatID]
	if queue == nil || len(queue) == 0 {
		return 0
	}

	return queue[0].Played
}

// SetPlayed sets the played duration for current track
func (q *QueueDB) SetPlayed(chatID int64, played int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	queue := q.queue[chatID]
	if queue == nil || len(queue) == 0 {
		return
	}

	q.queue[chatID][0].Played = played
}

// IsQueueEmpty checks if queue is empty
func (q *QueueDB) IsQueueEmpty(chatID int64) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	queue := q.queue[chatID]
	return queue == nil || len(queue) == 0
}

// GetCache returns cached file paths
func (q *QueueDB) GetCache(chatID int64) []string {
	q.mu.RLock()
	defer q.mu.RUnlock()

	cache := q.cache[chatID]
	if cache == nil {
		return []string{}
	}

	result := make([]string, len(cache))
	copy(result, cache)
	return result
}

// ClearCache clears cache for a chat
func (q *QueueDB) ClearCache(chatID int64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.cache[chatID] = []string{}
}

// Global queue instance
var Queue = NewQueueDB()
