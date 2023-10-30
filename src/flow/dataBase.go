package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	Delay = time.Minute
)

type Message struct {
	User    string `json:"user"`
	Link    int    `json:"link"`
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
	addDelayedMessage(messages Message) error
	getDelayedMessages(user string) ([]Message, error)
	pauseUser(user string) error
	unpauseUser(user string) error
	isPaused(user string) (bool, error)
	getID(user string) (int64, error)
}

type DataBase struct {
	*sql.DB
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
}

//go:embed migrations/init.sql
var initScript string

func NewDatabase(cfg DBConfig) (*DataBase, error) {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port,
		cfg.DB)
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(initScript); err != nil {
		return nil, err
	}
	return &DataBase{db}, nil
}

func (d *DataBase) add(user, channel, topic string) error {
	row := d.QueryRow("SELECT COUNT(*) FROM channels WHERE nickname=$1 AND channel=$2 AND topic=$3",
		user, channel, topic)

	var count int64
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	if count != 0 {
		return nil
	}

	_, err = d.Exec(
		"INSERT INTO channels (nickname, channel, topic, last_time) VALUES ($1,$2,$3,$4)",
		user,
		channel,
		topic,
		time.Now().Add(-Delay),
	)

	return err
}

func (d *DataBase) removeTopic(user, channel, topic string) error {
	_, err := d.Exec(
		"DELETE FROM channels WHERE nickname = $1 AND channel = $2 AND topic = $3",
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
	_, err := d.Exec(
		"DELETE FROM channels WHERE nickname = $1 AND channel = $2",
		user,
		channel,
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *DataBase) getTopics(channel string) ([]string, error) {
	rows, err := d.Query(
		"SELECT topic FROM channels WHERE channel = $1;",
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
	rows, err := d.Query(
		"SELECT channel, topic FROM channels WHERE nickname = $1",
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
		rows, err := d.Query(
			"SELECT nickname FROM channels WHERE channel = $1 AND topic = $2 AND last_time < $3 ",
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
	_, err := d.Exec(
		"UPDATE channels SET time = $1 WHERE  nickname = $2 AND channel = $3 AND topic = $4",
		time.Now(),
		user,
		channel,
		topic,
	)
	return err
}

func (d *DataBase) containsChannel(channel string) (bool, error) {
	row := d.QueryRow(
		"SELECT COUNT(*) FROM channels WHERE channel=$1",
		channel,
	)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	return count != 0, nil
}

func (d *DataBase) addDelayedMessage(message Message) error {
	_, err := d.Exec(
		"INSERT INTO messages (nickname, link, channel, topic, summary) VALUES ($1,$2,$3,$4,"+
			"$5)",
		message.User,
		message.Link,
		message.Channel,
		message.Topic,
		message.Summary,
	)

	return err
}

func (d *DataBase) getDelayedMessages(user string) ([]Message, error) {

	rows, err := d.Query(
		"SELECT * FROM messages WHERE nickname = $1",
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

	_, err = d.Exec(
		"DELETE FROM channels WHERE nickname = $1",
		user,
	)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (d *DataBase) isPaused(user string) (bool, error) {
	row := d.QueryRow(
		"SELECT paused FROM users WHERE nickname = $1",
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
	_, err = d.Exec(
		"UPDATE users SET pause = $1 WHERE nickname = $2 ",
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
	_, err = d.Exec(
		"UPDATE users SET pause = $1 WHERE  nickname = $2 ",
		false,
		user,
	)
	return err
}

func (d *DataBase) getID(user string) (int64, error) {
	row := d.QueryRow(
		"SELECT id FROM users WHERE nickname=$1",
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
	row := d.QueryRow(
		"SELECT COUNT(*) FROM users WHERE nickname=$1",
		user,
	)
	var count int64
	if err := row.Scan(&count); err != nil || count != 0 {
		return err
	}

	_, err := d.Exec(
		"INSERT INTO users (id, nickname, paused) VALUES ($1,$2,$3)",
		id,
		user,
		false,
	)

	return err
}
