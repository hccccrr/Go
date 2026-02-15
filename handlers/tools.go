package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	tg "github.com/amarnathcjd/gogram/telegram"
	"shizumusic/config"
	"shizumusic/core"
	"shizumusic/utils"
	"shizumusic/helpers"
)

// TEXTS templates from utils
var TEXTS = helpers.TextTemplates

// RegisterToolsHandlers registers tools command handlers
func RegisterToolsHandlers(client *core.Client) {
	// Telegraph/TGM upload
	client.BotClient.AddMessageHandler("/tgm", func(m *tg.NewMessage) error {
		return core.UserOnly(handleTelegraph(client))(m)
	})
	
	client.BotClient.AddMessageHandler("/telegraph", func(m *tg.NewMessage) error {
		return core.UserOnly(handleTelegraph(client))(m)
	})

	// Get group link
	client.BotClient.AddMessageHandler("/gclink", func(m *tg.NewMessage) error {
		return core.UserOnly(handleGCLink(client))(m)
	})
}

func handleTelegraph(client *core.Client) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		// Check if reply
		if m.ReplyToMsgID == 0 {
			return m.Reply("**Please reply to a media to upload**")
		}

		// Get replied message
		replied, err := client.BotClient.GetMessages(m.Chat.ID, []int32{int32(m.ReplyToMsgID)})
		if err != nil || len(replied) == 0 {
			return m.Reply("**Failed to get replied message!**")
		}

		repliedMsg := replied[0]
		
		// Check for media
		hasMedia := repliedMsg.Photo != nil || repliedMsg.Video != nil || repliedMsg.Document != nil
		if !hasMedia {
			return m.Reply("**Please reply to a valid media file**")
		}

		msg, _ := m.Reply("**⏳ Hold on baby....♡**")

		// Download media
		msg.Edit("**📥 Downloading media...**")
		localPath, err := client.BotClient.DownloadMedia(repliedMsg, &tg.DownloadOptions{
			FileName: fmt.Sprintf("upload_%d", time.Now().Unix()),
		})
		
		if err != nil {
			return msg.Edit(fmt.Sprintf("**❌ Download failed**\n\n`%s`", err.Error()))
		}

		defer os.Remove(localPath)

		// Check file size (200MB limit)
		fileInfo, _ := os.Stat(localPath)
		if fileInfo.Size() > 200*1024*1024 {
			return msg.Edit("**Please provide a media file under 200MB.**")
		}

		// Upload to catbox
		msg.Edit("**📤 Uploading to telegraph...**")
		success, uploadURL := uploadToCatbox(localPath)

		if success {
			// Create button
			buttons := [][]tg.ButtonObject{
				{
					{Text: "✨ Tap to open link ✨", URL: uploadURL},
				},
			}

			return msg.Edit(
				fmt.Sprintf("**🌐 | [👉Your link tap here👈](%s)**", uploadURL),
				&tg.MediaOptions{Buttons: buttons},
			)
		}

		return msg.Edit(
			fmt.Sprintf("**An error occurred while uploading your file**\n\n`%s`", uploadURL),
		)
	}
}

func handleGCLink(client *core.Client) core.HandlerFunc {
	return func(m *tg.NewMessage) error {
		if config.Cfg.IsBanned(m.From.ID) {
			return nil
		}

		// Get chat ID
		chatID := m.Chat.ID
		parts := strings.Fields(m.Text)
		
		if len(parts) > 1 {
			// Try to parse custom chat ID
			// chatID = parseChatID(parts[1])
		}

		// Export invite link
		link, err := client.BotClient.ExportChatInviteLink(chatID)
		if err != nil {
			if strings.Contains(err.Error(), "admin") {
				return m.Reply("❌ **Bot needs 'Invite Users via Link' permission in that group.**")
			}
			return m.Reply(fmt.Sprintf("**ERROR:** `%s`", err.Error()))
		}

		// Send with auto-delete
		sent, _ := m.Reply(fmt.Sprintf(
			"**🔗 Group Invite Link:**\n\n%s\n\n_This message will auto-delete in 30 seconds._",
			link,
		))

		// Auto delete after 30 seconds
		go func() {
			time.Sleep(30 * time.Second)
			sent.Delete()
			m.Delete()
		}()

		return nil
	}
}

// uploadToCatbox uploads file to catbox.moe
func uploadToCatbox(filePath string) (bool, string) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Sprintf("ERROR: Failed to open file - %s", err.Error())
	}
	defer file.Close()

	// Create multipart form
	body := &strings.Builder{}
	writer := multipart.NewWriter(body)

	// Add reqtype field
	writer.WriteField("reqtype", "fileupload")
	writer.WriteField("json", "true")

	// Add file
	part, err := writer.CreateFormFile("fileToUpload", filePath)
	if err != nil {
		return false, fmt.Sprintf("ERROR: Failed to create form - %s", err.Error())
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return false, fmt.Sprintf("ERROR: Failed to copy file - %s", err.Error())
	}

	writer.Close()

	// Make request
	req, err := http.NewRequest("POST", "https://catbox.moe/user/api.php", strings.NewReader(body.String()))
	if err != nil {
		return false, fmt.Sprintf("ERROR: Failed to create request - %s", err.Error())
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("ERROR: Request failed - %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, fmt.Sprintf("ERROR: HTTP %d", resp.StatusCode)
	}

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Sprintf("ERROR: Failed to read response - %s", err.Error())
	}

	return true, strings.TrimSpace(string(responseBody))
}
