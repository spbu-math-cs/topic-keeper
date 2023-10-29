package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

type AnalyzerRequest struct {
	Text   string   `json:"text"`
	Topics []string `json:"topics"`
}

type AnalyzerReturn struct {
	Summary string   `json:"summary"`
	Topics  []string `json:"topics"`
}

type OpenAIAnswer struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

type API interface {
	addTopic(username string, c Concern) error
	removeTopic(username string, c Concern) error
	viewTopics(username string) ([]Concern, error)
	postMessage(chanName string, msg string) ([]ReturnMessage, error)
	analyze(msg string, topics []string) ([]string, error)
	summarize(text string) (string, error)
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

func (b basicAPI) analyze(msg string, topics []string) ([]string, error) {

	body := AnalyzerRequest{Topics: topics, Text: msg}
	bodyAsBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s/analyze", apiAddr),
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
	var res AnalyzerReturn
	respBody, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(respBody, &res); err != nil {
		return nil, err
	}
	return res.Topics, nil
}

func (b basicAPI) summarize(text string) (string, error) {
	///TODO() вставить ключ перед демо
	apiKey := ""
	url := "https://api.openai.com/v1/chat/completions"

	requestBody := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]string{
			{"role": "user", "content": "Сократи объем сообщения до 1 предложения не меняя его суть" +
				"Вот текст : \n" + `"` + text + `"`},
		},
		"temperature": 0.5,
	}

	requestBodyJSON, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBodyJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return text, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var response OpenAIAnswer
	err = json.Unmarshal(body, &response)
	if err != nil {
		return text, err
	}

	choices := response.Choices
	if len(choices) == 0 {
		return text, nil
	}
	summary := choices[0].Message.Content

	return summary, nil
}

var _ API = (*basicAPI)(nil)
