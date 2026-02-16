package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// API Configuration
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

var (
	yourAPIURL     string
	fallbackAPIURL = "https://shrutibots.site"
	apiURLMutex    sync.RWMutex
)

// LoadAPIURL loads API URL from remote source
func LoadAPIURL(ctx context.Context) error {
	client := &http.Client{Timeout: 10 * time.Second}
	
	resp, err := client.Get("https://pastebin.com/raw/rLsBhAQa")
	if err != nil {
		apiURLMutex.Lock()
		yourAPIURL = fallbackAPIURL
		apiURLMutex.Unlock()
		return fmt.Errorf("using fallback API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		apiURLMutex.Lock()
		yourAPIURL = fallbackAPIURL
		apiURLMutex.Unlock()
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		apiURLMutex.Lock()
		yourAPIURL = fallbackAPIURL
		apiURLMutex.Unlock()
		return err
	}

	apiURLMutex.Lock()
	yourAPIURL = strings.TrimSpace(string(body))
	apiURLMutex.Unlock()

	return nil
}

// GetAPIURL returns the current API URL
func GetAPIURL() string {
	apiURLMutex.RLock()
	defer apiURLMutex.RUnlock()
	
	if yourAPIURL == "" {
		return fallbackAPIURL
	}
	return yourAPIURL
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// Download Functions
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// DownloadSong downloads audio from YouTube using API
func DownloadSong(ctx context.Context, link string) (string, error) {
	apiURL := GetAPIURL()
	if apiURL == "" {
		LoadAPIURL(ctx)
		apiURL = GetAPIURL()
	}

	// Extract video ID
	videoID := ExtractVideoIDFromLink(link)
	if videoID == "" || len(videoID) < 3 {
		return "", fmt.Errorf("invalid video ID")
	}

	// Setup download directory
	downloadDir := "downloads"
	os.MkdirAll(downloadDir, 0755)
	filePath := filepath.Join(downloadDir, videoID+".mp3")

	// Return if already downloaded
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	}

	// Create HTTP client
	client := &http.Client{Timeout: 60 * time.Second}

	// Request download token
	reqURL := fmt.Sprintf("%s/download?url=%s&type=audio", apiURL, videoID)
	resp, err := client.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download request failed: status %d", resp.StatusCode)
	}

	var tokenResp struct {
		DownloadToken string `json:"download_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token: %w", err)
	}

	if tokenResp.DownloadToken == "" {
		return "", fmt.Errorf("no download token received")
	}

	// Stream and save file
	streamURL := fmt.Sprintf("%s/stream/%s?type=audio", apiURL, videoID)
	req, err := http.NewRequestWithContext(ctx, "GET", streamURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Download-Token", tokenResp.DownloadToken)

	streamClient := &http.Client{Timeout: 300 * time.Second}
	streamResp, err := streamClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("stream request failed: %w", err)
	}
	defer streamResp.Body.Close()

	if streamResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("stream failed: status %d", streamResp.StatusCode)
	}

	// Save to file
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, streamResp.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// DownloadVideo downloads video from YouTube using API
func DownloadVideo(ctx context.Context, link string) (string, error) {
	apiURL := GetAPIURL()
	if apiURL == "" {
		LoadAPIURL(ctx)
		apiURL = GetAPIURL()
	}

	// Extract video ID
	videoID := ExtractVideoIDFromLink(link)
	if videoID == "" || len(videoID) < 3 {
		return "", fmt.Errorf("invalid video ID")
	}

	// Setup download directory
	downloadDir := "downloads"
	os.MkdirAll(downloadDir, 0755)
	filePath := filepath.Join(downloadDir, videoID+".mp4")

	// Return if already downloaded
	if _, err := os.Stat(filePath); err == nil {
		return filePath, nil
	}

	// Create HTTP client
	client := &http.Client{Timeout: 60 * time.Second}

	// Request download token
	reqURL := fmt.Sprintf("%s/download?url=%s&type=video", apiURL, videoID)
	resp, err := client.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download request failed: status %d", resp.StatusCode)
	}

	var tokenResp struct {
		DownloadToken string `json:"download_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token: %w", err)
	}

	if tokenResp.DownloadToken == "" {
		return "", fmt.Errorf("no download token received")
	}

	// Stream and save file
	streamURL := fmt.Sprintf("%s/stream/%s?type=video", apiURL, videoID)
	req, err := http.NewRequestWithContext(ctx, "GET", streamURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Download-Token", tokenResp.DownloadToken)

	streamClient := &http.Client{Timeout: 600 * time.Second}
	streamResp, err := streamClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("stream request failed: %w", err)
	}
	defer streamResp.Body.Close()

	if streamResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("stream failed: status %d", streamResp.StatusCode)
	}

	// Save to file
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, streamResp.Body)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

