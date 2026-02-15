package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	
	fmt.Println("T E A M    H E L L B O T   ! !")
	fmt.Println("Hello!! Welcome to HellBot Session Generator\n")
	fmt.Println("Human Verification Required !!")
	
	// Human verification
	for {
		verify := rand.Intn(50) + 1
		fmt.Printf("Enter %d to continue: ", verify)
		
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		okvai, err := strconv.Atoi(strings.TrimSpace(input))
		
		if err == nil && okvai == verify {
			fmt.Println()
			generateTelethonSession()
			break
		} else {
			fmt.Println("Verification Failed! Try Again:")
		}
	}
}

func generateTelethonSession() {
	reader := bufio.NewReader(os.Stdin)
	
	fmt.Println("✨ Gogram Session For HellBot Music!")
	fmt.Println()
	
	// Get API credentials
	fmt.Print("Enter APP ID here: ")
	appIDStr, _ := reader.ReadString('\n')
	appIDStr = strings.TrimSpace(appIDStr)
	appID, err := strconv.Atoi(appIDStr)
	if err != nil {
		fmt.Println("❌ Invalid APP ID!")
		os.Exit(1)
	}
	
	fmt.Print("\nEnter API HASH here: ")
	apiHash, _ := reader.ReadString('\n')
	apiHash = strings.TrimSpace(apiHash)
	
	fmt.Println()
	fmt.Println("📱 Please login to your Telegram account...")
	fmt.Println("Note: Use the account you want to use as the assistant/userbot for the music bot.")
	fmt.Println()
	
	// Create client
	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:   int32(appID),
		AppHash: apiHash,
	})
	if err != nil {
		fmt.Printf("❌ Failed to create client: %v\n", err)
		os.Exit(1)
	}
	
	// Start login (this will automatically prompt for phone, OTP, etc.)
	err = client.Start()
	if err != nil {
		fmt.Printf("❌ Login failed: %v\n", err)
		os.Exit(1)
	}
	
	// Export session
	sessionString := client.ExportStringSession()
	if sessionString == "" {
		fmt.Println("❌ Failed to generate session!")
		os.Exit(1)
	}
	
	// Display session
	fmt.Println()
	fmt.Println("✅ Session generated successfully!")
	fmt.Println()
	fmt.Println("🔐 Your HellBot Gogram Session String:")
	fmt.Println("============================================================")
	fmt.Println(sessionString)
	fmt.Println("============================================================")
	fmt.Println()
	
	// Try to send to saved messages
	me, err := client.GetMe()
	if err == nil {
		message := fmt.Sprintf(
			"**#HELLBOT #GOGRAM #MUSIC_BOT**\n\n"+
				"**Session String:**\n"+
				"`%s`\n\n"+
				"**⚠️ Keep this session string private!**\n"+
				"Add this to your bot's environment variables as `STRING_SESSION`",
			sessionString,
		)
		
		_, err = client.SendMessage(me.ID, message, nil)
		if err == nil {
			fmt.Println("📩 Session string also sent to your Telegram Saved Messages!")
			fmt.Println()
		}
	}
	
	fmt.Println("📝 Copy the session string above and add it to your .env file or environment variables as:")
	fmt.Println("STRING_SESSION=<your_session_string>")
	fmt.Println()
	
	// Disconnect
	client.Disconnect()
}
