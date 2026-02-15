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
	printHeader()
	generateSession()
}

func printHeader() {
	fmt.Println()
	fmt.Println("╭─────────────────────────────────────╮")
	fmt.Println("│  🎵 Gogram String Session Generator │")
	fmt.Println("│  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━  │")
	fmt.Println("│  Fast • Secure • Native Go ⚡       │")
	fmt.Println("╰─────────────────────────────────────╯")
	fmt.Println()
}

func generateSession() {
	reader := bufio.NewReader(os.Stdin)

	// Get API credentials
	fmt.Println("📋 Enter your API credentials:")
	fmt.Println("   (Get from https://my.telegram.org)")
	fmt.Println()
	
	apiID := getInput(reader, "API_ID: ")
	apiIDInt, err := strconv.Atoi(apiID)
	if err != nil {
		fmt.Println("❌ Invalid API_ID! Must be a number.")
		os.Exit(1)
	}

	apiHash := getInput(reader, "API_HASH: ")
	if apiHash == "" {
		fmt.Println("❌ API_HASH cannot be empty!")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	
	// Create Telegram client
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

	// Connect to Telegram
	fmt.Println("⏳ Connecting to Telegram servers...")
	err = client.Connect()
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("✅ Connected successfully!")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Get phone number
	fmt.Println("📱 Login to your Telegram account:")
	fmt.Println()
	phone := getInput(reader, "Phone number (with country code, e.g., +911234567890): ")
	
	if !strings.HasPrefix(phone, "+") {
		fmt.Println()
		fmt.Println("⚠️  Warning: Phone number should include country code")
		fmt.Println("   Example: +911234567890 (for India)")
		fmt.Println()
	}

	// Send verification code
	fmt.Println()
	fmt.Println("⏳ Sending verification code...")
	
	sentCode, err := client.AuthSendCode(phone, int32(apiIDInt), apiHash, &telegram.CodeSettings{})
	if err != nil {
		fmt.Printf("❌ Failed to send code: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Verification code sent to your Telegram!")
	fmt.Println()

	// Get verification code
	code := getInput(reader, "Enter verification code: ")
	
	// Extract phone code hash
	var phoneCodeHash string
	switch v := sentCode.(type) {
	case *telegram.AuthSentCodeObj:
		phoneCodeHash = v.PhoneCodeHash
	default:
		fmt.Println("❌ Failed to process verification code")
		os.Exit(1)
	}
	
	// Sign in
	fmt.Println()
	fmt.Println("⏳ Verifying code...")
	
	_, err = client.AuthSignIn(phone, phoneCodeHash, code, &telegram.EmailVerificationObj{})
	
	// Handle 2FA if needed
	if err != nil {
		if strings.Contains(err.Error(), "SESSION_PASSWORD_NEEDED") || 
		   strings.Contains(err.Error(), "password") {
			
			fmt.Println("🔐 Two-Factor Authentication detected")
			fmt.Println()
			
			password := getInput(reader, "Enter your 2FA password: ")
			
			fmt.Println()
			fmt.Println("⏳ Verifying 2FA password...")
			
			// Get password configuration
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
			
			// Check password
			_, err = client.AuthCheckPassword(inputPassword)
			if err != nil {
				fmt.Printf("❌ Wrong password! %v\n", err)
				os.Exit(1)
			}
			
		} else {
			fmt.Printf("❌ Login failed: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("✅ Login successful!")
	fmt.Println()

	// Get user info
	user, err := client.GetMe()
	if err == nil {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("👤 Logged in as: %s", user.FirstName)
		if user.LastName != "" {
			fmt.Printf(" %s", user.LastName)
		}
		if user.Username != "" {
			fmt.Printf(" (@%s)", user.Username)
		}
		fmt.Println()
		fmt.Printf("🆔 User ID: %d\n", user.ID)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()
	}

	// Export session string
	fmt.Println("⏳ Generating string session...")
	sessionString := client.ExportStringSession()

	if sessionString == "" {
		fmt.Println("❌ Failed to generate session string!")
		os.Exit(1)
	}

	// Display session
	displaySession(sessionString, apiID, apiHash)

	// Save option
	fmt.Print("💾 Save to file? (y/n): ")
	save, _ := reader.ReadString('\n')
	save = strings.TrimSpace(strings.ToLower(save))

	if save == "y" || save == "yes" {
		// Save session to file
		err := os.WriteFile("session.txt", []byte(sessionString), 0600)
		if err != nil {
			fmt.Printf("❌ Failed to save session: %v\n", err)
		} else {
			fmt.Println("✅ Session saved to: session.txt")
		}
		
		// Save .env file
		envContent := fmt.Sprintf(`# Telegram API Credentials
API_ID=%s
API_HASH=%s

# Bot Configuration
BOT_TOKEN=your_bot_token_here

# User Session
STRING_SESSION=%s

# Database
DATABASE_URL=mongodb://localhost:27017/shizumusic

# Required IDs
LOGGER_ID=-1001234567890
OWNER_ID=%d
`, apiID, apiHash, sessionString, user.ID)

		err = os.WriteFile(".env", []byte(envContent), 0600)
		if err != nil {
			fmt.Printf("⚠️  Failed to save .env file: %v\n", err)
		} else {
			fmt.Println("✅ Configuration saved to: .env")
		}
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Done! Your Gogram session is ready!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	
	// Disconnect
	client.Disconnect()
}

func displaySession(session, apiID, apiHash string) {
	fmt.Println()
	fmt.Println("╭─────────────────────────────────────╮")
	fmt.Println("│  ✅ STRING SESSION GENERATED!       │")
	fmt.Println("╰─────────────────────────────────────╯")
	fmt.Println()
	fmt.Println("📝 Your String Session:")
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println(session)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("⚠️  SECURITY WARNING:")
	fmt.Println("   • Keep this session string PRIVATE")
	fmt.Println("   • Never share with anyone")
	fmt.Println("   • Anyone with this can access your account")
	fmt.Println()
	fmt.Println("📋 Add to your .env file:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("API_ID=%s\n", apiID)
	fmt.Printf("API_HASH=%s\n", apiHash)
	fmt.Println("BOT_TOKEN=your_bot_token_here")
	fmt.Printf("STRING_SESSION=%s\n", session)
	fmt.Println("DATABASE_URL=mongodb://localhost:27017/shizumusic")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

func getInput(reader *bufio.Reader, prompt string) string {
	fmt.Print(prompt)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
