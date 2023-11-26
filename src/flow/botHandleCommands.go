package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleStart(username string) {
	userId, err := dataBase.getID(username)
	if err != nil {
		sendMessage(username, err.Error())
		log.Printf(err.Error())
		return
	}
	reply := "Привет! Бот предназначен для помощи в быстром и эффективном поиске нужной информации в чатах и темах на основе предоставленного списка."
	msg := tgbotapi.NewMessage(userId, reply)
	msg.ReplyMarkup = createMenuKeyboard()
	handleHelp(username)
}

func handleView(username string) {
	totalInfo, err := dataBase.getUserInfo(username)
	if err != nil {
		log.Println(err.Error())
		sendMessage(username, err.Error())
		return
	}
	str := strings.Builder{}
	for application, topicByChan := range totalInfo {
		str.WriteString(fmt.Sprintf("%s:\n", application))
		for ch, topics := range topicByChan {
			str.WriteString(fmt.Sprintf(" %s:\n", ch))
			for _, topic := range topics {
				str.WriteString(fmt.Sprintf("   - %s\n", topic))
			}
			str.WriteString("\n")
		}
	}
	ans := str.String()
	if ans == "" {
		ans = "Ничего не отслеживается"
	}
	sendMessage(username, ans)
}

func handleAdd(username string, msg string) {
	after, _ := strings.CutPrefix(msg, "/add")
	elements := strings.Fields(after)
	if len(elements) != 3 {
		sendMessage(username, "Неверное количество аргументов. Используйте /add <название канала> <ссылка/топик> <платформа>")
		return
	}
	platform := elements[2]

	if platform != "VK" && platform != "TG" {
		sendMessage(username, "Неподдерживаемая платформа. Используйте 'VK' или 'TG'.")
		return
	}
	if platform == "VK" {
		after = strings.TrimSuffix(after, "VK")
		handleAddVK(username, after)
	} else {
		after = strings.TrimSuffix(after, "TG")
		handleAddTelegram(username, after)
	}
}

func handleAddTelegram(username string, msg string) {
	concern, err := parseTopic(msg)
	fmt.Println(concern)
	if err != nil {
		sendMessage(username, err.Error())
		return
	}
	if err := dataBase.addTopic(username, concern.Channel, concern.Topic, Telegram); err != nil {
		sendMessage(username, err.Error())
		return
	}
	sendMessage(username, "Топик добавлен!")
}
func handleAddVK(username, text string) {
	after := strings.TrimSpace(text)
	link, topic, ok := strings.Cut(after, " ")

	if !ok {
		sendMessage(username, wrongFmtError.Error())
		return
	}

	objectName, objectType, id, err := getVKInfo(link, vkToken)
	if err != nil {
		log.Println(err.Error())
		return
	}

	if objectType != "group" {
		log.Println("Resolved object is not a group")
		return
	}

	groupID := fmt.Sprintf("%d", id)

	topic = strings.TrimSpace(topic)

	if err := dataBase.addTopic(username, groupID, topic, VK); err != nil {
		log.Println(err.Error())
		return
	}

	if err := dataBase.addVKPublic(objectName, groupID, 0); err != nil {
		log.Println(err.Error())
		return
	}

	sendMessage(username, "Топик добавлен")
}
func handleRemoveTopic(username, msg string) {
	after, _ := strings.CutPrefix(msg, "/remove")
	elements := strings.Fields(after)

	if len(elements) != 3 {
		sendMessage(username, "Неверное количество аргументов. Используйте /remove <название канала> <ссылка/топик> <платформа>")
		return
	}
	platform := elements[2]

	if platform != "VK" && platform != "TG" {
		sendMessage(username, "Неподдерживаемая платформа. Используйте 'VK' или 'TG'.")
		return
	}

	if platform == "VK" {
		after = strings.TrimSuffix(after, "VK")
		handleRemoveTopicVK(username, after)
	} else {
		after = strings.TrimSuffix(after, "TG")
		handleRemoveTopicTelegram(username, after)
	}
}
func handleRemoveTopicTelegram(username, msg string) {
	concern, err := parseTopic(msg)
	fmt.Println(concern)
	if err != nil {
		sendMessage(username, err.Error())
		return
	}
	if err := dataBase.removeTopic(username, concern.Channel, concern.Topic, Telegram); err != nil {
		sendMessage(username, err.Error())
		return
	}
	sendMessage(username, "Топик удалён!")
}
func handleRemoveTopicVK(username, text string) {
	after := strings.TrimSpace(text)
	link, topic, ok := strings.Cut(after, " ")

	if !ok {
		sendMessage(username, wrongFmtError.Error())
	}

	_, objectType, id, err := getVKInfo(link, vkToken)
	if err != nil {
		log.Println(err.Error())
		return
	}

	if objectType != "group" {
		log.Println("Resolved object is not a group")
		return
	}

	groupID := fmt.Sprintf("%d", id)

	topic = strings.TrimSpace(topic)

	if err := dataBase.removeTopic(username, groupID, topic, VK); err != nil {
		log.Println(err.Error())
		return
	}

	sendMessage(username, "Топик удален")
}
func handleRemoveChannel(username, msg string) {
	after, _ := strings.CutPrefix(msg, "/removeChannel")
	elements := strings.Fields(after)

	if len(elements) != 2 {
		sendMessage(username, "Неверное количество аргументов. Используйте /removeChannel <название канала> <платформа>")
		return
	}

	channel := elements[0]
	platform := elements[1]

	if platform != "VK" && platform != "TG" {
		sendMessage(username, "Неподдерживаемая платформа. Используйте 'VK' или 'TG'.")
		return
	}

	if platform == "VK" {
		handleRemoveChannelVK(username, channel)
	} else {
		handleRemoveChannelTelegram(username, channel)
	}
}
func handleRemoveChannelTelegram(username, channel string) {
	var err error
	if channel, err = parseChannelName(channel); err != nil {
		sendMessage(username, err.Error())
		return
	}
	if err := dataBase.removeChannel(username, channel, Telegram); err != nil {
		sendMessage(username, err.Error())
		return
	}
	sendMessage(username, "Канал удалён!")
}

