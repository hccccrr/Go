package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// 🎵 ShizuMusic - Complete Session Generator
// Supports: Gogram (Go), Telethon (Python), Pyrogram (Python)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func main() {
	printHeader()
	choice := showMenu()
	
	switch choice {
	case "1":
		generateGogramSession()
	case "2":
		generateTelethonSession()
	case "3":
		generatePyrogramSession()
	default:
		fmt.Println("❌ Invalid choice!")
		os.Exit(1)
	}
}

func printHeader() {
	fmt.Println()
	fmt.Println("╭─────────────────────────────────────╮")
	fmt.Println("│  🎵 ShizuMusic Session Generator    │")
	fmt.Println("│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │")
	fmt.Println("│  Supports All 3 Methods! ⭐         │")
	fmt.Println("╰─────────────────────────────────────╯")
	fmt.Println()
}

func showMenu() string {
	fmt.Println("📱 Choose Session Generation Method:")
	fmt.Println()
	fmt.Println("  1️⃣  Gogram (Native Go) ⭐ Recommended")
	fmt.Println("      • Pure Go, no dependencies")
	fmt.Println("      • Fast and secure")
	fmt.Println("      • Best performance")
	fmt.Println()
	fmt.Println("  2️⃣  Telethon (Python)")
	fmt.Println("      • Popular Python library")
	fmt.Println("      • Compatible with Gogram")
	fmt.Println()
	fmt.Println("  3️⃣  Pyrogram (Python)")
	fmt.Println("      • Modern Python library")
	fmt.Println("      • Fast and clean")
	fmt.Println()
	
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter choice (1-3): ")
	choice, _ := reader.ReadString('\n')
	return strings.TrimSpace(choice)
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// METHOD 1: GOGRAM SESSION (Native Go - Recommended!)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func generateGogramSession() {
	fmt.Println()
	fmt.Println("╭─────────────────────────────────────╮")
	fmt.Println("│  🔐 Gogram Session Generator        │")
	fmt.Println("│  (Native Go - Best Method!)         │")
	fmt.Println("╰─────────────────────────────────────╯")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Get API credentials
	apiID := getInput(reader, "Enter API_ID: ")
	apiIDInt, err := strconv.Atoi(apiID)
	if err != nil {
		fmt.Println("❌ Invalid API_ID!")
		os.Exit(1)
	}

	apiHash := getInput(reader, "Enter API_HASH: ")

	// Create Telegram client
	fmt.Println()
	fmt.Println("⏳ Creating Telegram client...")

	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:         int32(apiIDInt),
		AppHash:       apiHash,
		StringSession: "",
	})

	if err != nil {
		fmt.Printf("❌ Failed to create client: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Client created!")
	fmt.Println()

	// Connect to Telegram
	fmt.Println("⏳ Connecting to Telegram servers...")
	err = client.Connect()
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Connected to Telegram!")
	fmt.Println()

	// Login instructions
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("📱 Login Instructions:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("   1. Enter phone number with country code")
	fmt.Println("      Example: +911234567890")
	fmt.Println("   2. Enter verification code from Telegram")
	fmt.Println("   3. If 2FA enabled, enter password")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Get phone number
	phone := getInput(reader, "📱 Enter phone number (with country code): ")
	if !strings.HasPrefix(phone, "+") {
		fmt.Println("⚠️  Warning: Phone should start with + and country code")
		fmt.Println("   Example: +911234567890")
	}

	// Send code
	fmt.Println()
	fmt.Println("⏳ Sending verification code...")
	
	sentCode, err := client.AuthSendCode(phone, int32(apiIDInt), apiHash, &telegram.CodeSettings{})
	if err != nil {
		fmt.Printf("❌ Failed to send code: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Verification code sent to your Telegram account!")
	fmt.Println()

	// Get verification code
	code := getInput(reader, "📲 Enter verification code: ")
	
	// Extract phone code hash from sentCode
	var phoneCodeHash string
	switch v := sentCode.(type) {
	case *telegram.AuthSentCodeObj:
		phoneCodeHash = v.PhoneCodeHash
	default:
		fmt.Println("❌ Failed to get phone code hash")
		os.Exit(1)
	}
	
	// Sign in with code
	fmt.Println()
	fmt.Println("⏳ Verifying code...")
	
	_, err = client.AuthSignIn(phone, phoneCodeHash, code, &telegram.EmailVerificationObj{})
	
	// Check if 2FA is required
	if err != nil {
		if strings.Contains(err.Error(), "SESSION_PASSWORD_NEEDED") || 
		   strings.Contains(err.Error(), "password") {
			fmt.Println("🔐 Two-Factor Authentication enabled")
			fmt.Println()
			
			password := getInput(reader, "🔑 Enter 2FA password: ")
			
			fmt.Println()
			fmt.Println("⏳ Verifying password...")
			
			// Get password info first
			accountPassword, err := client.AccountGetPassword()
			if err != nil {
				fmt.Printf("❌ Failed to get password settings: %v\n", err)
				os.Exit(1)
			}
			
			// Compute password SRP
			inputPassword, err := telegram.GetInputCheckPassword(password, accountPassword)
			if err != nil {
				fmt.Printf("❌ Failed to compute password: %v\n", err)
				os.Exit(1)
			}
			
			_, err = client.AuthCheckPassword(inputPassword)
			if err != nil {
				fmt.Printf("❌ 2FA verification failed: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("❌ Sign in failed: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("✅ Login successful!")
	fmt.Println()

	// Get user info
	user, err := client.GetMe()
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not get user info: %v\n", err)
	} else {
		fmt.Printf("👤 Logged in as: %s %s (@%s)\n", user.FirstName, user.LastName, user.Username)
		fmt.Println()
	}

	// Export session string
	fmt.Println("⏳ Generating session string...")
	sessionString := client.ExportStringSession()

	if sessionString == "" {
		fmt.Println("❌ Failed to export session string!")
		os.Exit(1)
	}

	// Display success message
	displaySuccess(sessionString, apiID, apiHash)

	// Save option
	saveOption(reader, sessionString)

	fmt.Println()
	fmt.Println("✅ Done! Your Gogram session is ready to use!")
	fmt.Println("   Add it to your .env file and start ShizuMusic bot.")
	fmt.Println()
	
	// Disconnect
	client.Disconnect()
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// METHOD 2: TELETHON SESSION (Python)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func generateTelethonSession() {
	fmt.Println()
	fmt.Println("╭─────────────────────────────────────╮")
	fmt.Println("│  🐍 Telethon Session Generator      │")
	fmt.Println("│  (Python Method)                    │")
	fmt.Println("╰─────────────────────────────────────╯")
	fmt.Println()

	// Check Python
	if !checkPython() {
		fmt.Println("❌ Python3 not found!")
		fmt.Println("   Install: sudo apt install python3 python3-pip")
		os.Exit(1)
	}

	// Install Telethon
	fmt.Println("📦 Installing Telethon...")
	installPythonPackage("telethon")
	fmt.Println("✅ Telethon installed!")
	fmt.Println()

	// Python script for Telethon
	pythonScript := `
from telethon.sync import TelegramClient
from telethon.sessions import StringSession

print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
print("📱 Telethon Session Generator")
print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
print()

api_id = int(input("Enter API_ID: "))
api_hash = input("Enter API_HASH: ")

print()
print("⏳ Connecting to Telegram...")
print("   Follow the prompts below:")
print()

with TelegramClient(StringSession(), api_id, api_hash) as client:
    print()
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print("✅ Telethon Session Generated!")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print()
    print("📝 Your STRING_SESSION:")
    print()
    session_string = client.session.save()
    print(session_string)
    print()
    print("⚠️  IMPORTANT:")
    print("   • Save this session securely!")
    print("   • Add to .env as STRING_SESSION")
    print("   • Never share with anyone!")
    print()
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print()
    
    save = input("💾 Save to session.txt? (y/n): ").strip().lower()
    if save in ['y', 'yes']:
        with open('session.txt', 'w') as f:
            f.write(session_string)
        print("✅ Saved to session.txt")
    print()
    print("✅ Done! Compatible with Gogram and all MTProto libraries.")
`

	// Run Telethon script
	runPythonScript(pythonScript, "telethon")
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// METHOD 3: PYROGRAM SESSION (Python)
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func generatePyrogramSession() {
	fmt.Println()
	fmt.Println("╭─────────────────────────────────────╮")
	fmt.Println("│  🐍 Pyrogram Session Generator      │")
	fmt.Println("│  (Python Method)                    │")
	fmt.Println("╰─────────────────────────────────────╯")
	fmt.Println()

	// Check Python
	if !checkPython() {
		fmt.Println("❌ Python3 not found!")
		fmt.Println("   Install: sudo apt install python3 python3-pip")
		os.Exit(1)
	}

	// Install Pyrogram
	fmt.Println("📦 Installing Pyrogram + TgCrypto...")
	installPythonPackage("pyrogram", "tgcrypto")
	fmt.Println("✅ Pyrogram installed!")
	fmt.Println()

	// Python script for Pyrogram
	pythonScript := `
from pyrogram import Client
import os

print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
print("📱 Pyrogram Session Generator")
print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
print()

api_id = int(input("Enter API_ID: "))
api_hash = input("Enter API_HASH: ")

print()
print("⏳ Connecting to Telegram...")
print("   Follow the prompts below:")
print()

with Client("my_account", api_id, api_hash) as app:
    print()
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print("✅ Pyrogram Session Generated!")
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print()
    print("📝 Your STRING_SESSION:")
    print()
    session_string = app.export_session_string()
    print(session_string)
    print()
    print("⚠️  IMPORTANT:")
    print("   • Save this session securely!")
    print("   • Add to .env as STRING_SESSION")
    print("   • Never share with anyone!")
    print()
    print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    print()
    
    save = input("💾 Save to session.txt? (y/n): ").strip().lower()
    if save in ['y', 'yes']:
        with open('session.txt', 'w') as f:
            f.write(session_string)
        print("✅ Saved to session.txt")

# Cleanup
try:
    os.remove("my_account.session")
except:
    pass

print()
print("✅ Done! Compatible with Gogram and all MTProto libraries.")
`

	// Run Pyrogram script
	runPythonScript(pythonScript, "pyrogram")
}

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
// HELPER FUNCTIONS
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

func getInput(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func displaySuccess(session, apiID, apiHash string) {
	fmt.Println()
	fmt.Println("╭─────────────────────────────────────╮")
	fmt.Println("│  ✅ Session Generated Successfully! │")
	fmt.Println("╰─────────────────────────────────────╯")
	fmt.Println()
	fmt.Println("📝 Your STRING_SESSION:")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(session)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT:")
	fmt.Println("  • Keep this session string safe!")
	fmt.Println("  • Never share it with anyone!")
	fmt.Println("  • Add it to your .env file")
	fmt.Println()
	fmt.Println("📝 .env Configuration:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("API_ID=%s\n", apiID)
	fmt.Printf("API_HASH=%s\n", apiHash)
	fmt.Println("BOT_TOKEN=your_bot_token_here")
	fmt.Printf("STRING_SESSION=%s\n", session)
	fmt.Println("DATABASE_URL=mongodb://localhost:27017/shizumusic")
	fmt.Println("LOGGER_ID=-1001234567890")
	fmt.Println("OWNER_ID=your_user_id")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

func saveOption(reader *bufio.Reader, session string) {
	fmt.Print("💾 Save session to file? (y/n): ")
	save, _ := reader.ReadString('\n')
	save = strings.TrimSpace(strings.ToLower(save))

	if save == "y" || save == "yes" {
		err := os.WriteFile("session.txt", []byte(session), 0600)
		if err != nil {
			fmt.Printf("❌ Failed to save: %v\n", err)
		} else {
			fmt.Println("✅ Session saved to session.txt")
			fmt.Println("   File location: ./session.txt")
		}
	}
}

func checkPython() bool {
	cmd := exec.Command("python3", "--version")
	err := cmd.Run()
	return err == nil
}

func installPythonPackage(packages ...string) {
	args := append([]string{"install"}, packages...)
	args = append(args, "--break-system-packages", "--quiet")
	
	cmd := exec.Command("pip3", args...)
	err := cmd.Run()
	
	if err != nil {
		// Try without --break-system-packages
		args2 := append([]string{"install"}, packages...)
		args2 = append(args2, "--quiet")
		cmd2 := exec.Command("pip3", args2...)
		cmd2.Run()
	}
}

func runPythonScript(script, name string) {
	// Write script to temp file
	tmpFile := fmt.Sprintf("/tmp/%s_gen.py", name)
	err := os.WriteFile(tmpFile, []byte(script), 0644)
	if err != nil {
		fmt.Printf("❌ Failed to create script: %v\n", err)
		os.Exit(1)
	}

	// Run script with interactive mode
	cmd := exec.Command("python3", tmpFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err = cmd.Run()
	if err != nil {
		fmt.Printf("❌ Script failed: %v\n", err)
	}

	// Cleanup
	os.Remove(tmpFile)
}
