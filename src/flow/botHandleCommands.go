package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
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
	ans := str.String()
	if ans == "" {
		ans = "Ничего не отслеживается"
	}
	sendMessage(username, ans)
}

func handleAdd(username string, msg string) {
	after, _ := strings.CutPrefix(msg, "/add")
	concern, err := parseTopic(after)
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

func handleRemoveTopic(username, msg string) {
	after, _ := strings.CutPrefix(msg, "/remove")
	concern, err := parseTopic(after)
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

func handleRemoveChannel(username, msg string) {
	channel, _ := strings.CutPrefix(msg, "/removeChannel")
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

func handleUnknownCommand(username string) {
	reply := "Я не понимаю вашей команды. Воспользуйтесь \n /start \n /view \n /add <name>/<link> <topic> \n /remove <name>/<link> <topic> \n " +
		"/pause \n /continue \n /removeChannel <name>/<link> \n /help"
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
		"/add <@название канала>/<ссылка на канал> <слово> - добавляет указанное слово в список для поиска в конкретном канале. \n \n" +
		"/remove <@название канала>/<ссылка на канал> <слово> - удаляет указанное слово из списка для поиска в конкретном канале.\n \n" +
		"/pause - приостанавливает обновления в боте. \n \n" +
		"/continue - возобновляет поток обновлений в боте после приостановки. \n \n" +
		"/removeChannel <@название канала>/<ссылка на канал> - удаляет список для поиска в конкретном канале. \n \n" +
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

func handleAddVK(username, text string) {
	after, _ := strings.CutPrefix(text, "/addVK")
	after = strings.TrimSpace(after)
	link, topic, ok := strings.Cut(after, " ")

	if !ok {
		sendMessage(username, wrongFmtError.Error())
	}

	objectType, id, err := getVKInfo(link, vkToken)
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

	sendMessage(username, "Топик добавлен")
}

func handleRemoveTopicVK(username, text string) {
	after, _ := strings.CutPrefix(text, "/removeVK")
	after = strings.TrimSpace(after)
	link, topic, ok := strings.Cut(after, " ")

	if !ok {
		sendMessage(username, wrongFmtError.Error())
	}

	objectType, id, err := getVKInfo(link, vkToken)
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

func handleRemoveChannelVK(username, text string) {
	after, _ := strings.CutPrefix(text, "/removeChannelVK")
	after = strings.TrimSpace(after)
	link, _, ok := strings.Cut(after, " ")

	if !ok {
		sendMessage(username, wrongFmtError.Error())
	}

	objectType, id, err := getVKInfo(link, vkToken)
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
