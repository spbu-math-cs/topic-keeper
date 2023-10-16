package main

import (
	"errors"
	"sync"
	"time"
)

type Property string

const (
	Time                 Property = "time"
	Channels             Property = "channels"
	Topic                Property = "topic"
	Topics               Property = "topics"
	Count                Property = "count"
	UsersTopicsByChannel Property = "users and their topics in channel"
)

var (
	invalidPropertyError = errors.New("invalid property type")
	invalidChannelError  = errors.New("can`t find channel")
	invalidTopicError    = errors.New("can`t find topic")
	invalidUserError     = errors.New("can`t find user")
	unreachableCodeError = errors.New("unreachable code")
	notImplementedError  = errors.New("this function is not implemented")
)

type SafeStorage interface {
	add(parameter1 string, parameter2 string, parameter3 string, property Property) error
	remove(parameter1 string, parameter2 string, parameter3 string, property Property) error
	get(parameter1 string, parameter2 string, property Property) (any, error)
	set(parameter1 string, parameter2 string, property Property) error
	create()
}

// ChannelTopicTime хранение [ канал - [ топик - время новости ] ]
type ChannelTopicTime struct {
	interval   time.Duration
	mut        sync.RWMutex
	topicTimes map[string]map[string]time.Time
}

func (c *ChannelTopicTime) create() {
	c.topicTimes = make(map[string]map[string]time.Time)
	c.mut = sync.RWMutex{}
}

func (c *ChannelTopicTime) add(parameter1 string, parameter2 string, _ string, property Property) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	channel, topic := parameter1, parameter2

	switch property {
	case Topic:
		_, ok := c.topicTimes[channel]
		if !ok {
			m1 := make(map[string]time.Time)
			m1[topic] = time.Now().Add(-c.interval * 2)
			c.topicTimes[channel] = m1
			return nil
		}
		_, ok = c.topicTimes[channel][topic]
		if !ok {
			c.topicTimes[channel][topic] = time.Now().Add(-24 * time.Hour)
			return nil
		}

		return nil
	default:
		return invalidPropertyError
	}

	return unreachableCodeError
}

func (c *ChannelTopicTime) remove(parameter1 string, parameter2 string, _ string, property Property) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	channel, topic := parameter1, parameter2
	switch property {
	case Topic:
		_, ok := c.topicTimes[channel]
		if !ok {
			delete(c.topicTimes, channel)
			return invalidChannelError
		}

		_, ok = c.topicTimes[channel][topic]
		if !ok {
			delete(c.topicTimes[channel], topic)
			return invalidTopicError
		}

		delete(c.topicTimes[channel], topic)
		if len(c.topicTimes[channel]) == 0 {
			delete(c.topicTimes, channel)
		}
		return nil
	default:
		return invalidPropertyError
	}

	return unreachableCodeError
}

func (c *ChannelTopicTime) get(parameter1 string, _ string, property Property) (any, error) {
	c.mut.Lock()
	c.mut.Unlock()

	switch property {
	case Topics:
		channel := parameter1
		_, ok := c.topicTimes[channel]
		if !ok {
			delete(c.topicTimes, channel)
			return nil, invalidChannelError
		}
		var m []string

		for topic, last := range c.topicTimes[channel] {
			var cur = time.Now().Add(-c.interval)
			if cur.After(last) {
				m = append(m, topic)
				c.topicTimes[channel][topic] = time.Now()
			}
		}
		return m, nil
	default:
		return nil, invalidPropertyError
	}

	return nil, unreachableCodeError
}

func (c *ChannelTopicTime) set(parameter1 string, parameter2 string, property Property) error {
	c.mut.Lock()
	c.mut.Unlock()

	channel, topic := parameter1, parameter2

	_, ok := c.topicTimes[channel]
	if !ok {
		delete(c.topicTimes, channel)
		return invalidChannelError
	}

	switch property {
	case Time:
		_, ok = c.topicTimes[channel][topic]
		if !ok {
			delete(c.topicTimes[channel], topic)
			return invalidTopicError
		}

		c.topicTimes[channel][topic] = time.Now()

		return nil

	default:
		return invalidPropertyError
	}

	return unreachableCodeError
}

// ChannelTopicCount хранение [ название канала - [ топик  - кол-во ссылок] ]
type ChannelTopicCount struct {
	mut         sync.RWMutex
	topicCounts map[string]map[string]uint64
}

func (c *ChannelTopicCount) create() {
	c.topicCounts = make(map[string]map[string]uint64)
	c.mut = sync.RWMutex{}
}

func (c *ChannelTopicCount) add(parameter1 string, parameter2 string, _ string, property Property) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	channel, topic := parameter1, parameter2
	m1 := make(map[string]uint64)

	switch property {
	case Topic:
		_, ok := c.topicCounts[channel]
		if !ok {
			m1[topic] = 1
			c.topicCounts[channel] = m1
			return nil
		}
		_, ok = c.topicCounts[channel][topic]
		if !ok {
			c.topicCounts[channel][topic] = 1
			return nil
		}

		c.topicCounts[channel][topic]++
		return nil

	default:
		return invalidPropertyError
	}

	return unreachableCodeError

}

