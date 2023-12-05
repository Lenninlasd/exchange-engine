package telegrambot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// TelegramMessage represents the structure of a Telegram message.
type TelegramMessage struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func sendTelegramMessage(botToken string, chatID int64, message string, wg *sync.WaitGroup) {
	defer wg.Done()

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	telegramMessage := TelegramMessage{
		ChatID: chatID,
		Text:   message,
	}

	// Convert the message to JSON
	messageJSON, err := json.Marshal(telegramMessage)
	if err != nil {
		fmt.Println("Error Convert the message to JSON", err)
		return
	}

	// Make the HTTP POST request to send the message
	_, err = http.Post(apiURL, "application/json", bytes.NewBuffer(messageJSON))
	if err != nil {
		fmt.Println("Error sending message:", err)
		return
	}

	return
}

func InsufficientFundsAlert() {
	var botToken = os.Getenv("ALERT_BOT_TOKEN")
	var chatIDEnv = os.Getenv("TELEGRAM_CHAT_ID")

	chatID, err := strconv.ParseInt(chatIDEnv, 10, 64)
	if err != nil {
		fmt.Println("Error on chatID env variable", err)
	}

	message := "WITHDRAW INSUFFICIENT FUNDS"
	var wg sync.WaitGroup
	wg.Add(1)
	go sendTelegramMessage(botToken, chatID, message, &wg)
	wg.Wait()
}
