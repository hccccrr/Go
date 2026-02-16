package utils

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/fogleman/gg"
	"github.com/nfnt/resize"
)

// Thumbnail handles thumbnail generation for music tracks
type Thumbnail struct {
	baseURL string
	xcbSVG  string // Base64 encoded SVG template
}

// NewThumbnail creates a new Thumbnail generator
func NewThumbnail() *Thumbnail {
	return &Thumbnail{
		baseURL: "https://www.youtube.com/watch?v=",
		xcbSVG:  getBaseThumbnailTemplate(),
	}
}

// Generate creates a thumbnail image for a video
// Returns the file path of generated thumbnail
func (t *Thumbnail) Generate(width, height int, videoID string) string {
	if videoID == "" || videoID == "telegram" {
		return ""
	}

	// Generate thumbnail with video ID
	filePath, err := t.generateThumbnail(width, height, videoID)
	if err != nil {
		return ""
	}

	return filePath
}

// generateThumbnail creates the actual thumbnail image
func (t *Thumbnail) generateThumbnail(width, height int, videoID string) (string, error) {
	// Download video thumbnail from YouTube
	thumbnailURL := fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", videoID)
	thumbnailImg, err := t.downloadImage(thumbnailURL)
	if err != nil {
		// Try hqdefault as fallback
		thumbnailURL = fmt.Sprintf("https://img.youtube.com/vi/%s/hqdefault.jpg", videoID)
		thumbnailImg, err = t.downloadImage(thumbnailURL)
		if err != nil {
			return "", err
		}
	}

	// Resize thumbnail
	resizedThumb := resize.Resize(uint(width), uint(height), thumbnailImg, resize.Lanczos3)

	// Create base image (modern design)
	baseWidth := 1280
	baseHeight := 720
	dc := gg.NewContext(baseWidth, baseHeight)

	// Background gradient (dark blue to purple)
	t.drawGradientBackground(dc, baseWidth, baseHeight)

	// Add thumbnail with rounded corners
	thumbX := 95.0
	thumbY := 165.0
	t.drawRoundedImage(dc, resizedThumb, thumbX, thumbY, 10)

	// Add text overlay
	t.addTextOverlay(dc, videoID, baseWidth, baseHeight)

	// Save to file
	randKey := fmt.Sprintf("thumb_%s_%d", videoID, rand.Int63())
	outputPath := fmt.Sprintf("%s.png", randKey)

	if err := dc.SavePNG(outputPath); err != nil {
		return "", err
	}

	return outputPath, nil
}

// downloadImage downloads an image from URL
func (t *Thumbnail) downloadImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: status %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	return img, err
}

// drawGradientBackground draws a gradient background
func (t *Thumbnail) drawGradientBackground(dc *gg.Context, width, height int) {
	// Create gradient from dark blue to purple
	for y := 0; y < height; y++ {
		ratio := float64(y) / float64(height)
		
		// Start color (dark blue): #1a1a2e
		r1, g1, b1 := 26.0, 26.0, 46.0
		
		// End color (purple): #16213e
		r2, g2, b2 := 22.0, 33.0, 62.0
		
		r := r1 + (r2-r1)*ratio
		g := g1 + (g2-g1)*ratio
		b := b1 + (b2-b1)*ratio
		
		dc.SetRGB(r/255, g/255, b/255)
		dc.DrawRectangle(0, float64(y), float64(width), 1)
		dc.Fill()
	}
}

// drawRoundedImage draws an image with rounded corners
func (t *Thumbnail) drawRoundedImage(dc *gg.Context, img image.Image, x, y, radius float64) {
	bounds := img.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())

	// Create rounded rectangle path
	dc.DrawRoundedRectangle(x, y, width, height, radius)
	dc.Clip()
	dc.DrawImage(img, int(x), int(y))
	dc.ResetClip()
}

// addTextOverlay adds text information to the thumbnail
func (t *Thumbnail) addTextOverlay(dc *gg.Context, videoID string, width, height int) {
	// For now, just add a simple "Now Playing" text
	// In production, you'd fetch video details and add them
	
	dc.SetRGB(1, 1, 1) // White text
	
	// Title area
	titleY := float64(height) - 150
	dc.DrawString("🎵 Now Playing", float64(width)/2, titleY)
	
	// You can extend this to add:
	// - Video title
	// - Channel name
	// - Duration
	// - Views
	// - Current time
}

