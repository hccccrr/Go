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
	Link        string   `json:"link"`
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
	encodedSearch := url.QueryEscape(ys.SearchTerms)
	searchURL := fmt.Sprintf("https://www.youtube.com/results?search_query=%s", encodedSearch)

	response, err := ys.fetchPage(searchURL)
	if err != nil {
		return fmt.Errorf("failed to fetch page: %w", err)
	}

	videos, err := ys.parseHTML(response)
	if err != nil {
		return fmt.Errorf("failed to parse results: %w", err)
	}

	if ys.MaxResults > 0 && len(videos) > ys.MaxResults {
		videos = videos[:ys.MaxResults]
	}

	ys.Videos = videos
	return nil
}

// fetchPage fetches the YouTube search page with proper headers
func (ys *YouTubeSearch) fetchPage(reqURL string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			return "", err
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

		resp, err := client.Do(req)
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
	videos, err := ys.parseViaJSON(response)
	if err == nil && len(videos) > 0 {
		return videos, nil
	}
	return ys.parseViaRegex(response)
}

// parseViaJSON parses YouTube search results from ytInitialData JSON
func (ys *YouTubeSearch) parseViaJSON(response string) ([]YouTubeVideo, error) {
	startMarker := "var ytInitialData = "
	start := strings.Index(response, startMarker)
	if start == -1 {
		startMarker = "ytInitialData = "
		start = strings.Index(response, startMarker)
		if start == -1 {
			return nil, fmt.Errorf("ytInitialData not found")
		}
	}
	start += len(startMarker)

	jsonStr := extractJSON(response[start:])
	if jsonStr == "" {
		return nil, fmt.Errorf("could not extract JSON")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	videoRenderers := extractVideoRenderers(data)
	if len(videoRenderers) == 0 {
		return nil, fmt.Errorf("no video renderers found")
	}

	var results []YouTubeVideo
	for _, vr := range videoRenderers {
		video := ys.extractVideoData(vr)
		if video.ID != "" && video.Title != "" {
			results = append(results, video)
		}
	}

	return results, nil
}

// parseViaRegex fallback using regex to extract videoIds
func (ys *YouTubeSearch) parseViaRegex(response string) ([]YouTubeVideo, error) {
	videoIDRegex := regexp.MustCompile(`"videoId":"([a-zA-Z0-9_-]{11})"`)
	titleRegex := regexp.MustCompile(`"title":{"runs":\[{"text":"([^"]+)"`)
	durationRegex := regexp.MustCompile(`"simpleText":"(\d+:\d+(?::\d+)?)"`)

	idMatches := videoIDRegex.FindAllStringSubmatch(response, 10)
	titleMatches := titleRegex.FindAllStringSubmatch(response, 10)
	durationMatches := durationRegex.FindAllStringSubmatch(response, 10)

	seen := map[string]bool{}
	var results []YouTubeVideo

	for i, match := range idMatches {
		if len(match) < 2 {
			continue
		}
		videoID := match[1]
		if seen[videoID] {
			continue
		}
		seen[videoID] = true

		video := YouTubeVideo{
			ID:   videoID,
			Link: fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
		}

		if i < len(titleMatches) && len(titleMatches[i]) > 1 {
			video.Title = titleMatches[i][1]
		}
		if i < len(durationMatches) && len(durationMatches[i]) > 1 {
			video.Duration = durationMatches[i][1]
		}

		if video.Title != "" {
			results = append(results, video)
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no results found via regex")
	}

	return results, nil
}

// extractJSON extracts a complete JSON object from string
func extractJSON(s string) string {
	if len(s) == 0 || s[0] != '{' {
		return ""
	}

	depth := 0
	inString := false
	escaped := false

	for i, ch := range s {
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' && inString {
			escaped = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 {
				return s[:i+1]
			}
		}
	}
	return ""
}

// extractVideoRenderers walks JSON tree to find videoRenderer objects
func extractVideoRenderers(data map[string]interface{}) []map[string]interface{} {
	var renderers []map[string]interface{}

	try := func(path ...string) []interface{} {
		var cur interface{} = data
		for _, key := range path {
			m, ok := cur.(map[string]interface{})
			if !ok {
				return nil
			}
			cur = m[key]
		}
		if arr, ok := cur.([]interface{}); ok {
			return arr
		}
		return nil
	}

	sections := try("contents", "twoColumnSearchResultsRenderer", "primaryContents", "sectionListRenderer", "contents")
	for _, section := range sections {
		sMap, ok := section.(map[string]interface{})
		if !ok {
			continue
		}
		isr, ok := sMap["itemSectionRenderer"].(map[string]interface{})
		if !ok {
			continue
		}
		contents, ok := isr["contents"].([]interface{})
		if !ok {
			continue
		}
		for _, item := range contents {
			iMap, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			if vr, ok := iMap["videoRenderer"].(map[string]interface{}); ok {
				renderers = append(renderers, vr)
			}
		}
	}

	return renderers
}

// extractVideoData extracts data from a videoRenderer object
func (ys *YouTubeSearch) extractVideoData(vr map[string]interface{}) YouTubeVideo {
	video := YouTubeVideo{}

	getString := func(m map[string]interface{}, keys ...string) string {
		var cur interface{} = m
		for _, k := range keys {
			switch v := cur.(type) {
			case map[string]interface{}:
				cur = v[k]
			case []interface{}:
				if len(v) > 0 {
					cur = v[0]
				} else {
					return ""
				}
			default:
				return ""
			}
		}
		if s, ok := cur.(string); ok {
			return s
		}
		return ""
	}

	video.ID = getString(vr, "videoId")
	video.Title = getString(vr, "title", "runs", "text")
	video.Channel = getString(vr, "longBylineText", "runs", "text")
	video.Duration = getString(vr, "lengthText", "simpleText")
	video.Views = getString(vr, "viewCountText", "simpleText")
	video.PublishTime = getString(vr, "publishedTimeText", "simpleText")
	if video.PublishTime == "" {
		video.PublishTime = "Unknown"
	}
	video.URLSuffix = getString(vr, "navigationEndpoint", "commandMetadata", "webCommandMetadata", "url")

	if video.ID != "" {
		video.Link = fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.ID)
	}

	// Thumbnails
	if thumb, ok := vr["thumbnail"].(map[string]interface{}); ok {
		if thumbs, ok := thumb["thumbnails"].([]interface{}); ok {
			for _, t := range thumbs {
				if tMap, ok := t.(map[string]interface{}); ok {
					if u, ok := tMap["url"].(string); ok {
						video.Thumbnails = append(video.Thumbnails, u)
					}
				}
			}
		}
	}

	return video
}

// ToDict returns videos as slice
func (ys *YouTubeSearch) ToDict(clearCache bool) []YouTubeVideo {
	result := ys.Videos
	if clearCache {
		ys.Videos = nil
	}
	return result
}

// ToJSON returns videos as JSON string
func (ys *YouTubeSearch) ToJSON(clearCache bool) (string, error) {
	data := map[string]interface{}{"videos": ys.Videos}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	if clearCache {
		ys.Videos = nil
	}
	return string(jsonBytes), nil
}

// SearchYouTube is a convenience function to search YouTube
func SearchYouTube(query string, maxResults int) ([]YouTubeVideo, error) {
	ys, err := NewYouTubeSearch(query, maxResults)
	if err != nil {
		return nil, err
	}
	return ys.Videos, nil
}

// FormatViews cleans up view count string
func FormatViews(views string) string {
	views = strings.TrimSpace(views)
	views = strings.ReplaceAll(views, " views", "")
	views = strings.ReplaceAll(views, " view", "")
	return views
}

// ExtractVideoID extracts video ID from YouTube URL
func ExtractVideoID(urlStr string) string {
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
	if len(urlStr) == 11 {
		return urlStr
	}
	return ""
}
