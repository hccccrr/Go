package utils

// VideoInfo represents YouTube video information
// This is the single source of truth for video metadata
type VideoInfo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Duration  string `json:"duration"`
	Channel   string `json:"channel"`
	Views     string `json:"views"`
	Link      string `json:"link"`
	Thumbnail string `json:"thumbnail"`
}

// YouTubeSearcher interface for getting video data
type YouTubeSearcher interface {
	GetVideoInfo(videoID string) (*VideoInfo, error)
}
