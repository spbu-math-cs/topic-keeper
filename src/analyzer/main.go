package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
	"io/ioutil"
	"net/http"
)

type commandMessage struct {
	User    string `json:"user"`
	Channel string `json:"channel"`
	Topic   string `json:"topic"`
}

type newsMessage struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
}

type returnMessage struct {
	User    string `json:"user"`
	Channel string `json:"channel"`
	Topic   string `json:"topic"`
	Summary string `json:"summary"`
}

var (
	topicTimes channelTopicTime
	topicCount channelTopicCount
	dataBase   usersChannelsTopics
)

func main() {

	dataBase.topicCount = &topicCount
	topicCount.topicTime = &topicTimes

	router := gin.Default()

	router.POST("/add", add)
	router.POST("/remove", remove)

	router.OPTIONS("/add", auto200)
	router.OPTIONS("/remove", auto200)

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

func setError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"error information": message,
	})
}

func add(c *gin.Context) {
	dataBase.mut.Lock()
	defer dataBase.mut.Unlock()
	fmt.Println(c.Request.Header)

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		setError(c, http.StatusInternalServerError, "body reading error")
		return
	}

	var message commandMessage

	err = json.Unmarshal(body, &message)

	if err != nil {
		setError(c, http.StatusInternalServerError, "json parsing error")
		return
	}

	dataBase.add(message.User, message.Channel, message.Topic)

}

func remove(c *gin.Context) {
	dataBase.mut.Lock()
	defer dataBase.mut.Unlock()
	fmt.Println(c.Request.Header)

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		setError(c, http.StatusInternalServerError, "body reading error")
		return
	}

	var message commandMessage

	err = json.Unmarshal(body, &message)

	if err != nil {
		setError(c, http.StatusInternalServerError, "json parsing error")
		return
	}

	_, ok := dataBase.channelTopics[message.User]
	if !ok {
		setError(c, http.StatusBadRequest, "can`t find user")
		delete(dataBase.channelTopics, message.User)
		return
	}

	_, ok = dataBase.channelTopics[message.User][message.Channel]
	if !ok {
		setError(c, http.StatusBadRequest, "can`t find channel in this user`s channels")
		delete(dataBase.channelTopics[message.User], message.Channel)
		return
	}

	dataBase.remove(message.User, message.Channel, message.Topic)

}
