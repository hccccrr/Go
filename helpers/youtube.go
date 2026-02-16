package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// YouTubeVideo represents a single YouTube video result
type YouTubeVideo struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Thumbnails  []string `json:"thumbnails"`
	LongDesc    string   `json:"long_desc"`
	Channel     string   `json:"channel"`
	Duration    string   `json:"duration"`
	Views       string   `json:"views"`
	PublishTime string   `json:"publish_time"`
	URLSuffix   string   `json:"url_suffix"`
}

// YouTubeSearch handles YouTube video searches
type YouTubeSearch struct {
	SearchTerms string
	MaxResults  int
	Videos      []YouTubeVideo
}

// NewYouTubeSearch creates a new YouTube search instance
func NewYouTubeSearch(searchTerms string, maxResults int) (*YouTubeSearch, error) {
	ys := &YouTubeSearch{
		SearchTerms: searchTerms,
		MaxResults:  maxResults,
	}

	if err := ys.Search(); err != nil {
		return nil, err
	}

	return ys, nil
}

// Search performs the YouTube search
func (ys *YouTubeSearch) Search() error {
	// Encode search query
	encodedSearch := url.QueryEscape(ys.SearchTerms)
	searchURL := fmt.Sprintf("https://youtube.com/results?search_query=%s", encodedSearch)

	// Fetch page
	response, err := ys.fetchPage(searchURL)
	if err != nil {
		return fmt.Errorf("failed to fetch page: %w", err)
	}

	// Parse results
	videos, err := ys.parseHTML(response)
	if err != nil {
		return fmt.Errorf("failed to parse results: %w", err)
	}

	// Limit results if needed
	if ys.MaxResults > 0 && len(videos) > ys.MaxResults {
		videos = videos[:ys.MaxResults]
	}

	ys.Videos = videos
	return nil
}

// fetchPage fetches the YouTube search page
func (ys *YouTubeSearch) fetchPage(url string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Retry until we get ytInitialData
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(url)
		if err != nil {
			if i == maxRetries-1 {
				return "", err
			}
			time.Sleep(time.Second)
			continue
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		response := string(body)
		if strings.Contains(response, "ytInitialData") {
			return response, nil
		}

		time.Sleep(time.Second)
	}

	return "", fmt.Errorf("failed to get valid YouTube page after %d retries", maxRetries)
}

// parseHTML extracts video data from YouTube page HTML
func (ys *YouTubeSearch) parseHTML(response string) ([]YouTubeVideo, error) {
	var results []YouTubeVideo

	// Find ytInitialData
	startMarker := "ytInitialData"
	start := strings.Index(response, startMarker)
	if start == -1 {
		return nil, fmt.Errorf("ytInitialData not found in response")
	}

	start += len(startMarker) + 3 // Skip "ytInitialData = "

	// Find end of JSON ("};" pattern)
	end := strings.Index(response[start:], "};")
	if end == -1 {
		return nil, fmt.Errorf("end of ytInitialData not found")
	}
	end = start + end + 1 // Include the closing brace

	jsonStr := response[start:end]

	// Parse JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Navigate through JSON structure
	contents, ok := data["contents"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid JSON structure: contents")
	}

	twoColumn, ok := contents["twoColumnSearchResultsRenderer"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid JSON structure: twoColumnSearchResultsRenderer")
	}

	primaryContents, ok := twoColumn["primaryContents"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid JSON structure: primaryContents")
	}

	sectionList, ok := primaryContents["sectionListRenderer"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid JSON structure: sectionListRenderer")
	}

	sectionContents, ok := sectionList["contents"].([]interface{})
	if !ok || len(sectionContents) == 0 {
		return nil, fmt.Errorf("invalid JSON structure: sectionListRenderer.contents")
	}

	itemSection, ok := sectionContents[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid JSON structure: itemSection")
	}

	itemSectionRenderer, ok := itemSection["itemSectionRenderer"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid JSON structure: itemSectionRenderer")
	}

	videos, ok := itemSectionRenderer["contents"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid JSON structure: videos")
	}

	// Extract video data
	for _, video := range videos {
		videoMap, ok := video.(map[string]interface{})
		if !ok {
			continue
		}

		videoRenderer, ok := videoMap["videoRenderer"].(map[string]interface{})
		if !ok {
			continue
		}

		videoData := ys.extractVideoData(videoRenderer)
		results = append(results, videoData)
		
		// Break after first video (like Python code)
		break
	}

	return results, nil
}

