package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

func main() {
	fmt.Println()
	fmt.Println("╭─────────────────────────────────────╮")
	fmt.Println("│  🔐 Gogram STRING Session Generator │")
	fmt.Println("│  🚀 Pure Go • No Python Needed     │")
	fmt.Println("╰─────────────────────────────────────╯")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// API_ID
	fmt.Print("Enter API_ID: ")
	apiIDStr, _ := reader.ReadString('\n')
	apiIDStr = strings.TrimSpace(apiIDStr)

	apiID, err := strconv.Atoi(apiIDStr)
	if err != nil {
		fmt.Println("❌ Invalid API_ID")
		os.Exit(1)
	}

	// API_HASH
	fmt.Print("Enter API_HASH: ")
	apiHash, _ := reader.ReadString('\n')
	apiHash = strings.TrimSpace(apiHash)

	fmt.Println()
	fmt.Println("⏳ Connecting to Telegram...")
	fmt.Println("📱 If required, enter phone / OTP / 2FA")
	fmt.Println()

	// Create client
	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:   int32(apiID),
		AppHash: apiHash,
	})
	if err != nil {
		fmt.Printf("❌ Client error: %v\n", err)
		os.Exit(1)
	}

	// Start login
	if err := client.Start(); err != nil {
		fmt.Printf("❌ Login failed: %v\n", err)
		os.Exit(1)
	}

	// Export string session
	session := client.ExportStringSession()

	fmt.Println()
	fmt.Println("╭─────────────────────────────────────╮")
	fmt.Println("│  ✅ STRING_SESSION GENERATED        │")
	fmt.Println("╰─────────────────────────────────────╯")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(session)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Save option
	fmt.Print("💾 Save to session.txt? (y/n): ")
	save, _ := reader.ReadString('\n')
	save = strings.TrimSpace(strings.ToLower(save))

	if save == "y" || save == "yes" {
		err := os.WriteFile("session.txt", []byte(session), 0600)
		if err != nil {
			fmt.Println("❌ Failed to save file")
		} else {
			fmt.Println("✅ Saved as session.txt")
		}
	}

	fmt.Println()
	fmt.Println("✅ Done! Use this STRING_SESSION in your bot.")
}
