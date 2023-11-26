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
	vkObjectInfoRequest = "https://api.vk.com/method/utils.resolveScreenName?screen_name=%s&access_token=%s&v=%s"
	VKGetPostRequest    = "https://api.vk.com/method/wall.get?owner_id=-%s&count=%s&access_token=%s&v=%s"
	VKPostLink          = "https://vk.com/wall-%s_%d"
	apiVersion          = "5.154"
	VKNWorkers          = 30
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

type VKHandler struct {
	accessToken       string
	groupSegmentation map[int][]string
	VKChans           []chan []string
}

func (vk *VKHandler) handleUpdates() {
	groups, err := dataBase.getVKPublic()
	if err != nil {
		log.Println(err.Error())
	}

	for _, group := range groups {
		vk.initLastPostID(group, vk.accessToken)
	}

	vkChans := make([]chan []string, VKNWorkers)
	for i := 0; i < VKNWorkers; i++ {
		vkChans[i] = make(chan []string)
		go vk.vkWorker(vkChans[i])
	}

	segmentation := make(map[int][]string)

	vk.VKChans = vkChans
	vk.groupSegmentation = segmentation

	go vk.refreshGroups(time.Second * 30)
}

func (vk *VKHandler) vkWorker(groups chan []string) {
	curGroups := make([]string, 0)
	for {
		select {
		case curGroups = <-groups:
		default:
		}

		for _, group := range curGroups {
			posts := vk.fetchPosts(group)
			if posts == nil {
				continue
			}
			hsh := getHash(group) % NWorkers
			for _, post := range posts {
				msg := workEvent{
					application: VK,
					channel:     group,
					channelID:   group,
					text:        post.Text,
					link:        post.URL,
					messageID:   string(rune(post.ID)),
				}
				workChans[hsh] <- msg
			}
		}

		time.Sleep(time.Second * 30)
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
			vk.VKChans[i] <- cur
		}
	}

	time.Sleep(period)
}

func (vk *VKHandler) initLastPostID(groupID, accessToken string) {
	posts, err := vk.getLatestPosts(groupID, "1")
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

func (vk *VKHandler) getLatestPosts(groupID string, count string) ([]Post, error) {
	url := fmt.Sprintf(VKGetPostRequest, groupID, count, vk.accessToken, apiVersion)

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

func (vk *VKHandler) fetchPosts(groupID string) []Post {
	posts, err := vk.getLatestPosts(groupID, "1")
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

func getVKInfo(groupLink, accessToken string) (string, int, error) {
	name := extractGroupNameFromURL(groupLink)
	if name == "" {
		return "", -1, errors.New("Неправильная ссылка")

	}
	resolve, err := resolveScreenName(name, accessToken)
	if err != nil {
		log.Println(err.Error())
		return "", -1, err
	}
	return resolve.ObjectType, resolve.ObjectID, nil
}
