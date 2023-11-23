package main

type VKListener struct {
	accessToken string
}

func (*VKListener) handleUpdates() {

}

func newVKListener() VKListener {
	return VKListener{}
}

// import (
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"net/http"
// 	"os"
// 	"os/signal"
// 	"strings"
// 	"sync"
// 	"time"
// )

// type Post struct {
// 	ID          int       `json:"id"`
// 	Text        string    `json:"text"`
// 	FetchedTime time.Time `json:"fetched_time"`
// 	URL         string    `json:"url"`
// }

// var (
// 	posts []Post
// 	mutex sync.Mutex
// )

// type VKAPIResponse struct {
// 	Response struct {
// 		Items []Post `json:"items"`
// 	} `json:"response"`
// }

// func resolveScreenName(accessToken, screenName string) (*ResolvedInfo, error) {
// 	apiVersion := "5.154"
// 	requestURL := fmt.Sprintf("https://api.vk.com/method/utils.resolveScreenName?screen_name=%s&access_token=%s&v=%s",
// 		screenName, accessToken, apiVersion)

// 	resp, err := http.Get(requestURL)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	var result struct {
// 		Response ResolvedInfo `json:"response"`
// 	}

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = json.Unmarshal(body, &result)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &result.Response, nil
// }

// type ResolvedInfo struct {
// 	ObjectType string `json:"type"`
// 	ObjectID   int    `json:"object_id"`
// }

// func extractGroupNameFromURL(url string) string {
// 	url = strings.TrimPrefix(url, "http://")
// 	url = strings.TrimPrefix(url, "https://")
// 	url = strings.TrimPrefix(url, "www.")

// 	parts := strings.Split(url, "/")
// 	if len(parts) > 1 && parts[0] == "vk.com" {
// 		return parts[1]
// 	}

// 	return ""
// }

// func fetchPosts(groupID string) {
// 	accessToken := "b397ce84b397ce84b397ce8432b0819482bb397b397ce84d6cdd2ca964821d7fd266b76"
// 	apiVersion := "5.154"
// 	count := "10"

// 	url := fmt.Sprintf("https://api.vk.com/method/wall.get?owner_id=-%s&count=%s&access_token=%s&v=%s",
// 		groupID, count, accessToken, apiVersion)

// 	resp, err := http.Get(url)
// 	if err != nil {
// 		log.Printf("Error fetching posts: %s", err)
// 		return
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Printf("Error reading response body: %s", err)
// 		return
// 	}

// 	var apiResp VKAPIResponse
// 	err = json.Unmarshal(body, &apiResp)
// 	if err != nil {
// 		log.Printf("Error decoding JSON: %s", err)
// 		return
// 	}

// 	mutex.Lock()
// 	defer mutex.Unlock()
// 	for _, post := range apiResp.Response.Items {
// 		if !postExists(post.ID) {
// 			post.FetchedTime = time.Now()
// 			post.URL = fmt.Sprintf("https://vk.com/wall-%s_%d", groupID)
// 			posts = append(posts, post)
// 			printBlue(fmt.Sprintf("%d\n", post.ID))
// 			log.Printf("New post fetched: %+v\n", post)
// 		}
// 	}
// }

// func postExists(id int) bool {
// 	for _, p := range posts {
// 		if p.ID == id {
// 			return true
// 		}
// 	}
// 	return false
// }

// func startFetchingPosts(groupID string) {
// 	ticker := time.NewTicker(5 * time.Second)
// 	go func() {
// 		for {
// 			<-ticker.C
// 			fetchPosts(groupID)
// 		}
// 	}()
// }

// func printBlue(text string) {
// 	fmt.Printf("\033[34m%s\033[0m", text)
// }

// func main() {
// 	groupName := extractGroupNameFromURL("https://vk.com/club223517383")
// 	resolvedInfo, err := resolveScreenName("b397ce84b397ce84b397ce8432b0819482bb397b397ce84d6cdd2ca964821d7fd266b76", groupName)
// 	if err != nil {
// 		log.Fatalf("Error resolving screen name: %s", err)
// 	}

// 	if resolvedInfo.ObjectType != "group" {
// 		log.Fatalf("Resolved object is not a group")
// 	}

// 	groupID := fmt.Sprintf("%d", resolvedInfo.ObjectID)

// 	startFetchingPosts(groupID)
// 	quit := make(chan os.Signal, 1)
// 	signal.Notify(quit, os.Interrupt)

// 	// Запуск фоновой задачи
// 	startFetchingPosts(groupID)

// 	// Ожидание сигнала завершения
// 	<-quit

// 	log.Println("Shutting down the application...")
// }
