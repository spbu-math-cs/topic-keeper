package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v4/stdlib"
	"io/ioutil"
	"net/http"
	"slices"
)

type Message struct {
	User    string `json:"user"`
	Channel string `json:"channel"`
	Topic   string `json:"topic"`
}

// словарь типа [пользователь] - [словарь типа [канал] - [топики канала]]
var users map[string]map[string][]string

// словарь типа [канал] - [словарь типа [топик] - [кол-во ссылок]]
var channelTopicCount map[string]map[string]int

func main() {
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
	fmt.Println(c.Request.Header)

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		setError(c, http.StatusInternalServerError, "body reading error")
		return
	}

	var message Message

	err = json.Unmarshal(body, &message)

	if err != nil {
		setError(c, http.StatusInternalServerError, "json parsing error")
		return
	}

	if !slices.Contains(users[message.User][message.Channel], message.Topic) {
		users[message.User][message.Channel] = append(users[message.User][message.Channel], message.Topic)
		channelTopicCount[message.Channel][message.Topic] += 1
	}
}

func remove(c *gin.Context) {
	fmt.Println(c.Request.Header)

	body, err := ioutil.ReadAll(c.Request.Body)

	if err != nil {
		setError(c, http.StatusInternalServerError, "body reading error")
		return
	}

	var message Message

	err = json.Unmarshal(body, &message)

	if err != nil {
		setError(c, http.StatusInternalServerError, "json parsing error")
		return
	}

	_, ok := users[message.User]
	if !ok {
		setError(c, http.StatusBadRequest, "can`t find user")
		delete(users, message.User)
		return
	}

	_, ok = users[message.User][message.Channel]
	if !ok {
		setError(c, http.StatusBadRequest, "can`t find channel in this user`s channels")
		delete(users[message.User], message.Channel)
		return
	}

	for i := 0; i < len(users[message.User][message.Channel]); i += 1 {
		if users[message.User][message.Channel][i] == message.Topic {
			users[message.User][message.Channel] = append(users[message.User][message.Channel][:i], users[message.User][message.Channel][i+1:]...)
			channelTopicCount[message.Channel][message.Topic] -= 1
			break
		}
	}

	if len(users[message.User][message.Channel]) == 0 {
		delete(users[message.User], message.Channel)
	}

	if len(users[message.User]) == 0 {
		delete(users, message.User)
	}

	if channelTopicCount[message.Channel][message.Topic] <= 0 {
		delete(channelTopicCount[message.Channel], message.Topic)
	}
}
