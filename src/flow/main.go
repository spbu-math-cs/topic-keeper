package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ChannelData struct {
	Command string `json:"command"`
	Name    string `json:"name"`
	Topic   string `json:"topic"`
}

//echo $TOPIC_KEEPER_TOKEN
//export TOPIC_KEEPER_TOKEN="6638697091:AAHhpaS-rXlgWXHQzlfa3tAGUoRctKp8n2Q"

func main() {
	token := os.Getenv("TOPIC_KEEPER_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err.Error())
	}

	bot.Debug = true
	log.Printf("Authorized on account: %s\n", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/start"),
			tgbotapi.NewKeyboardButton("/view"),
			tgbotapi.NewKeyboardButton("/add"),
			tgbotapi.NewKeyboardButton("/remove"),
		),
	)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите команду:")
		msg.ReplyMarkup = keyboard
		switch update.Message.Text {
		case "/start":
			handleStart(bot, update.Message.Chat.ID)
		case "/view":
			handleView(bot, update.Message.Chat.ID)
		default:
			if strings.HasPrefix(update.Message.Text, "/remove") || strings.HasPrefix(update.Message.Text, "/add") {
				handleRemoveAdd(bot, update.Message.Chat.ID, update.Message.Text)
			} else {
				handleUnknownCommand(bot, update.Message.Chat.ID)
			}
		}
	}
}

func sendMessage(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Panic(err)
	}
}

func handleStart(bot *tgbotapi.BotAPI, chatID int64) {
	reply := "Привет! Это ботик для обработки команд /start, /view, /add и /remove."
	sendMessage(bot, chatID, reply)
}

func handleView(bot *tgbotapi.BotAPI, chatID int64) {
	//как-то взять json распарсить и вывести корректно
	reply := "Список каналов с выранными темами: \n Канал: Как стать миллионером \n Темы: новичок"
	sendMessage(bot, chatID, reply)
}

func handleRemoveAdd(bot *tgbotapi.BotAPI, chatID int64, text string) {
	args := strings.Fields(text)
	if len(args) == 3 {
		command := strings.TrimLeft(args[0], "/")
		channelData := ChannelData{Command: command, Name: args[1], Topic: args[2]}
		channelJSON, _ := json.Marshal(channelData)

		//отправка бэку
		// Отправка данных на бэкенд
		err := sendJSONToBackend(channelJSON, "add")
		if err != nil {
			sendMessage(bot, chatID, "Ошибка в полученных данных"+err.Error())
		} else {
			sendMessage(bot, chatID, "Данные успешно зафиксированы")
		}
		sendMessage(bot, chatID, "JSON данных канала: "+string(channelJSON))
	} else {
		handleUnknownCommand(bot, chatID)
	}
}

func handleUnknownCommand(bot *tgbotapi.BotAPI, chatID int64) {
	reply := "Я не понимаю вашей команды. Воспользуйтесь \n /start \n /view \n /remove <name> <topic> \n /add <name> <topic>"
	sendMessage(bot, chatID, reply)
}

func sendJSONToBackend(jsonData []byte, endpoint string) error {
	//отправка json
	return nil
}
