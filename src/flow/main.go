package main

import (
	_ "embed"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var wrongFmtError = errors.New("Неправильный формат команды")

//go:embed db_config.yml
var rawDBConfig []byte

var (
	api      API
	bot      *tgbotapi.BotAPI
	dataBase LocalStorage
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

const format = `Topic was detected: [%s]
In channel: @%s
Summary: %s
link: https://t.me/%s/%d
`

func main() {
	var dbConfig DBConfig
	var err error
	if err := yaml.Unmarshal(rawDBConfig, &dbConfig); err != nil {
		panic(err)
	}
	dataBase, err = NewDatabase(dbConfig)
	if err != nil {
		panic(err)
	}

	token := os.Getenv("TOPIC_KEEPER_TOKEN")

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

			channel := update.ChannelPost.Chat.UserName
			msg := update.ChannelPost.Text

			if found, err := dataBase.containsChannel(channel); !found || err != nil {
				if err != nil {
					log.Printf(err.Error())
				}
				continue
			}

			possibleTopics, err := dataBase.getTopics(channel)
			if err != nil {
				log.Printf(err.Error())
				continue
			}

			foundTopics, summary, err := api.analyze(msg, possibleTopics)
			if err != nil {
				log.Printf(err.Error())
				continue
			}

			sendUsers, err := dataBase.getUsers(channel, foundTopics)
			for user, userTopics := range sendUsers {
				for _, topic := range userTopics {
					if err := dataBase.setTime(user, channel, topic); err != nil {
						log.Printf(err.Error())
						continue
					}
				}
				finalTopics := strings.Join(userTopics, ", ")
				ans := fmt.Sprintf(format, finalTopics, channel, summary, channel, update.ChannelPost.MessageID)

				sendMessage(user, ans)
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
		//users[uname] = update.Message.Chat.ID
		err := dataBase.addUser(uname, update.Message.Chat.ID)
		if err != nil {
			continue
		}

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
	userId, err := dataBase.getID(username)
	if err != nil {
		panic(err.Error())
	}
	msg := tgbotapi.NewMessage(userId, text)
	_, err = bot.Send(msg)
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
	topicByChan, err := dataBase.getUserInfo(username)
	if err != nil {
		sendMessage(username, err.Error())
		return
	}
	str := strings.Builder{}
	for ch, topics := range topicByChan {
		str.WriteString(fmt.Sprintf("%s:\n", ch))
		for _, topic := range topics {
			str.WriteString(fmt.Sprintf("  - %s\n", topic))
		}
		str.WriteString("\n")
	}
	sendMessage(username, str.String())
}

func handleAdd(username string, msg string) {
	after, _ := strings.CutPrefix(msg, "/add")
	concern, err := parseTopic(after)
	fmt.Println(concern)
	if err != nil {
		sendMessage(username, err.Error())
		return
	}
	if err := dataBase.add(username, concern.Channel, concern.Topic); err != nil {
		sendMessage(username, err.Error())
		return
	}
	sendMessage(username, "Топик добавлен!")
}

func handleRemove(username, msg string) {
	after, _ := strings.CutPrefix(msg, "/remove")
	concern, err := parseTopic(after)
	fmt.Println(concern)
	if err != nil {
		sendMessage(username, err.Error())
		return
	}
	if err := dataBase.removeTopic(username, concern.Channel, concern.Topic); err != nil {
		sendMessage(username, err.Error())
		return
	}
	sendMessage(username, "Топик удалён!")
}

func handleUnknownCommand(username string) {
	reply := "Я не понимаю вашей команды. Воспользуйтесь \n /start \n /view \n /remove <name> <topic> \n /add <name> <topic>"
	sendMessage(username, reply)
}
