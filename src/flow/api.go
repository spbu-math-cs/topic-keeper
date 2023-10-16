package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var apiAddr = "localhost:8080"

type Concern struct {
	Channel string `json:"channel"`
	Topic   string `json:"topic"`
}

type ReturnMessage struct {
	User    string `json:"user"`
	Channel string `json:"channel"`
	Topic   string `json:"topic"`
	Summary string `json:"summary"`
}

type API interface {
	addTopic(username string, c Concern) error
	removeTopic(username string, c Concern) error
	viewTopics(username string) ([]Concern, error)
	postMessage(chanName string, msg string) ([]ReturnMessage, error)
}

type basicAPI struct{}

func postQuery(uri string, body any) ([]byte, error) {
	bodyAsBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(uri, "application/json",
		bytes.NewReader(bodyAsBytes))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

func (b basicAPI) addTopic(username string, c Concern) error {
	body := map[string]any{
		"user":    username,
		"channel": c.Channel,
		"topic":   c.Topic,
	}
	_, err := postQuery(fmt.Sprintf("http://%s/add", apiAddr), body)
	return err
}

func (b basicAPI) removeTopic(username string, c Concern) error {
	body := map[string]any{
		"user":    username,
		"channel": c.Channel,
		"topic":   c.Topic,
	}
	_, err := postQuery(fmt.Sprintf("http://%s/remove", apiAddr), body)
	return err
}

func (b basicAPI) viewTopics(username string) ([]Concern, error) {
	body := map[string]any{
		"user": username,
	}
	bodyAsBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/view", apiAddr),
		bytes.NewReader(bodyAsBytes))
	if err != nil {
		return nil, err
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}
	var res []Concern
	respBody, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(respBody, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (b basicAPI) postMessage(chanName string, msg string) ([]ReturnMessage, error) {
	body := map[string]any{
		"channel": chanName,
		"text":    msg,
	}
	resp, err := postQuery(fmt.Sprintf("http://%s/news", apiAddr), body)
	if err != nil {
		return nil, err
	}
	var repl []ReturnMessage
	if err := json.Unmarshal(resp, &repl); err != nil {
		return nil, err
	}
	return repl, nil
}

var _ API = (*basicAPI)(nil)