// GenerateWithDetails creates thumbnail with detailed information
func (t *Thumbnail) GenerateWithDetails(videoID string, info *VideoInfo) (string, error) {
	baseWidth := 1280
	baseHeight := 720
	dc := gg.NewContext(baseWidth, baseHeight)

	// Background
	t.drawGradientBackground(dc, baseWidth, baseHeight)

	// Download and add video thumbnail
	thumbnailURL := fmt.Sprintf("https://img.youtube.com/vi/%s/maxresdefault.jpg", videoID)
	thumbnailImg, err := t.downloadImage(thumbnailURL)
	if err != nil {
		thumbnailURL = fmt.Sprintf("https://img.youtube.com/vi/%s/hqdefault.jpg", videoID)
		thumbnailImg, err = t.downloadImage(thumbnailURL)
		if err != nil {
			return "", err
		}
	}

	// Resize and add thumbnail
	resizedThumb := resize.Resize(600, 400, thumbnailImg, resize.Lanczos3)
	t.drawRoundedImage(dc, resizedThumb, 95, 165, 10)

	// Add text information
	if info != nil {
		dc.SetRGB(1, 1, 1) // White text
		
		textX := 750.0
		textY := 200.0
		lineHeight := 50.0

		// Title
		dc.DrawString(fmt.Sprintf("🎵 %s", truncateString(info.Title, 30)), textX, textY)
		
		// Channel
		textY += lineHeight
		dc.DrawString(fmt.Sprintf("📺 %s", truncateString(info.Channel, 25)), textX, textY)
		
		// Duration
		textY += lineHeight
		dc.DrawString(fmt.Sprintf("⏱️ %s", info.Duration), textX, textY)
		
		// Views
		textY += lineHeight
		dc.DrawString(fmt.Sprintf("👁️ %s", info.Views), textX, textY)
		
		// Playing since (current time)
		textY += lineHeight
		loc, _ := time.LoadLocation("Asia/Kolkata")
		currentTime := time.Now().In(loc)
		timeStr := currentTime.Format("15:04:05")
		dc.DrawString(fmt.Sprintf("🕐 Playing since: %s (India Time)", timeStr), textX, textY)
	}

	// Save
	randKey := fmt.Sprintf("thumb_%s_%d", videoID, rand.Int63())
	outputPath := fmt.Sprintf("%s.png", randKey)

	if err := dc.SavePNG(outputPath); err != nil {
		return "", err
	}

	return outputPath, nil
}

// SimpleGenerate creates a simple thumbnail with base64 template
func (t *Thumbnail) SimpleGenerate(videoID string) string {
	// This uses the base64 SVG template approach
	// For production use, implement proper SVG rendering
	return t.Generate(600, 400, videoID)
}

// Helper functions

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// getBaseThumbnailTemplate returns the base thumbnail template
// This would be the base64 SVG in production
func getBaseThumbnailTemplate() string {
	// Placeholder - in production this would be the actual base64 SVG
	return ""
}

// CreateRoundedMask creates a rounded rectangle mask
func createRoundedMask(width, height, radius int) *image.Alpha {
	mask := image.NewAlpha(image.Rect(0, 0, width, height))
	dc := gg.NewContext(width, height)
	dc.DrawRoundedRectangle(0, 0, float64(width), float64(height), float64(radius))
	dc.Fill()
	
	// Copy to mask
	bounds := dc.Image().Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, _, _, _ := dc.Image().At(x, y).RGBA()
			mask.SetAlpha(x, y, color.Alpha{uint8(r >> 8)})
		}
	}
	
	return mask
}

// ApplyRoundedCorners applies rounded corners to an image
func applyRoundedCorners(img image.Image, radius int) *image.RGBA {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	
	mask := createRoundedMask(width, height, radius)
	result := image.NewRGBA(bounds)
	
	draw.Draw(result, bounds, img, bounds.Min, draw.Src)
	draw.DrawMask(result, bounds, result, bounds.Min, mask, bounds.Min, draw.Over)
	
	return result
}

// SaveImage saves an image to file
func saveImage(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	
	return png.Encode(f, img)
}

// LoadImageFromBase64 loads an image from base64 string
func loadImageFromBase64(b64 string) (image.Image, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}
	
	return png.Decode(bytes.NewReader(data))
}

// EncodeImageToBase64 encodes an image to base64
func encodeImageToBase64(img image.Image) (string, error) {
	buf := new(bytes.Buffer)
	if err := png.Encode(buf, img); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// DownloadAndSave downloads an image and saves it
func downloadAndSave(url, path string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// Global thumbnail instance
var Thumb = NewThumbnail()
