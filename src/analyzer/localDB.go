package main

import (
	"errors"
	"sync"
	"time"
)

type SafeStorage interface {
	add(parameter1 string, parameter2 string, parameter3 string) error
	remove(parameter1 string, parameter2 string, parameter3 string) error
	get(parameter1 string, parameter2 string) (any, error)
	set(parameter1 string, parameter2 string) error
	lock()
	unLock()
}

// ChannelTopicTime хранение [ канал - [ топик - время новости ] ]
type ChannelTopicTime struct {
	mut        sync.RWMutex
	topicTimes map[string]map[string]time.Time
}

func (c *ChannelTopicTime) add(parameter1 string, parameter2 string, _ string) error {
	channel, topic := parameter1, parameter2

	c.topicTimes[channel][topic] = time.Now().Add(-24 * time.Hour)

	return nil

}

func (c *ChannelTopicTime) remove(parameter1 string, parameter2 string, _ string) error {
	channel, topic := parameter1, parameter2
	delete(c.topicTimes[channel], topic)
	if len(c.topicTimes[channel]) == 0 {
		delete(c.topicTimes, channel)
	}
	return nil
}

func (c *ChannelTopicTime) lock() {
	c.mut.Lock()
}

func (c *ChannelTopicTime) unLock() {
	c.mut.Unlock()
}

func (c *ChannelTopicTime) get(parameter1 string, parameter2 string) (any, error) {
	channel, topic := parameter1, parameter2
	_, ok := c.topicTimes[channel]
	if !ok {
		delete(c.topicTimes, channel)
		return nil, errors.New("can`t find channel")
	}
	_, ok = c.topicTimes[channel][topic]
	if !ok {
		delete(c.topicTimes[channel], topic)
		return nil, errors.New("can`t find topic")
	}
	return c.topicTimes[channel][topic], nil
}

func (c *ChannelTopicTime) set(parameter1 string, parameter2 string) error {
	channel, topic := parameter1, parameter2
	_, ok := c.topicTimes[channel]
	if !ok {
		delete(c.topicTimes, channel)
		return errors.New("can`t find channel")
	}
	_, ok = c.topicTimes[channel][topic]
	if !ok {
		delete(c.topicTimes[channel], topic)
		return errors.New("can`t find topic")
	}
	c.topicTimes[channel][topic] = time.Now()
	return nil
}

// ChannelTopicCount хранение [ название канала - [ топик  - кол-во ссылок] ]
type ChannelTopicCount struct {
	mut         sync.RWMutex
	topicCounts map[string]map[string]uint64
	topicTime   *ChannelTopicTime
}

func (c *ChannelTopicCount) add(parameter1 string, parameter2 string, _ string) error {
	channel, topic := parameter1, parameter2
	c.topicCounts[channel][topic] += 1
	return nil
}

func (c *ChannelTopicCount) remove(parameter1 string, parameter2 string, _ string) error {
	channel, topic := parameter1, parameter2

	if c.topicCounts[channel][topic] <= 1 {
		delete(c.topicCounts[channel], topic)
		if len(c.topicCounts[channel]) == 0 {
			delete(c.topicCounts, channel)
		}
	} else {
		c.topicCounts[channel][topic]--
	}

	return nil
}

func (c *ChannelTopicCount) lock() {
	c.mut.Lock()
}

func (c *ChannelTopicCount) unLock() {
	c.mut.Unlock()
}

func (c *ChannelTopicCount) get(parameter1 string, parameter2 string) (any, error) {
	return c.topicTime.get(parameter1, parameter2)
}

func (c *ChannelTopicCount) set(parameter1 string, parameter2 string) error {
	return c.topicTime.set(parameter1, parameter2)
}

// UsersChannelsTopics хранение [ пользователь - [ канал  -  топики ] ]
type UsersChannelsTopics struct {
	mut           sync.RWMutex
	channelTopics map[string]map[string]map[string]struct{}
	topicCount    *ChannelTopicCount
}

func (c *UsersChannelsTopics) add(parameter1 string, parameter2 string, parameter3 string) error {
	user, channel, topic := parameter1, parameter2, parameter3
	_, ok := c.channelTopics[user][channel][topic]
	if !ok {
		err := c.topicCount.add(channel, topic, "")
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *UsersChannelsTopics) remove(parameter1 string, parameter2 string, parameter3 string) error {

	user, channel, topic := parameter1, parameter2, parameter3

	_, ok := c.channelTopics[user]
	if !ok {
		delete(c.channelTopics, user)
		return errors.New("can`t find user")
	}

	_, ok = c.channelTopics[user][channel]
	if !ok {
		delete(c.channelTopics[user], channel)
		return errors.New("can`t find channel in this user`s channels")
	}

	_, ok = c.channelTopics[user][channel][topic]
	if ok {
		err := c.topicCount.remove(channel, topic, "")
		if err != nil {
			return err
		}
	}

	delete(c.channelTopics[user][channel], topic)
	if len(c.channelTopics[user][channel]) == 0 {
		delete(c.channelTopics[user], channel)
		if len(c.channelTopics[user]) == 0 {
			delete(c.channelTopics, user)
		}
	}

	return nil
}

func (c *UsersChannelsTopics) lock() {
	c.mut.Lock()
}

func (c *UsersChannelsTopics) unLock() {
	c.mut.Unlock()
}

func (c *UsersChannelsTopics) get(parameter1 string, parameter2 string) (any, error) {
	return c.topicCount.get(parameter1, parameter2)
}

func (c *UsersChannelsTopics) set(parameter1 string, parameter2 string) error {
	return c.topicCount.set(parameter1, parameter2)
}
