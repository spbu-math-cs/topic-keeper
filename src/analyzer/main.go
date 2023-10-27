package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type AnalyzerRequest struct {
	Text   string   `json:"text"`
	Topics []string `json:"topics"`
}

type AnalyzerReturn struct {
	Summary string   `json:"summary"`
	Topics  []string `json:"topics"`
}

type CommandMessage struct {
	User    string `json:"user"`
	Channel string `json:"channel"`
	Topic   string `json:"topic"`
}

type NewsMessage struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type ReturnMessage struct {
	User    string `json:"user"`
	Channel string `json:"channel"`
	Topic   string `json:"topic"`
	Summary string `json:"summary"`
}

type User struct {
	User string `json:"user"`
}

type ChannelTopic struct {
	Channel string `json:"channel"`
	Topic   string `json:"topic"`
}

var (
	topicTimes SafeStorage
	topicCount SafeStorage
	dataBase   SafeStorage
	analyzer   BasicTextAnalyzer
	summarizer Summarizer
)

const (
	summaryLength   = 100
	minimalInterval = time.Second
)

func main() {

	analyzer = &Analyzer{}
	summarizer = &MessageSummarizer{textLength: summaryLength}
	topicTimes = &ChannelTopicTime{interval: minimalInterval}
	topicTimes.create()
	topicCount = &ChannelTopicCount{}
	topicCount.create()
	dataBase = &UsersChannelsTopics{}
	dataBase.create()

	router := gin.Default()

	router.POST("/add", add)
	router.POST("/remove", remove)
	router.POST("/news", news)
	router.POST("/view", view)
	router.POST("/analyze", analyze)

	router.OPTIONS("/add", auto200)
	router.OPTIONS("/remove", auto200)
	router.OPTIONS("/news", auto200)
	router.OPTIONS("/view", auto200)
	router.OPTIONS("/analyze", auto200)

	err := router.Run("0.0.0.0:8080")
	if err != nil {
		return
	}
}

func auto200(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
	c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	c.Header("Access-Control-Allow-Credentials", "true")
}

func setAnswer(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"message": message,
	})
}

func add(c *gin.Context) {
	fmt.Println(c.Request.Header)

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "body reading error")
		return
	}

	var message CommandMessage

	err = json.Unmarshal(body, &message)

	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "json parsing error")
		return
	}

	err = dataBase.add(message.User, message.Channel, message.Topic, Topic)
	if err != nil {
		return
	}

	err = topicCount.add(message.Channel, message.Topic, "", Topic)
	if err != nil {
		return
	}

	err = topicTimes.add(message.Channel, message.Topic, "", Topic)
	if err != nil {
		return
	}

}

func remove(c *gin.Context) {
	fmt.Println(c.Request.Header)

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "body reading error")
		return
	}

	var message CommandMessage

	err = json.Unmarshal(body, &message)

	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "json parsing error")
		return
	}

	err = dataBase.remove(message.User, message.Channel, message.Topic, Topic)
	if err != nil {
		setAnswer(c, http.StatusBadRequest, err.Error())
	}

	err = topicCount.remove(message.Channel, message.Topic, "", Topic)

	_, err = topicCount.get(message.Channel, message.Topic, Count)

	if errors.Is(err, invalidTopicError) || errors.Is(err, invalidChannelError) {
		err = topicTimes.remove(message.Channel, message.Topic, "", Topic)
		if err != nil {
			return
		}
	}
}

func news(c *gin.Context) {

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "body reading error")
		return
	}

	message := NewsMessage{}

	err = json.Unmarshal(body, &message)
	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "json parsing error")
		return
	}

	summary, err := summarizer.summarize(message.Text)
	if err != nil {
		setAnswer(c, http.StatusInternalServerError, err.Error())
		return
	}

	possibleTopicsAny, err := topicTimes.get(message.Channel, "", Topics)
	if err != nil {
		setAnswer(c, http.StatusInternalServerError, err.Error())
		return
	}

	possibleTopics, ok := (possibleTopicsAny).([]string)
	if !ok {
		setAnswer(c, http.StatusInternalServerError, "topics conversion error")
		return
	}

	topicsInMessage, err := analyzer.analyze(possibleTopics, message.Text)

	usersTopicInChannelAny, err := dataBase.get(message.Channel, "", UsersTopicsByChannel)
	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "data base error")
		return
	}

	usersTopicInChannel, ok := (usersTopicInChannelAny).(map[string]map[string]struct{})
	if !ok {
		setAnswer(c, http.StatusInternalServerError, "conversion error")
		return
	}

	sendMessageMap := map[string]ReturnMessage{}

	for user, userTopicsInChannel := range usersTopicInChannel {
		for topic, _ := range userTopicsInChannel {
			if contains(topicsInMessage, topic) {
				_, ok = sendMessageMap[user]
				if !ok {
					sendMessageMap[user] = ReturnMessage{
						Summary: summary,
						User:    user,
						Topic:   topic,
						Channel: message.Channel,
					}
				} else {
					wasTopics := sendMessageMap[user].Topic
					sendMessageMap[user] = ReturnMessage{
						Summary: summary,
						User:    user,
						Topic:   wasTopics + ", " + topic,
						Channel: message.Channel,
					}
				}

			}
		}
	}

	var bd []ReturnMessage
	for _, returnMessage := range sendMessageMap {
		bd = append(bd, returnMessage)
	}

	c.JSON(http.StatusOK, bd)
}

func view(c *gin.Context) {
	fmt.Println(c.Request.Header)

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "body reading error")
		return
	}

	var user User

	err = json.Unmarshal(body, &user)

	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "json parsing error")
		return
	}

	var db []ChannelTopic

	dataAny, err := dataBase.get(user.User, "", Channels)
	if err != nil {
		if errors.Is(err, invalidUserError) {
			c.JSON(http.StatusOK, db)
			return
		}
		setAnswer(c, http.StatusInternalServerError, "data bse error")
		return
	}

	data, ok := (dataAny).(map[string]map[string]struct{})
	if !ok {
		setAnswer(c, http.StatusInternalServerError, "conversion error")
		return
	}

	for channel, topics := range data {
		for topic, _ := range topics {
			db = append(db, ChannelTopic{Channel: channel, Topic: topic})
		}
	}

	c.JSON(http.StatusOK, db)
}

func analyze(c *gin.Context) {
	fmt.Println(c.Request.Header)

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "body reading error")
		return
	}

	var request AnalyzerRequest

	err = json.Unmarshal(body, &request)

	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "json parsing error")
		return
	}

	topics, err := analyzer.analyze(request.Topics, request.Text)
	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "analyzer error")
		return
	}

	summary, err := summarizer.summarize(request.Text)
	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "summarizer error")
		return
	}

	c.JSON(http.StatusOK, AnalyzerReturn{Topics: topics, Summary: summary})

}
