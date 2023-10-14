package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
	"io/ioutil"
	"net/http"
)

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

var (
	topicTimes    ChannelTopicTime
	topicCount    ChannelTopicCount
	dataBase      UsersChannelsTopics
	analyzer      BasicTextAnalyzer
	summarizer    Summarizer
	summaryLength = 100
)

func main() {

	topicCount.topicTime = &topicTimes
	dataBase.topicCount = &topicCount
	summarizer = MessageSummarizer{textLength: summaryLength}
	analyzer = Analyzer{}

	router := gin.Default()

	router.POST("/add", add)
	router.POST("/remove", remove)
	router.GET("/news", news)

	router.OPTIONS("/add", auto200)
	router.OPTIONS("/remove", auto200)
	router.OPTIONS("/news", auto200)

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
	dataBase.lock()
	defer dataBase.unLock()
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

	err = dataBase.add(message.User, message.Channel, message.Topic)
	if err != nil {
		return
	}

}

func remove(c *gin.Context) {
	dataBase.lock()
	defer dataBase.unLock()
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

	err = dataBase.remove(message.User, message.Channel, message.Topic)
	if err != nil {
		setAnswer(c, http.StatusBadRequest, err.Error())
	}

}

func news(c *gin.Context) {

	dataBase.lock()
	defer dataBase.unLock()

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "body reading error")
		return
	}

	expected := NewsMessage{}

	err = json.Unmarshal(body, &expected)
	if err != nil {
		setAnswer(c, http.StatusInternalServerError, "json parsing error")
		return
	}

}
