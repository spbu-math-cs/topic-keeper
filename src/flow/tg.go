package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func tgMain() int {
	var err error
	token := os.Getenv("TOPIC_KEEPER_TOKEN")
	openAIkey = os.Getenv("TOPIC_KEEPER_OPENAI_TOKEN")
	vkToken = os.Getenv("TOPIC_KEEPER_VK_TOKEN")

	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err.Error())
	}
	if openAIkey != "" {
		log.Printf("using openAI summarizer with key: %s", openAIkey)
	}
	setBotCommands(bot)
	bot.Debug = true
	log.Printf("Authorized on account: %s\n", bot.Self.UserName)

	api = basicAPI{}

	telegramListener = newTelegramHandler(bot)
	go telegramListener.handleUpdates()

	vkListener = VKHandler{accessToken: vkToken}
	go vkListener.handleUpdates()

	sender()
	return 0
}
