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
	fmt.Println("│  OTP + 2FA ONLY • NO QR LOGIN 🔒    │")
	fmt.Println("╰─────────────────────────────────────╯")
	fmt.Println()
}

func generateSession() {
	reader := bufio.NewReader(os.Stdin)

	// API DETAILS
	apiIDStr := getInput(reader, "API_ID: ")
	apiID, err := strconv.Atoi(apiIDStr)
	if err != nil {
		fmt.Println("❌ API_ID must be a number")
		os.Exit(1)
	}

	apiHash := getInput(reader, "API_HASH: ")
	if apiHash == "" {
		fmt.Println("❌ API_HASH cannot be empty")
		os.Exit(1)
	}

	// CREATE CLIENT
	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:   int32(apiID),
		AppHash: apiHash,
	})
	if err != nil {
		fmt.Println("❌ Client error:", err)
		os.Exit(1)
	}

	// CONNECT
	fmt.Println("⏳ Connecting to Telegram...")
	if err := client.Connect(); err != nil {
		fmt.Println("❌ Connection failed:", err)
		os.Exit(1)
	}

	// 🔒 FORCE LOGOUT (ANTI-QR)
	if client.IsAuthorized() {
		fmt.Println("⚠️ Existing session found, logging out...")
		_, _ = client.AuthLogOut()
	}

	fmt.Println("✅ Connected (Phone login only)")

	// PHONE NUMBER
	phone := getInput(reader, "📱 Enter phone number (+91xxxx): ")

	// SEND OTP
	fmt.Println("⏳ Sending OTP...")
	sent, err := client.AuthSendCode(
		phone,
		int32(apiID),
		apiHash,
		&telegram.CodeSettings{},
	)
	if err != nil {
		fmt.Println("❌ OTP send failed:", err)
		os.Exit(1)
	}

	var phoneCodeHash string
	if v, ok := sent.(*telegram.AuthSentCodeObj); ok {
		phoneCodeHash = v.PhoneCodeHash
	} else {
		fmt.Println("❌ Invalid OTP response")
		os.Exit(1)
	}

	// ENTER OTP
	code := getInput(reader, "🔑 Enter OTP: ")

	// SIGN IN
	fmt.Println("⏳ Verifying OTP...")
	_, err = client.AuthSignIn(phone, phoneCodeHash, code, nil)

	// 2FA HANDLING
	if err != nil {
		if strings.Contains(err.Error(), "SESSION_PASSWORD_NEEDED") {
			fmt.Println("🔐 2FA detected")

			password := getInput(reader, "Enter 2FA password: ")

			pwdInfo, err := client.AccountGetPassword()
			if err != nil {
				fmt.Println("❌ Password info error:", err)
				os.Exit(1)
			}

			inputPwd, err := telegram.GetInputCheckPassword(password, pwdInfo)
			if err != nil {
				fmt.Println("❌ Password processing error:", err)
				os.Exit(1)
			}

			_, err = client.AuthCheckPassword(inputPwd)
			if err != nil {
				fmt.Println("❌ Wrong 2FA password:", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("❌ Login failed:", err)
			os.Exit(1)
		}
	}

	fmt.Println("✅ Login successful!")

	// USER INFO
	me, _ := client.GetMe()
	fmt.Printf("👤 Logged in as: %s (%d)\n", me.FirstName, me.ID)

	// EXPORT SESSION
	session := client.ExportStringSession()
	if session == "" {
		fmt.Println("❌ Session export failed")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("🔐 STRING SESSION:")
	fmt.Println(session)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// SAVE OPTION
	save := getInput(reader, "💾 Save session to file? (y/n): ")
	if strings.ToLower(save) == "y" {
		_ = os.WriteFile("session.txt", []byte(session), 0600)

		env := fmt.Sprintf(
			"API_ID=%d\nAPI_HASH=%s\nSTRING_SESSION=%s\nOWNER_ID=%d\n",
			apiID, apiHash, session, me.ID,
		)
		_ = os.WriteFile(".env", []byte(env), 0600)

		fmt.Println("✅ Saved: session.txt & .env")
	}

	client.Disconnect()
	fmt.Println("🎉 Done! Secure session generated.")
}

func getInput(r *bufio.Reader, msg string) string {
	fmt.Print(msg)
	t, _ := r.ReadString('\n')
	return strings.TrimSpace(t)
}
