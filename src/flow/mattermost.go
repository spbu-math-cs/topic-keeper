package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/rs/zerolog"
)

type config struct {
	userName string
	teamName string
	token    string
	server   *url.URL
}

func loadConfig() config {
	var settings config

	// topic-keeper
	settings.userName = os.Getenv("MM_USERNAME")
	// MKNPI
	settings.teamName = os.Getenv("MM_TEAM")
	// ...
	settings.token = os.Getenv("MM_TOKEN")
	// http://146.185.240.118:8065
	var err error
	settings.server, err = url.Parse(os.Getenv("MM_SERVER"))
	if err != nil {
		panic(err)
	}

	return settings
}

// application struct to hold the dependencies for our bot.
type application struct {
	config config
	logger zerolog.Logger
	client *model.Client4
	user   *model.User
	team   *model.Team
}

func mattermostMain() int {
	app := &application{
		logger: zerolog.New(
			zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: time.RFC822,
			},
		).With().Timestamp().Logger(),
	}

	app.config = loadConfig()
	app.logger.Info().Str("config", fmt.Sprint(app.config)).Msg("")

	app.client = model.NewAPIv4Client(app.config.server.String())
	app.client.SetToken(app.config.token)

	if user, resp, err := app.client.GetUser("me", ""); err != nil {
		app.logger.Fatal().Err(err).Msg("Could not log in")
	} else {
		app.logger.Debug().Interface("user", user).Interface("resp", resp).Msg("")
		app.logger.Info().Msg("Logged in to mattermost")
		app.user = user
	}

	// Find and save the bot's team to app struct.
	if team, resp, err := app.client.GetTeamByName(app.config.teamName, ""); err != nil {
		app.logger.Fatal().Err(err).Msg("Could not find team. Is this bot a member ?")
	} else {
		app.logger.Debug().Interface("team", team).Interface("resp", resp).Msg("")
		app.team = team
	}

	var wsClient *model.WebSocketClient
	var err error
	fails := 0
	for {
		wsClient, err = model.NewWebSocketClient4(
			fmt.Sprintf("ws://%s", app.config.server.Host+app.config.server.Path),
			app.client.AuthToken,
		)
		if err != nil {
			app.logger.Warn().Err(err).Msg("Mattermost websocket disconnected, retrying")
			fails += 1
			time.Sleep(time.Second)
			continue
		}
		break
	}
	wsClient.Listen()
	for event := range wsClient.EventChannel {
		// Consider only posts for now.
		if event.EventType() != model.WebsocketEventPosted {
			continue
		}
		post := &model.Post{}
		err := json.Unmarshal([]byte(event.GetData()["post"].(string)), &post)
		if err != nil {
			app.logger.Error().Err(err).Msg("Could not cast event to *model.Post")
			continue
		}
		// Ignore messages sent by this bot itself.
		if post.UserId == app.user.Id {
			continue
		}

		chanType := event.GetData()["channel_type"]
		id := post.ChannelId
		if chanType == "D" {
			// Direct.
			err = dataBase.addUser(id, 0)
			if err != nil {
				app.logger.Error().Err(err).Msg("addUser error")
				continue
			}

			cmd, body, _ := strings.Cut(post.Message, " ")
			cmd = strings.ToUpper(cmd)
			body = strings.TrimSpace(body)
			switch cmd {
			case "ADD":
				app.handleAdd(id, body)
			case "REMOVE":
				app.handleRemove(id, body)
			case "VIEW":
				app.handleView(id, body)
			case "PAUSE":
				app.handlePause(id)
			case "CONTINUE":
				app.handleContinue(id)
			case "HELP":
				app.handleHelp(id)
			default:
				app.handleUnknown(id)
			}
		} else {
			app.handleUpdate(id, post.Message, post.Id)
		}
	}

	return 0
}

func MmMessageToText(msg Message) string {
	format := `Topic was detected: [%s]
	Summary: %s
	Link: %s
	`
	text := fmt.Sprintf(
		format, msg.Topic, msg.Summary, msg.Link)
	return text
}

func (a *application) handleUpdate(id, msg, msgId string) {
	if found, err := dataBase.containsChannel(id, MatterMost); !found || err != nil {
		if err != nil {
			a.logger.Error().Err(err).Msg("handleUpdate error")
		}
		return
	}
	possibleTopics, err := dataBase.getTopics(id, MatterMost)
	if err != nil {
		a.logger.Error().Err(err).Msg("handleUpdate error")
		return
	}
	if len(possibleTopics) == 0 {
		return
	}
	var foundTopics []string
	if foundTopics, err = api.analyze(msg, possibleTopics); err != nil || len(foundTopics) == 0 {
		if err != nil {
			a.logger.Error().Err(err).Msg("handleUpdate error")
		}
		return
	}

	var summary string
	if openAIkey != "" && len(msg) > summaryLength {
		if summary, err = api.summarize(msg, openAIkey); err != nil {
			err := fmt.Errorf("error in OpenAI uisng with error: %s \n", err.Error())
			a.logger.Error().Err(err).Msg("handleUpdate error")
			summary = summarize(msg)
		}
	} else {
		summary = summarize(msg)
	}

	sendUsers, err := dataBase.getUsers(id, foundTopics, MatterMost)
	for userId, userTopics := range sendUsers {
		isPaused, err := dataBase.isPaused(userId)
		if err != nil {
			a.logger.Error().Err(err)
			continue
		}
		for _, topic := range userTopics {
			if err := dataBase.setTime(userId, id, topic, MatterMost); err != nil {
				a.logger.Error().Err(err).Msg("handleUpdate error")
				continue
			}
		}

		finalTopics := strings.Join(userTopics, ", ")
		msg := Message{
			User: userId,
			Link: fmt.Sprintf("%s/%s/pl/%s",
				a.config.server, a.config.teamName, msgId),
			Topic:       finalTopics,
			Summary:     summary,
			Application: MatterMost,
		}
		if isPaused {
			if err := dataBase.addDelayedMessage(msg); err != nil {
				a.logger.Error().Err(err).Msg("handleUpdate error")
				continue
			}
		} else {
			if isPaused {
				if err := dataBase.addDelayedMessage(msg); err != nil {
					log.Printf(err.Error())
					continue
				}
			} else {
				a.sendMsg(msg.User, MmMessageToText(msg))
			}
		}
	}
}

