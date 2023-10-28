package main

import (
	"database/sql"
	"time"
)

const (
	Delay = time.Minute
)

type Message struct {
	User    string `json:"user"`
	Link    string `json:"link"`
	Channel string `json:"channel"`
	Topic   string `json:"topic"`
	Summary string `json:"summary"`
}

type LocalStorage interface {
	addUser(user string, id int64) error
	add(user, channel, topic string) error
	removeTopic(user, channel, topic string) error
	removeChannel(user, channel string) error
	getTopics(channel string) ([]string, error)
	getUserInfo(user string) (map[string][]string, error)
	getUsers(channel string, topics []string) (map[string][]string, error)
	setTime(user, channel, topic string) error
	containsChannel(channel string) (bool, error)
	addDelayedMessages(messages []Message) error
	getDelayedMessages(user string) ([]Message, error)
	pauseUser(user string) error
	unpauseUser(user string) error
	isPaused(user string) (bool, error)
	getID(user string) (int64, error)
}

type Table struct {
	Storage *sql.DB
	Name    string
}

type DataBase struct {
	Channels        Table
	DelayedMessages Table
	Users           Table
}

func (d *DataBase) add(user, channel, topic string) error {
	row := d.Channels.Storage.QueryRow(
		"SELECT COUNT(*) FROM $1 WHERE nickname=$2 AND channel=$3 AND topic=$4",
		d.Channels.Name,
		user,
		channel,
		topic,
	)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	if count != 0 {
		return nil
	}

	_, err = d.Channels.Storage.Exec(
		"INSERT INTO $1 (nickname, channel, topic, last_time) VALUES ($2,$3,$4,$5)",
		d.Channels.Name,
		user,
		channel,
		topic,
		time.Now().Add(-Delay),
	)

	return err
}

func (d *DataBase) removeTopic(user, channel, topic string) error {
	_, err := d.Channels.Storage.Exec(
		"DELETE FROM $1 WHERE nickname = $2 AND channel = $3 AND topic $4",
		d.Channels.Name,
		user,
		channel,
		topic,
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *DataBase) removeChannel(user, channel string) error {
	_, err := d.Channels.Storage.Exec(
		"DELETE FROM $1 WHERE nickname = $2 AND channel = $3",
		d.Channels.Name,
		user,
		channel,
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *DataBase) getTopics(channel string) ([]string, error) {
	rows, err := d.Channels.Storage.Query(
		"SELECT topic FROM $1 WHERE channel = $2;",
		d.Channels.Name,
		channel,
	)
	if err != nil {
		return nil, err
	}
	var topics []string
	for rows.Next() {
		var topic string
		err = rows.Scan(&topic)
		if err != nil {
			return nil, err
		}
		topics = append(topics, topic)
	}
	return topics, nil
}

func (d *DataBase) getUserInfo(user string) (map[string][]string, error) {
	rows, err := d.DelayedMessages.Storage.Query(
		"SELECT channel, topic FROM $1 WHERE nickname = $2",
		d.Channels.Name,
		user,
	)
	if err != nil {
		return nil, err
	}
	answer := make(map[string][]string)
	for rows.Next() {
		var channel, topic string
		err = rows.Scan(&channel, &topic)
		if err != nil {
			return nil, err
		}
		curTopics := answer[channel]
		curTopics = append(curTopics, topic)
		answer[channel] = curTopics
	}

	return answer, nil
}

func (d *DataBase) getUsers(channel string, topics []string) (map[string][]string, error) {
	answer := make(map[string][]string)
	for _, topic := range topics {
		rows, err := d.Channels.Storage.Query(
			"SELECT nickname FROM $1 WHERE channel = $2 AND topic = $3 AND last_time < $4 ",
			d.Channels.Name,
			channel,
			topic,
			time.Now().Add(-Delay),
		)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var user string
			err = rows.Scan(&user)
			if err != nil {
				return nil, err
			}
			answer[user] = append(answer[user], topic)
		}
	}

	return answer, nil
}

func (d *DataBase) setTime(user, channel, topic string) error {
	_, err := d.Channels.Storage.Exec(
		"UPDATE $1 SET time = $2 WHERE  nickname = $3 AND channel = $4 AND topic = $5",
		d.Channels.Name,
		time.Now(),
		user,
		channel,
		topic,
	)
	return err
}

func (d *DataBase) containsChannel(channel string) (bool, error) {
	row := d.Channels.Storage.QueryRow(
		"SELECT COUNT(*) FROM $1 WHERE channel=$2",
		d.Channels.Name,
		channel,
	)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

func (d *DataBase) addDelayedMessages(messages []Message) error {
	for _, message := range messages {
		_, err := d.Channels.Storage.Exec(
			"INSERT INTO $1 (nickname, link, channel, topic, summary) VALUES ($2,$3,$4,$5,$6)",
			d.DelayedMessages.Name,
			message.User,
			message.Link,
			message.Channel,
			message.Topic,
			message.Summary,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *DataBase) getDelayedMessages(user string) ([]Message, error) {

	rows, err := d.DelayedMessages.Storage.Query(
		"SELECT * FROM $1 WHERE nickname = $2",
		d.DelayedMessages.Name,
		user,
	)
	if err != nil {
		return nil, err
	}
	var messages []Message
	for rows.Next() {
		var message Message
		err = rows.Scan(&message.User, &message.Link, &message.Channel, &message.Topic, &message.Summary)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	_, err = d.Channels.Storage.Exec(
		"DELETE FROM $1 WHERE nickname = $2",
		d.Channels.Name,
		user,
	)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (d *DataBase) isPaused(user string) (bool, error) {
	row := d.Users.Storage.QueryRow(
		"SELECT paused FROM $1 WHERE nickname = $2",
		d.Users.Name,
		user,
	)
	var isPaused bool
	err := row.Scan(&isPaused)
	if err != nil {
		return false, err
	}

	return isPaused, nil
}

func (d *DataBase) pauseUser(user string) error {
	isPaused, err := d.isPaused(user)
	if err != nil {
		return err
	}
	if isPaused {
		return nil
	}
	_, err = d.Users.Storage.Exec(
		"UPDATE $1 SET pause = $2 WHERE  nickname = $3 ",
		d.Users.Name,
		true,
		user,
	)
	return err
}

func (d *DataBase) unpauseUser(user string) error {
	isPaused, err := d.isPaused(user)
	if err != nil {
		return err
	}
	if !isPaused {
		return nil
	}
	if isPaused {
		return nil
	}
	_, err = d.Users.Storage.Exec(
		"UPDATE $1 SET pause = $2 WHERE  nickname = $3 ",
		d.Users.Name,
		false,
		user,
	)
	return err
}

func (d *DataBase) getID(user string) (int64, error) {
	row := d.Users.Storage.QueryRow(
		"SELECT id FROM $1 WHERE nickname=$2",
		d.Users.Name,
		user,
	)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		return -1, err
	}

	return id, nil
}

func (d *DataBase) addUser(user string, id int64) error {
	row := d.Users.Storage.QueryRow(
		"SELECT COUNT(*) FROM $1 WHERE nickname=$2",
		d.Channels.Name,
		user,
	)
	var count int64
	if err := row.Scan(&count); err != nil || count != 0 {
		return err
	}

	_, err := d.Channels.Storage.Exec(
		"INSERT INTO $1 (id, nickname, paused) VALUES ($2,$3,$4)",
		d.Users.Name,
		id,
		user,
		false,
	)

	return err
}
