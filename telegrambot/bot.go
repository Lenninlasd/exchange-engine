package telegrambot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// TelegramMessage represents the structure of a Telegram message.
type TelegramMessage struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func sendTelegramMessage(botToken string, chatID int64, message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	telegramMessage := TelegramMessage{
		ChatID: chatID,
		Text:   message,
	}

	// Convert the message to JSON
	messageJSON, err := json.Marshal(telegramMessage)
	if err != nil {
		return err
	}

	// Make the HTTP POST request to send the message
	_, err = http.Post(apiURL, "application/json", bytes.NewBuffer(messageJSON))
	if err != nil {
		return err
	}

	return nil
}

func InsufficientFundsAlert() {
	// Replace these values with your actual bot token, chat ID, and message
	var botToken = os.Getenv("ALERT_BOT_TOKEN")
	var chatIDEnv = os.Getenv("TELEGRAM_CHAT_ID")

	chatID, err := strconv.ParseInt(chatIDEnv, 10, 64)
	if err != nil {
		fmt.Println("Error on chatID env variable", err)
	}

	message := "WITHDRAW INSUFFICIENT FUNDS"
	err = sendTelegramMessage(botToken, chatID, message)

	if err != nil {
		fmt.Println("Failed to send message:", err)
	} else {
		fmt.Println("Message sent successfully")
	}
}
