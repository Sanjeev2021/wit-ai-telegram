package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Krognol/go-wolfram"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/tidwall/gjson"
	witai "github.com/wit-ai/wit-go"
)

var wolframClient *wolfram.Client

func main() {
	godotenv.Load(".env")

	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	client := witai.NewClient(os.Getenv("WIT_AI_TOKEN"))
	wolframClient = &wolfram.Client{AppID: os.Getenv("WOLFRAM_APP_ID")}

	bot, err := tgbotapi.NewBotAPI(telegramBotToken)
	if err != nil {
		log.Fatal(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, " ")
			command := update.Message.Command()
			switch command {
			case "ask":
				query := update.Message.CommandArguments()
				msg.Text = askWolfram(client, query)
			default:
				msg.Text = "I'm sorry, I don't understand your command."
			}
			response, err := bot.Send(msg)
			if err != nil {
				log.Println("Error sending message:", err)
			} else {
				log.Println("Message sent successfully. Message ID:", response.MessageID)
			}
		}
	}
}

func askWolfram(client *witai.Client, query string) string {
	msg, _ := client.Parse(&witai.MessageRequest{
		Query: query,
	})
	// MARSHALIDENT WILL IDENT THE OUTPUT BASICALLY GIVE IT FROM THE NEW LINE
	data, _ := json.MarshalIndent(msg, "", "  ")
	rough := string(data[:])
	value := gjson.Get(rough, "entities.wolfram_search_query.0.value")
	answer := value.String()

	if answer == "" {
		return "I'm sorry, I couldn't understand the question"
	}

	res, err := wolframClient.GetSpokentAnswerQuery(answer, wolfram.Metric, 1000)
	if err != nil {
		return fmt.Sprintf("There was an error: %s", err)
	}

	return res
}
