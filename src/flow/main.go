package main

import (
	_ "embed"
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"hash/fnv"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var wrongFmtError = errors.New("Неправильный формат команды")

const (
	NWorkers      = 30
	BaseCap       = 20
	summaryLength = 50
)

//go:embed db_config.yml
var rawDBConfig []byte

type historyRequest struct {
	user string
}

type workEvent struct {
	application    Application
	channel        string
	channelID      string
	privateChannel bool
	text           string
	link           string
	messageID      string
	historyRequest *historyRequest
}

var (
	vkToken          string
	api              API
	bot              *tgbotapi.BotAPI
	dataBase         LocalStorage
	sendChan         chan Message
	openAIkey        string
	workChans        []chan workEvent
	telegramListener UpdatesListener
	vkListener       VKHandler
)

func parseTopic(s string) (Concern, error) {
	s = strings.TrimSpace(s)
	chanName, topicName, found := strings.Cut(s, " ")
	if !found {
		return Concern{}, wrongFmtError
	}
	chanName = strings.Trim(chanName, "\n ")
	channelName, err := parseChannelName(chanName)
	if err != nil {
		return Concern{}, err
	}

	return Concern{
		Channel: channelName,
		Topic:   topicName,
	}, nil
}

func parseChannelName(s string) (string, error) {
	chanName := strings.Trim(s, "\n ")
	if strings.HasPrefix(chanName, "@") {
		return chanName[1:], nil
	} else if strings.HasPrefix(chanName, "https://t.me/") {
		return chanName[len("https://t.me/"):], nil
	} else {
		return "", wrongFmtError
	}
}

//echo $TOPIC_KEEPER_TOKEN
//export TOPIC_KEEPER_TOKEN="6638697091:AAHhpaS-rXlgWXHQzlfa3tAGUoRctKp8n2Q"

//echo $TOPIC_KEEPER_OPENAI_TOKEN
//export TOPIC_KEEPER_OPENAI_TOKEN=""

const format = `In application: %s 
Topic was detected: [%s]
In channel: %s
Summary: %s
link: %s
`

func worker(workChan chan workEvent) {
	for update := range workChan {
		channel := update.channel
		msg := update.text
		application := update.application

		if update.application == VK {
			channel = update.channelID
		}

		if found, err := dataBase.containsChannel(channel, application); !found || err != nil {
			if err != nil {
				log.Printf(err.Error())
			}
			continue
		}

		possibleTopics, err := dataBase.getTopics(channel, application)
		if err != nil {
			log.Printf(err.Error())
			continue
		}

		var foundTopics []string
		if foundTopics, err = api.analyze(msg, possibleTopics); err != nil || len(foundTopics) == 0 {
			if err != nil {
				log.Println(err.Error())
			}
			continue
		}

		var summary string
		if openAIkey != "" && len(msg) > summaryLength {
			if summary, err = api.summarize(msg, openAIkey); err != nil {
				log.Printf("error in OpenAI uisng with error: %s \n", err.Error())
				summary = summarize(msg)
			}
		} else {
			summary = summarize(msg)
		}

		sendUsers := make(map[string][]string)
		if update.historyRequest == nil {
			sendUsers, err = dataBase.getUsers(channel, foundTopics, application)
		} else {
			sendUsers[update.historyRequest.user] = foundTopics
		}

		for user, userTopics := range sendUsers {
			isPaused := false
			if update.historyRequest == nil {
				isPaused, err = dataBase.isPaused(user)
				if err != nil {
					log.Printf(err.Error())
					continue
				}
				for _, topic := range userTopics {
					if err := dataBase.setTime(user, channel, topic, application); err != nil {
						log.Printf(err.Error())
						continue
					}
				}
			}

			if update.privateChannel && update.application == Telegram {
				update.channel, err = dataBase.getTGPublicNameByID(update.channel)
				if err != nil {
					log.Println(err.Error())
					continue
				}
			}

			finalTopics := strings.Join(userTopics, ", ")
			message := Message{
				Application: update.application,
				User:        user,
				Link:        update.link,
				Channel:     update.channel,
				Topic:       finalTopics,
				Summary:     summary,
			}
			if isPaused {
				if err := dataBase.addDelayedMessage(message); err != nil {
					log.Printf(err.Error())
					continue
				}
			} else {
				sendChan <- message
			}

		}
	}
}

func sender() {
	for msg := range sendChan {
		sendNews(msg)
	}
}

func getHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func main() {
	workChans = make([]chan workEvent, NWorkers)
	sendChan = make(chan Message, BaseCap)

	for i := 0; i < NWorkers; i++ {
		workChans[i] = make(chan workEvent)
		go worker(workChans[i])
	}

	var dbConfig DBConfig
	var err error
	if err := yaml.Unmarshal(rawDBConfig, &dbConfig); err != nil {
		panic(err)
	}
	dataBase, err = NewDatabase(dbConfig,
		TablesNames{
			Messages: "messages",
			Users:    "users",
			Channels: "channels",
			VKPostID: "vk_last_post_by_public",
			TGPostID: "telegram_channel_by_id",
		},
	)
	if err != nil {
		panic(err)
	}

	token := os.Getenv("TOPIC_KEEPER_TOKEN")
	openAIkey = os.Getenv("TOPIC_KEEPER_OPENAI_TOKEN")
	vkToken = "b397ce84b397ce84b397ce8432b0819482bb397b397ce84d6cdd2ca964821d7fd266b76"

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

	vkListener = VKHandler{accessToken: "b397ce84b397ce84b397ce8432b0819482bb397b397ce84d6cdd2ca964821d7fd266b76"}
	go vkListener.handleUpdates()

	sender()

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

func sendNews(msg Message) {
	userId, err := dataBase.getID(msg.User)
	if err != nil {
		panic(err.Error())
	}
	text := fmt.Sprintf(format, msg.Application, msg.Topic, msg.Channel, msg.Summary, msg.Link)
	ans := tgbotapi.NewMessage(userId, text)
	_, err = bot.Send(ans)
	if err != nil {
		log.Println(err.Error())
	}
}

func summarize(text string) string {
	testRunes := []rune(text)

	length := len(testRunes)
	if length > summaryLength {
		length = summaryLength
	}

	return string(testRunes[:length])
}