func (a *application) handleUnknown(id string) {
	reply := "Я вас не понимаю, используйте HELP для справки"
	a.sendMsg(id, reply)
}

func (a *application) handleHelp(id string) {
	reply := "\n Мой набор команд включает в себя следующие опции: \n \n" +
		"VIEW - для просмотра доступных каналов и связанных с ними тем. \n \n" +
		"ADD <название канала> <слово>- добавляет указанное слово в список для поиска в конкретном канале. \n \n" +
		"REMOVE <название канала> <слово> - удаляет указанное слово из списка для поиска в конкретном канале.\n \n" +
		"PAUSE- приостанавливает обновления в боте. \n \n" +
		"CONTINUE - возобновляет поток обновлений в боте после приостановки. \n \n" +
		"Эти команды помогут вам управлять списком тем и слов для поиска, чтобы быстро находить нужную информацию в чатах."
	a.sendMsg(id, reply)
}

const errReply = "Упс, произошла ошибка..."

func (a *application) handleContinue(id string) {
	var err error
	var messages []Message
	if err = dataBase.unpauseUser(id); err != nil {
		a.logger.Error().Err(err).Msg("Failed to add topic")
		a.sendMsg(id, errReply)
		return
	}

	if messages, err = dataBase.getDelayedMessages(id); err != nil {
		a.logger.Error().Err(err).Msg("Failed to add topic")
		a.sendMsg(id, errReply)
		return
	}

	a.sendMsg(id, "Обновления сняты с паузы!")
	for _, msg := range messages {
		a.sendMsg(msg.User, MmMessageToText(msg))
	}
}

func (a *application) handlePause(id string) {
	if err := dataBase.pauseUser(id); err != nil {
		a.logger.Error().Err(err).Msg("Failed to add topic")
		a.sendMsg(id, errReply)
		return
	}
	a.sendMsg(id, "Обновления поставлены на паузу!")
}

func (a *application) handleView(id, body string) {
	totalInfo, err := dataBase.getUserInfo(id)
	if err != nil {
		a.logger.Error().Err(err).Msg("Failed to view topics")
		a.sendMsg(id, errReply)
		return
	}
	str := strings.Builder{}
	for _, topicByChan := range totalInfo {
		for ch, topics := range topicByChan {
			chName, err := dataBase.getMmChan(ch)
			if err != nil {
				a.logger.Error().Err(err).Msg("Failed to view topics")
				a.sendMsg(id, errReply)
				return
			}
			str.WriteString(fmt.Sprintf("%s:\n", chName))
			for _, topic := range topics {
				str.WriteString(fmt.Sprintf("   - %s\n", topic))
			}
			str.WriteString("\n")
			str.WriteString("\n")
		}
	}
	ans := str.String()
	if ans == "" {
		ans = "Ничего не отслеживается"
	}
	a.sendMsg(id, ans)
}

func (a *application) handleRemove(id, body string) {
	elements := strings.Fields(body)
	if len(elements) != 2 {
		a.sendMsg(id, "Неверное количество аргументов. Используйте REMOVE <канал> <топик>")
		return
	}
	channel := elements[0]
	// Determine channel ID.
	ch, _, err := a.client.GetChannelByName(channel, a.team.Id, "")
	if err != nil {
		a.logger.Error().Err(err).Msg("Failed to add topic")
		a.sendMsg(id, errReply)
		return
	}
	dataBase.addMmChan(ch.Id, channel)
	concern := Concern{
		Channel: ch.Id,
		Topic:   elements[1],
	}
	if err := dataBase.removeTopic(id, concern.Channel, concern.Topic, MatterMost); err != nil {
		a.logger.Error().Err(err).Msg("Failed to remove topic")
		a.sendMsg(id, errReply)
		return
	}
	a.sendMsg(id, "Топик удалён!")
}

func (a *application) handleAdd(id, body string) {
	elements := strings.Fields(body)
	if len(elements) != 2 {
		a.sendMsg(id, "Неверное количество аргументов. Используйте ADD <название канала> <топик>")
		return
	}
	channel := elements[0]
	// Determine channel ID.
	ch, _, err := a.client.GetChannelByName(channel, a.team.Id, "")
	if err != nil {
		a.logger.Error().Err(err).Msg("Failed to add topic")
		a.sendMsg(id, errReply)
		return
	}
	dataBase.addMmChan(ch.Id, channel)
	concern := Concern{
		Channel: ch.Id,
		Topic:   elements[1],
	}
	if err := dataBase.addTopic(id, concern.Channel, concern.Topic, MatterMost); err != nil {
		a.logger.Error().Err(err).Msg("Failed to add topic")
		a.sendMsg(id, errReply)
		return
	}
	a.sendMsg(id, "Топик добавлен!")
}

func (a *application) sendMsg(id, msg string) {
	post := &model.Post{}
	post.ChannelId = id
	post.Message = msg
	if _, _, err := a.client.CreatePost(post); err != nil {
		a.logger.Error().Err(err).Msg("Failed to create post")
	}
}