// extractVideoData extracts data from a videoRenderer object
func (ys *YouTubeSearch) extractVideoData(videoRenderer map[string]interface{}) YouTubeVideo {
	video := YouTubeVideo{}

	// ID
	if id, ok := videoRenderer["videoId"].(string); ok {
		video.ID = id
	}

	// Thumbnails
	if thumbnail, ok := videoRenderer["thumbnail"].(map[string]interface{}); ok {
		if thumbs, ok := thumbnail["thumbnails"].([]interface{}); ok {
			for _, thumb := range thumbs {
				if thumbMap, ok := thumb.(map[string]interface{}); ok {
					if url, ok := thumbMap["url"].(string); ok {
						video.Thumbnails = append(video.Thumbnails, url)
					}
				}
			}
		}
	}

	// Title
	if title, ok := videoRenderer["title"].(map[string]interface{}); ok {
		if runs, ok := title["runs"].([]interface{}); ok && len(runs) > 0 {
			if run, ok := runs[0].(map[string]interface{}); ok {
				if text, ok := run["text"].(string); ok {
					video.Title = text
				}
			}
		}
	}

	// Description
	if desc, ok := videoRenderer["descriptionSnippet"].(map[string]interface{}); ok {
		if runs, ok := desc["runs"].([]interface{}); ok && len(runs) > 0 {
			if run, ok := runs[0].(map[string]interface{}); ok {
				if text, ok := run["text"].(string); ok {
					video.LongDesc = text
				}
			}
		}
	}

	// Channel
	if byline, ok := videoRenderer["longBylineText"].(map[string]interface{}); ok {
		if runs, ok := byline["runs"].([]interface{}); ok && len(runs) > 0 {
			if run, ok := runs[0].(map[string]interface{}); ok {
				if text, ok := run["text"].(string); ok {
					video.Channel = text
				}
			}
		}
	}

	// Duration
	if length, ok := videoRenderer["lengthText"].(map[string]interface{}); ok {
		if simpleText, ok := length["simpleText"].(string); ok {
			video.Duration = simpleText
		}
	}

	// Views
	if viewCount, ok := videoRenderer["viewCountText"].(map[string]interface{}); ok {
		if simpleText, ok := viewCount["simpleText"].(string); ok {
			video.Views = simpleText
		}
	}

	// Publish time (simplified - would need pytube equivalent for exact date)
	if publishTime, ok := videoRenderer["publishedTimeText"].(map[string]interface{}); ok {
		if simpleText, ok := publishTime["simpleText"].(string); ok {
			video.PublishTime = simpleText
		}
	}
	if video.PublishTime == "" {
		video.PublishTime = "Unknown"
	}

	// URL suffix
	if nav, ok := videoRenderer["navigationEndpoint"].(map[string]interface{}); ok {
		if cmd, ok := nav["commandMetadata"].(map[string]interface{}); ok {
			if webCmd, ok := cmd["webCommandMetadata"].(map[string]interface{}); ok {
				if urlSuffix, ok := webCmd["url"].(string); ok {
					video.URLSuffix = urlSuffix
				}
			}
		}
	}

	return video
}

// ToDict returns videos as slice (clears cache if specified)
func (ys *YouTubeSearch) ToDict(clearCache bool) []YouTubeVideo {
	result := ys.Videos
	if clearCache {
		ys.Videos = nil
	}
	return result
}

// ToJSON returns videos as JSON string (clears cache if specified)
func (ys *YouTubeSearch) ToJSON(clearCache bool) (string, error) {
	data := map[string]interface{}{
		"videos": ys.Videos,
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	if clearCache {
		ys.Videos = nil
	}

	return string(jsonBytes), nil
}

// GetVideoURL returns full YouTube URL for a video
func (ys *YouTubeSearch) GetVideoURL(videoID string) string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
}

// FormatDuration converts YouTube duration format to human readable
func FormatDuration(duration string) string {
	// Already formatted by YouTube (e.g., "3:45" or "1:23:45")
	return duration
}

// FormatViews converts view count to readable format
func FormatViews(views string) string {
	// Remove "views" text and clean up
	views = strings.TrimSpace(views)
	views = strings.Replace(views, " views", "", -1)
	views = strings.Replace(views, " view", "", -1)
	return views
}

// ExtractVideoID extracts video ID from YouTube URL
func ExtractVideoID(urlStr string) string {
	// Pattern: https://www.youtube.com/watch?v=VIDEO_ID
	// or youtu.be/VIDEO_ID
	patterns := []string{
		`(?:youtube\.com/watch\?v=|youtu\.be/)([a-zA-Z0-9_-]{11})`,
		`youtube\.com/embed/([a-zA-Z0-9_-]{11})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(urlStr)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// If already just the ID
	if len(urlStr) == 11 {
		return urlStr
	}

	return ""
}

// SearchYouTube is a convenience function to search YouTube
func SearchYouTube(query string, maxResults int) ([]YouTubeVideo, error) {
	ys, err := NewYouTubeSearch(query, maxResults)
	if err != nil {
		return nil, err
	}
	return ys.Videos, nil
}
