package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"
)

const (
	vkObjectInfoRequest   = "https://api.vk.com/method/utils.resolveScreenName?screen_name=%s&access_token=%s&v=%s"
	vkGetGroupNameRequest = "https://api.vk.com/method/groups.getById?group_id=%d&access_token=%s&v=%s"
	VKGetPostRequest      = "https://api.vk.com/method/wall.get?owner_id=-%s&count=%s&access_token=%s&v=%s"
	VKPostLink            = "https://vk.com/wall-%s_%d"
	apiVersion            = "5.154"
	VKNWorkers            = 15
	VKNHistoryWorkers     = 3
)

var (
	VKHistoryChans []chan UserHistory
	VKChans        []chan []string
)

type Post struct {
	ID          int       `json:"id"`
	Text        string    `json:"text"`
	FetchedTime time.Time `json:"fetched_time"`
	URL         string    `json:"url"`
}

type VKAPIResponse struct {
	Response struct {
		Items []Post `json:"items"`
	} `json:"response"`
}

type ResolvedInfo struct {
	ObjectType string `json:"type"`
	ObjectID   int    `json:"object_id"`
}

type UserHistory struct {
	user       string
	link       string
	publicName string
	publicID   string
	postsCount int
}

type VKHandler struct {
	accessToken       string
	groupSegmentation map[int][]string
}

func (vk *VKHandler) handleUpdates() {
	groups, err := dataBase.getVKPublic()
	if err != nil {
		log.Println(err.Error())
	}

	for _, group := range groups {
		vk.initLastPostID(group, vk.accessToken)
	}

	VKChans = make([]chan []string, VKNWorkers)
	for i := 0; i < VKNWorkers; i++ {
		VKChans[i] = make(chan []string)
		go vkWorker(VKChans[i])
	}

	VKHistoryChans = make([]chan UserHistory, VKNHistoryWorkers)
	for i := 0; i < VKNHistoryWorkers; i++ {
		VKHistoryChans[i] = make(chan UserHistory)
		go vkHistoryWorker(VKHistoryChans[i])
	}

	segmentation := make(map[int][]string)

	vk.groupSegmentation = segmentation

	go vk.refreshGroups(time.Second * 60)
}

func vkWorker(groups chan []string) {
	curGroups := make([]string, 0)
	for {
		select {
		case curGroups = <-groups:
		default:
		}

		for _, group := range curGroups {
			posts := fetchPosts(group)
			if posts == nil {
				continue
			}
			hsh := getHash(group) % NWorkers
			for _, post := range posts {
				groupName, err := dataBase.getVKPublicNameByID(group)
				if err != nil {
					log.Println(err.Error())
					continue
				}
				msg := workEvent{
					application:    VK,
					channel:        groupName,
					channelID:      group,
					text:           post.Text,
					link:           post.URL,
					messageID:      string(rune(post.ID)),
					historyRequest: nil,
				}
				workChans[hsh] <- msg
			}
		}

		time.Sleep(time.Second * 60)
	}
}

func vkHistoryWorker(historyRequestChan chan UserHistory) {
	for request := range historyRequestChan {

		count := fmt.Sprintf("%d", request.postsCount)

		posts, err := getLatestPosts(request.publicID, count)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		hsh := getHash(request.publicName) % NWorkers

		for _, post := range posts {
			workChans[hsh] <- workEvent{
				application:    VK,
				channel:        request.publicName,
				channelID:      request.publicID,
				text:           post.Text,
				link:           fmt.Sprintf(VKPostLink, request.publicID, post.ID),
				messageID:      fmt.Sprintf("%d", post.ID),
				historyRequest: &historyRequest{request.user},
			}

		}
	}
}

