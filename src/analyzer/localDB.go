package main

import (
	"sync"
	"time"
)

type SafeStorage interface {
	add(parameter1 string, parameter2 string, parameter3 string)
	remove(parameter1 string, parameter2 string, parameter3 string)
}

// ChannelTopicTime хранение [ канал - [ топик - время новости ] ]
type ChannelTopicTime struct {
	mut        sync.RWMutex
	topicTimes map[string]map[string]time.Time
}

func (c *ChannelTopicTime) add(parameter1 string, parameter2 string, _ string) {
	channel, topic := parameter1, parameter2

	c.topicTimes[channel][topic] = time.Now().Add(-24 * time.Hour)

}

func (c *ChannelTopicTime) remove(parameter1 string, parameter2 string, _ string) {
	channel, topic := parameter1, parameter2
	delete(c.topicTimes[channel], topic)
	if len(c.topicTimes[channel]) == 0 {
		delete(c.topicTimes, channel)
	}
}

// ChannelTopicCount хранение [ название канала - [ топик  - кол-во ссылок] ]
type ChannelTopicCount struct {
	mut         sync.RWMutex
	topicCounts map[string]map[string]uint64
	topicTime   *ChannelTopicTime
}

func (c *ChannelTopicCount) add(parameter1 string, parameter2 string, _ string) {
	channel, topic := parameter1, parameter2
	c.topicCounts[channel][topic] += 1
	if c.topicCounts[channel][topic] == 1 {
		topicTimes.add(channel, topic, "")
	}
}

func (c *ChannelTopicCount) remove(parameter1 string, parameter2 string, _ string) {
	channel, topic := parameter1, parameter2

	if c.topicCounts[channel][topic] <= 1 {
		delete(c.topicCounts[channel], topic)
		topicTimes.remove(channel, topic, "")
		if len(c.topicCounts[channel]) == 0 {
			delete(c.topicCounts, channel)
		}
	} else {
		c.topicCounts[channel][topic]--
	}
}

// UsersChannelsTopics хранение [ пользователь - [ канал  -  топики ] ]
type UsersChannelsTopics struct {
	mut           sync.RWMutex
	channelTopics map[string]map[string]map[string]struct{}
	topicCount    *ChannelTopicCount
}

func (u *UsersChannelsTopics) add(parameter1 string, parameter2 string, parameter3 string) {
	user, channel, topic := parameter1, parameter2, parameter3
	_, ok := u.channelTopics[user][channel][topic]
	if !ok {
		u.topicCount.add(channel, topic, "")
	}
}

func (u *UsersChannelsTopics) remove(parameter1 string, parameter2 string, parameter3 string) {
	user, channel, topic := parameter1, parameter2, parameter3
	_, ok := u.channelTopics[user][channel][topic]
	if ok {
		u.topicCount.remove(channel, topic, "")
	}
	delete(u.channelTopics[user][channel], topic)
	if len(u.channelTopics[user][channel]) == 0 {
		delete(u.channelTopics[user], channel)
		if len(u.channelTopics[user]) == 0 {
			delete(u.channelTopics, user)
		}
	}
}