func (c *ChannelTopicCount) remove(parameter1 string, parameter2 string, _ string, property Property) error {
	c.mut.Lock()
	defer c.mut.Unlock()

	channel, topic := parameter1, parameter2

	switch property {
	case Topic:
		_, ok := c.topicCounts[channel]
		if !ok {
			delete(c.topicCounts, channel)
			return invalidChannelError
		}
		_, ok = c.topicCounts[channel][topic]
		if !ok {
			delete(c.topicCounts[channel], topic)
			return invalidTopicError
		}
		c.topicCounts[channel][topic] -= 1

		if c.topicCounts[channel][topic] == 0 {
			delete(c.topicCounts[channel], topic)
		}
		if len(c.topicCounts[channel]) == 0 {
			delete(c.topicCounts, channel)
		}
		return nil
	default:
		return invalidPropertyError
	}

	return unreachableCodeError
}

func (c *ChannelTopicCount) get(parameter1 string, parameter2 string, property Property) (any, error) {
	c.mut.Lock()
	defer c.mut.Unlock()

	channel, topic := parameter1, parameter2

	_, ok := c.topicCounts[channel]
	if !ok {
		delete(c.topicCounts, channel)
		return nil, invalidChannelError
	}

	switch property {
	case Count:
		_, ok = c.topicCounts[channel][topic]
		if !ok {
			delete(c.topicCounts[channel], topic)
			return nil, invalidTopicError
		}
		return c.topicCounts[channel][topic], nil
	case Topics:
		return c.topicCounts[channel], nil
	default:
		return nil, invalidPropertyError
	}

	return nil, unreachableCodeError
}

func (c *ChannelTopicCount) set(_ string, _ string, _ Property) error {
	return notImplementedError
}

// UsersChannelsTopics хранение [ пользователь - [ канал  -  топики ] ]
type UsersChannelsTopics struct {
	mut           sync.RWMutex
	channelTopics map[string]map[string]map[string]struct{}
}

func (u *UsersChannelsTopics) create() {
	u.channelTopics = make(map[string]map[string]map[string]struct{})
	u.mut = sync.RWMutex{}
}

func (u *UsersChannelsTopics) add(parameter1 string, parameter2 string, parameter3 string, _ Property) error {
	u.mut.Lock()
	defer u.mut.Unlock()
	user, channel, topic := parameter1, parameter2, parameter3
	_, ok := u.channelTopics[user]
	if !ok {
		m1 := make(map[string]map[string]struct{})
		m2 := make(map[string]struct{})
		m2[topic] = struct{}{}
		m1[channel] = m2
		u.channelTopics[user] = m1
		return nil
	}
	_, ok = u.channelTopics[user][channel]
	if !ok {
		m2 := make(map[string]struct{})
		m2[topic] = struct{}{}
		u.channelTopics[user][channel] = m2
		return nil
	}
	_, ok = u.channelTopics[user][channel][topic]
	u.channelTopics[user][channel][topic] = struct{}{}

	return nil
}

func (u *UsersChannelsTopics) remove(parameter1 string, parameter2 string, parameter3 string, _ Property) error {
	u.mut.Lock()
	defer u.mut.Unlock()

	user, channel, topic := parameter1, parameter2, parameter3
	_, ok := u.channelTopics[user]
	if !ok {
		delete(u.channelTopics, user)
		return invalidUserError
	}

	_, ok = u.channelTopics[user][channel]
	if !ok {
		delete(u.channelTopics[user], channel)
		return invalidChannelError
	}

	_, ok = u.channelTopics[user][channel][topic]
	if !ok {
		delete(u.channelTopics[user][channel], topic)
		return invalidTopicError
	}

	delete(u.channelTopics[user][channel], topic)

	if len(u.channelTopics[user][channel]) == 0 {
		delete(u.channelTopics[user], channel)
	}
	if len(u.channelTopics[user]) == 0 {
		delete(u.channelTopics, user)
	}
	return nil
}

func (u *UsersChannelsTopics) get(parameter1 string, parameter2 string, property Property) (any, error) {
	u.mut.Lock()
	defer u.mut.Unlock()
	switch property {
	case Channels:
		var user = parameter1
		_, ok := u.channelTopics[user]
		if !ok {
			delete(u.channelTopics, user)
			return nil, invalidUserError
		}

		return u.channelTopics[user], nil
	case Topics:
		var user, channel = parameter1, parameter2

		_, ok := u.channelTopics[user]
		if !ok {
			delete(u.channelTopics, user)
			return nil, invalidUserError
		}

		_, ok = u.channelTopics[user][channel]
		if !ok {
			delete(u.channelTopics[user], channel)
			return nil, invalidChannelError
		}
		return u.channelTopics[user][channel], nil

	case UsersTopicsByChannel:

		channel := parameter1
		var m = make(map[string]map[string]struct{})

		for user, userChannels := range u.channelTopics {
			topics, ok := userChannels[channel]
			if !ok {
				delete(userChannels, channel)
				continue
			}
			m[user] = topics
		}

		return m, nil

	default:
		return nil, invalidPropertyError
	}

	return nil, unreachableCodeError

}

func (u *UsersChannelsTopics) set(_ string, _ string, _ Property) error {
	return notImplementedError
}