func handleRemoveChannelVK(username, text string) {
	after := strings.TrimSpace(text)
	link, _, ok := strings.Cut(after, " ")

	if !ok {
		sendMessage(username, wrongFmtError.Error())
	}

	_, objectType, id, err := getVKInfo(link, vkToken)
	if err != nil {
		log.Println(err.Error())
		return
	}

	if objectType != "group" {
		log.Println("Resolved object is not a group")
		return
	}

	groupID := fmt.Sprintf("%d", id)

	if err := dataBase.removeChannel(username, groupID, VK); err != nil {
		log.Println(err.Error())
		return
	}

	sendMessage(username, "Канал удален")
}

func handleUnknownCommand(username string) {
	reply := "Я не понимаю вашей команды. Воспользуйтесь \n /start \n /view \n /add <name>/<link> <topic> <platform> \n /remove <name>/<link> <topic> <platform> \n " +
		"/pause \n /continue \n /removeChannel <name>/<link> <platform> \n /help"
	sendMessage(username, reply)
}

func handlePause(username string) {
	if err := dataBase.pauseUser(username); err != nil {
		sendMessage(username, err.Error())
		return
	}
	sendMessage(username, "Обновления поставлены на паузу!")
}

func handleContinue(username string) {
	var err error
	var messages []Message
	if err = dataBase.unpauseUser(username); err != nil {
		sendMessage(username, err.Error())
		return
	}

	if messages, err = dataBase.getDelayedMessages(username); err != nil {
		sendMessage(username, err.Error())
		return
	}

	for _, msg := range messages {
		sendNews(msg)
	}

	sendMessage(username, "Обновления сняты с паузы!")
}

func handleHelp(username string) {
	reply := "\n Мой набор команд включает в себя следующие опции: \n \n" +
		"/view - для просмотра доступных каналов и связанных с ними тем. \n \n" +
		"/add <@название канала>/<ссылка на канал> <слово> <платформа> - добавляет указанное слово в список для поиска в конкретном канале. \n \n" +
		"/remove <@название канала>/<ссылка на канал> <слово> <платформа> - удаляет указанное слово из списка для поиска в конкретном канале.\n \n" +
		"/pause - приостанавливает обновления в боте. \n \n" +
		"/continue - возобновляет поток обновлений в боте после приостановки. \n \n" +
		"/removeChannel <@название канала>/<ссылка на канал> <платформа> - удаляет список для поиска в конкретном канале. \n" +
		"В качестве платформы нужно указывать либо VK, либо TG.\n \n" +
		"Эти команды помогут вам управлять списком тем и слов для поиска, чтобы быстро находить нужную информацию в чатах."
	sendMessage(username, reply)
}

func setBotCommands(bot *tgbotapi.BotAPI) {
	commands := []tgbotapi.BotCommand{
		{
			Command:     "start",
			Description: "Начать работу с ботом",
		},
		{
			Command:     "view",
			Description: "Просмотреть доступные каналы и темы",
		},
		{
			Command:     "add",
			Description: "Добавить слово в список для поиска",
		},
		{
			Command:     "remove",
			Description: "Удалить слово из списка для поиска",
		},
		{
			Command:     "pause",
			Description: "Приостановка получения обновлений",
		},
		{
			Command:     "continue",
			Description: "Возобновление получений обновлений",
		},
		{
			Command:     "removeChannel",
			Description: "Удалить канал с его историей поиска",
		},
		{
			Command:     "help",
			Description: "Получить помощь",
		},
	}
	config := tgbotapi.NewSetMyCommands(commands...)
	_, err := bot.Request(config)
	if err != nil {
		log.Printf(err.Error())
	}
}

func createMenuKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/start"),
			tgbotapi.NewKeyboardButton("/view"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/pause"),
			tgbotapi.NewKeyboardButton("/continue"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/help"),
		),
	)
	return keyboard
}

func HandlegetHistoryVK(username, msg string) {
	after, _ := strings.CutPrefix(msg, "/historyVK")
	txt := strings.Fields(after)
	if len(txt) != 2 {
		sendMessage(username, wrongFmtError.Error())
		return
	}
	link := txt[0]
	count, err := strconv.Atoi(txt[1])
	if err != nil {
		sendMessage(username, "Incorrect count of posts")
		return
	}

	publicName, objectType, objectID, err := getVKInfo(link, vkToken)
	if err != nil {
		sendMessage(username, err.Error())
		return
	}
	if objectType != "group" {
		sendMessage(username, "Resolved object is not a group")
		return
	}

	ID := fmt.Sprintf("%d", objectID)

	hsh := getHash(publicName) % VKNHistoryWorkers
	VKHistoryChans[hsh] <- UserHistory{
		user:       username,
		link:       link,
		publicID:   ID,
		postsCount: count,
		publicName: publicName,
	}
}
