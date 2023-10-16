package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var wrongFmtError = errors.New("Неправильный формат команды")

var (
	api API
	bot *tgbotapi.BotAPI
)

func parseTopic(s string) (Concern, error) {
	s = strings.TrimSpace(s)
	chanName, topicName, found := strings.Cut(s, " ")
	if !found {
		return Concern{}, wrongFmtError
	}
	chanName = strings.Trim(chanName, "\n ")
	if !strings.HasPrefix(chanName, "@") {
		return Concern{}, wrongFmtError
	}
	return Concern{
		Channel: chanName[1:],
		Topic:   topicName,
	}, nil
}

//echo $TOPIC_KEEPER_TOKEN
//export TOPIC_KEEPER_TOKEN="6638697091:AAHhpaS-rXlgWXHQzlfa3tAGUoRctKp8n2Q"

var users map[string]int64

func main() {
	users = map[string]int64{}
	token := os.Getenv("TOPIC_KEEPER_TOKEN")

	var err error
	bot, err = tgbotapi.NewBotAPI(token)
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
			tgbotapi.NewKeyboardButton("/help"),
		),
	)

	api = basicAPI{}
	for update := range updates {
		if update.ChannelPost != nil {
			username := update.ChannelPost.Chat.UserName
			msg := update.ChannelPost.Text
			resp, err := api.postMessage(username, msg)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			for _, e := range resp {
				sendMessage(e.User, e.Summary)
			}
			continue
		}
		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Выберите команду:")
		msg.ReplyMarkup = keyboard

		updText := strings.Trim(update.Message.Text, "\n ")
		uname := update.Message.Chat.UserName
		users[uname] = update.Message.Chat.ID

		switch updText {
		case "/start":
			handleStart(uname)
		case "/view":
			handleView(uname)
		case "/help":
			handleHelp(uname)
		default:
			if strings.HasPrefix(updText, "/add") {
				handleAdd(uname, updText)
			} else if strings.HasPrefix(updText, "/remove") {
				handleRemove(uname, updText)
			} else {
				handleUnknownCommand(uname)
			}
		}
	}
}

func sendMessage(username string, text string) {
	userId, ok := users[username]
	if !ok {
		panic("!ok")
	}
	msg := tgbotapi.NewMessage(userId, text)
	_, err := bot.Send(msg)
	if err != nil {
		log.Println(err.Error())
	}
}

func handleStart(username string) {
	reply := "Привет! Бот предназначен для помощи в быстром и эффективном поиске нужной информации в чатах и темах на основе предоставленного списка."
	sendMessage(username, reply)
	handleHelp(username)
}

func handleHelp(username string) {
	reply := "\n Мой набор команд включает в себя следующие опции:" + "\n" +
		"/view - для просмотра доступных каналов и связанных с ними тем. \n \n" +
		"/remove <@название канала> <слово> - удаляет указанное слово из списка для поиска в конкретном канале.\n \n " +
		"/add <@название канала> <слово> - добавляет указанное слово в список для поиска в конкретном канале. \n \n " +
		"Эти команды помогут вам управлять списком тем и слов для поиска, чтобы быстро находить нужную информацию в чатах."
	sendMessage(username, reply)
}

func handleView(username string) {
	resp, err := api.viewTopics(username)
	if err != nil {
		sendMessage(username, err.Error())
		return
	}
	sendMessage(username, fmt.Sprintln(resp))
}

func handleAdd(username string, msg string) {
	after, _ := strings.CutPrefix(msg, "/add")
	topic, err := parseTopic(after)
	fmt.Println(topic)
	if err != nil {
		sendMessage(username, err.Error())
		return
	}
	if err := api.addTopic(username, topic); err != nil {
		sendMessage(username, err.Error())
		return
	}
	sendMessage(username, "Топик добавлен!")
}

func handleRemove(username, msg string) {
	after, _ := strings.CutPrefix(msg, "/remove")
	topic, err := parseTopic(after)
	fmt.Println(topic)
	if err != nil {
		sendMessage(username, err.Error())
		return
	}
	if err := api.removeTopic(username, topic); err != nil {
		sendMessage(username, err.Error())
		return
	}
	sendMessage(username, "Топик удалён!")
}

func handleUnknownCommand(username string) {
	reply := "Я не понимаю вашей команды. Воспользуйтесь \n /start \n /view \n /remove <name> <topic> \n /add <name> <topic>"
	sendMessage(username, reply)
}
