package main

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strconv"
	"strings"
)

const (
	privateLinkTelegram = "https://t.me/c/%s/%s"
	publicLinkTelegram  = "https://t.me/%s/%s"
)

type TelegramHandler struct {
	keyBoard tgbotapi.ReplyKeyboardMarkup
	updates  tgbotapi.UpdatesChannel
}

func newTelegramListener(b *tgbotapi.BotAPI) *TelegramHandler {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.GetUpdatesChan(u)
	keyBoard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/start"),
			tgbotapi.NewKeyboardButton("/view"),
			tgbotapi.NewKeyboardButton("/pause"),
			tgbotapi.NewKeyboardButton("/continue"),
			tgbotapi.NewKeyboardButton("/help"),
		),
	)
	return &TelegramHandler{keyBoard, updates}
}

func (t *TelegramHandler) handleUpdates() {
	for update := range t.updates {
		switch {
		case update.ChannelPost != nil:
			if update.ChannelPost.Chat.UserName != "" {
				hsh := getHash(update.ChannelPost.Chat.UserName)
				w := workEvent{
					application: Telegram,
					channel:     update.ChannelPost.Chat.UserName,
					channelID:   strconv.FormatInt(update.ChannelPost.Chat.ID, 10),
					text:        update.ChannelPost.Text,
					messageID:   strconv.Itoa(update.ChannelPost.MessageID),
				}
				w.link = createPublicLink(w)
				workChans[hsh%NWorkers] <- w
			} else {
				hsh := getHash(update.ChannelPost.Chat.Title)
				w := workEvent{
					application: Telegram,
					channel:     update.ChannelPost.Chat.Title,
					channelID:   getPrivateID(update.ChannelPost.Chat.ID),
					text:        update.ChannelPost.Text,
					messageID:   strconv.Itoa(update.ChannelPost.MessageID),
				}
				w.link = createPrivateLink(w)
				workChans[hsh%NWorkers] <- w
			}
		case update.Message != nil:
			if update.Message.Chat.IsSuperGroup() {
				if update.Message.Chat.UserName != "" {
					hsh := getHash(update.Message.Chat.UserName)
					w := workEvent{
						application: Telegram,
						channel:     update.Message.Chat.UserName,
						channelID:   strconv.FormatInt(update.Message.Chat.ID, 10),
						text:        update.Message.Text,
						messageID:   update.Message.Text,
					}
					w.link = createPublicLink(w)
					workChans[hsh%NWorkers] <- w
				} else {
					hsh := getHash(update.Message.Chat.Title)
					w := workEvent{
						application: Telegram,
						channel:     update.Message.Chat.Title,
						channelID:   getPrivateID(update.Message.Chat.ID),
						text:        update.Message.Text,
						messageID:   update.Message.Text,
					}
					w.link = createPrivateLink(w)
					workChans[hsh%NWorkers] <- w
				}
				return
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			msg.ReplyMarkup = t.keyBoard
			_, err := bot.Send(msg)
			if err != nil {
				log.Println(err.Error())
			}

			updText := strings.Trim(update.Message.Text, "\n ")
			uname := update.Message.Chat.UserName
			err = dataBase.addUser(uname, update.Message.Chat.ID)
			if err != nil {
				log.Println(err.Error())
				return
			}

			switch updText {
			case "/start":
				handleStart(uname)
			case "/view":
				handleView(uname)
			case "/help":
				handleHelp(uname)
			case "/pause":
				handlePause(uname)
			case "/continue":
				handleContinue(uname)
			default:
				if strings.HasPrefix(updText, "/add") {
					handleAdd(uname, updText)
				} else if strings.HasPrefix(updText, "/removeChannel") {
					handleRemoveChannel(uname, updText)
				} else if strings.HasPrefix(updText, "/remove") {
					handleRemove(uname, updText)
				} else {
					handleUnknownCommand(uname)
				}
			}
		}
	}
}

func getPrivateID(id int64) string {
	ID := strconv.FormatInt(id, 10)
	s, _ := strings.CutPrefix(ID, "-100")
	return s
}

func createPrivateLink(w workEvent) string {
	return fmt.Sprintf(privateLinkTelegram, w.channelID, w.messageID)
}

func createPublicLink(w workEvent) string {
	return fmt.Sprintf(publicLinkTelegram, w.channel, w.messageID)
}