// ExtractVideoIDFromLink extracts video ID from YouTube URL
func ExtractVideoIDFromLink(link string) string {
	// Handle different URL formats
	patterns := []string{
		`(?:youtube\.com\/watch\?v=|youtu\.be\/)([a-zA-Z0-9_-]{11})`,
		`^([a-zA-Z0-9_-]{11})$`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(link)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// Try extracting from query parameter
	if strings.Contains(link, "v=") {
		parts := strings.Split(link, "v=")
		if len(parts) > 1 {
			videoID := strings.Split(parts[1], "&")[0]
			return videoID
		}
	}

	// Return as-is if it looks like a video ID
	if len(link) == 11 {
		return link
	}

	return ""
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// YouTube Handler
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// YouTubeHandler handles YouTube operations
type YouTubeHandler struct {
	downloadDir string
}

// NewYouTubeHandler creates a new YouTube handler
func NewYouTubeHandler() *YouTubeHandler {
	return &YouTubeHandler{
		downloadDir: "downloads",
	}
}



// Download downloads audio/video from YouTube
func (y *YouTubeHandler) Download(ctx context.Context, link string, isVideoID, isVideo bool) (string, error) {
	ytURL := link
	if isVideoID {
		ytURL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", link)
	}

	var filePath string
	var err error

	// Try API download first
	if isVideo {
		filePath, err = DownloadVideo(ctx, ytURL)
	} else {
		filePath, err = DownloadSong(ctx, ytURL)
	}

	if err != nil {
		// Fallback to yt-dlp
		return y.downloadFallback(ctx, ytURL, isVideo)
	}

	return filePath, nil
}

// downloadFallback uses yt-dlp as fallback
func (y *YouTubeHandler) downloadFallback(ctx context.Context, ytURL string, isVideo bool) (string, error) {
	videoID := ExtractVideoIDFromLink(ytURL)
	if videoID == "" {
		return "", fmt.Errorf("invalid URL")
	}

	os.MkdirAll(y.downloadDir, 0755)

	var ext string
	var opts []string

	if isVideo {
		ext = "mp4"
		opts = []string{
			"-f", "best[height<=720]",
			"-o", filepath.Join(y.downloadDir, videoID+".%(ext)s"),
			ytURL,
		}
	} else {
		ext = "mp3"
		opts = []string{
			"-x",
			"--audio-format", "mp3",
			"-o", filepath.Join(y.downloadDir, videoID+".%(ext)s"),
			ytURL,
		}
	}

	cmd := exec.CommandContext(ctx, "yt-dlp", opts...)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w", err)
	}

	filePath := filepath.Join(y.downloadDir, videoID+"."+ext)
	return filePath, nil
}

// GetData gets video information from YouTube
func (y *YouTubeHandler) GetData(ctx context.Context, query string, single bool, limit int) ([]VideoInfo, error) {
	// This would use YouTube search API or scraping
	// For now, return placeholder
	// In production, implement proper YouTube search
	return []VideoInfo{}, nil
}

// FormatLink formats a link from video ID or URL
func (y *YouTubeHandler) FormatLink(link string, isVideoID bool) string {
	if isVideoID {
		return fmt.Sprintf("https://www.youtube.com/watch?v=%s", link)
	}
	return link
}

// ShellCmd executes a shell command
func ShellCmd(ctx context.Context, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// GetLyrics gets song lyrics (placeholder)
func (y *YouTubeHandler) GetLyrics(ctx context.Context, song, artist string) (map[string]string, error) {
	// In production, integrate with Genius API or similar
	return map[string]string{
		"title":  song,
		"lyrics": "Lyrics not available",
		"image":  "",
	}, nil
}

// Cleanup removes downloaded files
func (y *YouTubeHandler) Cleanup(filePath string) error {
	if filePath != "" && strings.HasPrefix(filePath, y.downloadDir) {
		return os.Remove(filePath)
	}
	return nil
}

// CleanupAll removes all files in download directory
func (y *YouTubeHandler) CleanupAll() error {
	return os.RemoveAll(y.downloadDir)
}

// GetVideoInfo gets detailed video information
func (y *YouTubeHandler) GetVideoInfo(ctx context.Context, videoID string) (*VideoInfo, error) {
	// Use yt-dlp to get info
	cmd := exec.CommandContext(ctx, "yt-dlp", "-j", "--no-warnings",
		fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID))
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get video info: %w", err)
	}

	var info struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Duration  int    `json:"duration"`
		Channel   string `json:"channel"`
		ViewCount int    `json:"view_count"`
		Thumbnail string `json:"thumbnail"`
	}

	if err := json.Unmarshal(output, &info); err != nil {
		return nil, err
	}

	return &VideoInfo{
		ID:        info.ID,
		Title:     info.Title,
		Duration:  SecsToMins(info.Duration),
		Channel:   info.Channel,
		Views:     fmt.Sprintf("%d", info.ViewCount),
		Link:      fmt.Sprintf("https://www.youtube.com/watch?v=%s", info.ID),
		Thumbnail: info.Thumbnail,
	}, nil
}

// IsPlaylistURL checks if URL is a playlist
func IsPlaylistURL(url string) bool {
	return strings.Contains(url, "playlist?list=") || strings.Contains(url, "&list=")
}

// ExtractPlaylistID extracts playlist ID from URL
func ExtractPlaylistID(url string) string {
	re := regexp.MustCompile(`[?&]list=([a-zA-Z0-9_-]+)`)
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// Global YouTube handler
var YTube = NewYouTubeHandler()

// Initialize API URL on package load
func init() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		LoadAPIURL(ctx)
	}()
}
