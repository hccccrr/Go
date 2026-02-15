package main

import (
	"bufio"
	"fmt"
	"os"

	tg "github.com/amarnathcjd/gogram/telegram"
)

func main() {
	apiId := int32(25742938)          // apna API_ID
	apiHash := "b35b715fe8dc0a58e8048988286fc5b6"      // apna API_HASH

	client := tg.NewClient(apiId, apiHash, tg.ClientConfig{
		Session: "",
	})

	err := client.Run()
	if err != nil {
		panic(err)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter phone number (+91xxxxxxxxxx): ")
	phone, _ := reader.ReadString('\n')

	err = client.Auth().SendCode(phone)
	if err != nil {
		panic(err)
	}

	fmt.Print("Enter OTP: ")
	code, _ := reader.ReadString('\n')

	err = client.Auth().SignIn(phone, code)
	if err != nil {
		panic(err)
	}

	session := client.ExportSession()
	fmt.Println("\n✅ gogram StringSession:\n")
	fmt.Println(session)
}
