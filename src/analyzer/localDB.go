package main

import (
	"errors"
	"sync"
	"time"
)

type Property string

const (
	Time     Property = "time"
	Channels Property = "channels"
	Topic    Property = "topic"
	Topics   Property = "topics"
	Count    Property = "count"
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
}

// ChannelTopicTime хранение [ канал - [ топик - время новости ] ]
type ChannelTopicTime struct {
	mut        sync.RWMutex
	topicTimes map[string]map[string]time.Time
}

func (c *ChannelTopicTime) add(parameter1 string, parameter2 string, _ string, property Property) error {
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
		c.topicTimes[channel][topic] = time.Now().Add(-24 * time.Hour)
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

func (c *ChannelTopicTime) get(parameter1 string, parameter2 string, property Property) (any, error) {
	c.mut.Lock()
	c.mut.Unlock()

	channel, topic := parameter1, parameter2

	_, ok := c.topicTimes[channel]
	if !ok {
		delete(c.topicTimes, channel)
		return nil, invalidChannelError
	}

	switch property {
	case Time:
		_, ok = c.topicTimes[channel][topic]
		if !ok {
			delete(c.topicTimes[channel], topic)
			return nil, invalidTopicError
		}

		return c.topicTimes[channel][topic], nil
	case Topics:
		return c.topicTimes[channel], nil
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

func (c *ChannelTopicCount) add(parameter1 string, parameter2 string, _ string, property Property) error {
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
		c.topicCounts[channel][topic] += 1
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
	_, ok = c.topicCounts[channel][topic]
	if !ok {
		delete(c.topicCounts[channel], topic)
		return nil, invalidTopicError
	}

	switch property {
	case Count:
		return c.topicCounts[channel][topic], nil
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

func (u *UsersChannelsTopics) add(parameter1 string, parameter2 string, parameter3 string, _ Property) error {
	u.mut.Lock()
	defer u.mut.Unlock()
	user, channel, topic := parameter1, parameter2, parameter3
	_ = u.channelTopics[user][channel][topic]
	return nil
}

func (u *UsersChannelsTopics) remove(parameter1 string, parameter2 string, parameter3 string, property Property) error {
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
	var user, channel = parameter1, parameter2
	switch property {
	case Channels:
		_, ok := u.channelTopics[user]
		if !ok {
			delete(u.channelTopics, user)
			return nil, invalidUserError
		}

		return u.channelTopics[user], nil
	case Topics:

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

	default:
		return nil, invalidPropertyError
	}

	return nil, unreachableCodeError

}

func (u *UsersChannelsTopics) set(_ string, _ string, _ Property) error {
	return notImplementedError
}
