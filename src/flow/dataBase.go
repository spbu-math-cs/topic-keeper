package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	Delay = 10 * time.Second
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
	DB    *sql.DB
	Names TablesNames
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
}

type TablesNames struct {
	Channels string
	Users    string
	Messages string
}

//go:embed migrations/init.sql
var initScript string

func NewDatabase(cfg DBConfig, names TablesNames) (*DataBase, error) {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port,
		cfg.DB)
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(initScript); err != nil {
		return nil, err
	}
	return &DataBase{db, names}, nil
}

func (d *DataBase) add(user, channel, topic string) error {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE nickname=$1 AND channel=$2 AND topic=$3", d.Names.Channels)
	row := d.DB.QueryRow(query,
		user, channel, topic)

	var count int64
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	if count != 0 {
		return nil
	}

	query = fmt.Sprintf("INSERT INTO %s (nickname, channel, topic, last_time) VALUES ($1,$2,$3,$4)", d.Names.Channels)
	_, err = d.DB.Exec(
		query,
		user,
		channel,
		topic,
		time.Now().Add(-Delay),
	)

	return err
}

func (d *DataBase) removeTopic(user, channel, topic string) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE nickname = $1 AND channel = $2 AND topic = $3", d.Names.Channels)
	_, err := d.DB.Exec(
		query,
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
	query := fmt.Sprintf("DELETE FROM %s WHERE nickname = $1 AND channel = $2", d.Names.Channels)
	_, err := d.DB.Exec(
		query,
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
	query := fmt.Sprintf("SELECT channel, topic FROM %s WHERE nickname = $1", d.Names.Channels)
	rows, err := d.DB.Query(
		query,
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
		query := fmt.Sprintf("SELECT nickname FROM %s WHERE channel = $1 AND topic = $2 AND last_time < $3 ", d.Names.Channels)
		rows, err := d.DB.Query(
			query,
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
	query := fmt.Sprintf("UPDATE %s SET last_time = $1 WHERE  nickname = $2 AND channel = $3 AND topic = $4", d.Names.Channels)
	_, err := d.DB.Exec(
		query,
		time.Now(),
		user,
		channel,
		topic,
	)
	return err
}

func (d *DataBase) containsChannel(channel string) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE channel=$1", d.Names.Channels)
	row := d.DB.QueryRow(
		query,
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
	query := fmt.Sprintf("INSERT INTO %s (nickname, link, channel, topic, summary) VALUES ($1,$2,$3,$4,$5)", d.Names.Messages)
	_, err := d.DB.Exec(
		query,
		message.User,
		strconv.Itoa(message.Link),
		message.Channel,
		message.Topic,
		message.Summary,
	)

	return err
}

func (d *DataBase) getDelayedMessages(user string) ([]Message, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE nickname = $1", d.Names.Messages)
	rows, err := d.DB.Query(
		query,
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

	query = fmt.Sprintf("DELETE FROM %s WHERE nickname = $1", d.Names.Messages)
	_, err = d.DB.Exec(
		query,
		user,
	)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (d *DataBase) isPaused(user string) (bool, error) {
	query := fmt.Sprintf("SELECT paused FROM %s WHERE nickname = $1", d.Names.Users)
	row := d.DB.QueryRow(
		query,
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
	query := fmt.Sprintf("UPDATE %s SET paused = $1 WHERE nickname = $2 ", d.Names.Users)
	_, err = d.DB.Exec(
		query,
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
	query := fmt.Sprintf("UPDATE %s SET paused = $1 WHERE  nickname = $2 ", d.Names.Users)
	_, err = d.DB.Exec(
		query,
		false,
		user,
	)
	fmt.Println(user)
	return err
}

func (d *DataBase) getID(user string) (int64, error) {
	query := fmt.Sprintf("SELECT id FROM %s WHERE nickname=$1", d.Names.Users)
	row := d.DB.QueryRow(
		query,
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
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE nickname=$1", d.Names.Users)
	row := d.DB.QueryRow(
		query,
		user,
	)
	var count int64
	if err := row.Scan(&count); err != nil || count != 0 {
		return err
	}

	query = fmt.Sprintf("INSERT INTO %s (id, nickname, paused) VALUES ($1,$2,$3)", d.Names.Users)
	_, err := d.DB.Exec(
		query,
		id,
		user,
		false,
	)

	return err
}