func (vk *VKHandler) refreshGroups(period time.Duration) {
	curGroups, err := dataBase.getVKPublic()
	log.Println(curGroups)
	if err != nil {
		log.Println(err.Error())
		return
	}

	curSegmentation := make(map[int][]string)
	for _, group := range curGroups {
		hash := int(getHash(group) % VKNWorkers)
		curSegmentation[hash] = append(curSegmentation[hash], group)
	}

	for i := 0; i < VKNWorkers; i++ {
		cur := curSegmentation[i]
		sort.Strings(cur)
		was := vk.groupSegmentation[i]
		sort.Strings(was)
		if !slices.Equal(was, cur) {
			VKChans[i] <- cur
		}
	}

	time.Sleep(period)
}

func (vk *VKHandler) initLastPostID(groupID, accessToken string) {
	posts, err := getLatestPosts(groupID, "1")
	if err != nil {
		log.Printf("Error fetching initial post: %s", err)
		return
	}

	if len(posts) > 0 {
		err := dataBase.updateVKLastPostID(groupID, posts[0].ID)
		if err != nil {
			log.Println(err.Error())
			return
		}
	}
}

func getLatestPosts(groupID string, count string) ([]Post, error) {
	url := fmt.Sprintf(VKGetPostRequest, groupID, count, vkToken, apiVersion)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp VKAPIResponse
	err = json.Unmarshal(body, &apiResp)
	if err != nil {
		return nil, err
	}

	return apiResp.Response.Items, nil
}

func fetchPosts(groupID string) []Post {
	posts, err := getLatestPosts(groupID, "1")
	if err != nil {
		log.Printf("Error fetching posts: %s", err)
		return nil
	}

	lastPostID, err := dataBase.getVKLastPostID(groupID)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	var ans []Post
	for _, post := range posts {
		if post.ID > lastPostID {
			lastPostID = post.ID
			post.FetchedTime = time.Now()
			post.URL = fmt.Sprintf(VKPostLink, groupID, post.ID)
			log.Printf("New post fetched: %+v\n", post)
			ans = append(ans, post)
		}
	}

	if err := dataBase.updateVKLastPostID(groupID, lastPostID); err != nil {
		return nil
	}

	return ans
}

func resolveScreenName(screenName, accessToken string) (*ResolvedInfo, error) {
	requestURL := fmt.Sprintf(vkObjectInfoRequest, screenName, accessToken, apiVersion)

	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Response ResolvedInfo `json:"response"`
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return &result.Response, nil
}

type GroupsGetByIdResponse struct {
	Groups []Group `json:"groups"`
}

type Group struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	ScreenName string `json:"screen_name"`
	IsClosed   int    `json:"is_closed"`
	Type       string `json:"type"`
	Photo50    string `json:"photo_50"`
	Photo100   string `json:"photo_100"`
	Photo200   string `json:"photo_200"`
}

func getGroupName(groupID int, accessToken string) (string, error) {
	url := fmt.Sprintf(vkGetGroupNameRequest, groupID, accessToken, apiVersion)
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	var result struct {
		Response GroupsGetByIdResponse `json:"response"`
	}
	body, err := ioutil.ReadAll(response.Body)
	log.Println(string(body))
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}

	if len(result.Response.Groups) > 0 {
		return result.Response.Groups[0].Name, nil
	}

	return "", fmt.Errorf("group with ID %d not found", groupID)
}

func extractGroupNameFromURL(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "www.")

	if strings.HasPrefix(url, "@") {
		return strings.TrimPrefix(url, "@")
	}

	parts := strings.Split(url, "/")
	if len(parts) > 1 && parts[0] == "vk.com" {
		return parts[1]
	}

	return ""
}

func getVKInfo(groupLink, accessToken string) (string, string, int, error) {
	name := extractGroupNameFromURL(groupLink)
	if name == "" {
		return "", "", -1, errors.New("Неправильная ссылка")

	}
	resolve, err := resolveScreenName(name, accessToken)
	if err != nil {
		log.Println(err.Error())
		return "", "", -1, err
	}

	publicName, err := getGroupName(resolve.ObjectID, accessToken)
	if err != nil {
		log.Println(err.Error())
		return "", "", -1, err
	}

	return publicName, resolve.ObjectType, resolve.ObjectID, nil
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
