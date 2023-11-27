package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type Application = string

const (
	Delay                  = 10 * time.Second
	Telegram   Application = "telegram"
	VK         Application = "vk"
	MatterMost Application = "matter most"
)

func getUsingApplications() []Application {
	return []Application{VK, Telegram}
}

type Message struct {
	Application Application `json:"application"`
	User        string      `json:"user"`
	Link        string      `json:"link"`
	Channel     string      `json:"channel"`
	Topic       string      `json:"topic"`
	Summary     string      `json:"summary"`
}

type LocalStorage interface {
	addUser(user string, id int64) error
	addTopic(user, channel, topic string, application Application) error
	removeTopic(user, channel, topic string, application Application) error
	removeChannel(user, channel string, application Application) error
	getTopics(channel string, application Application) ([]string, error)
	getUserInfo(user string) (map[Application]map[string][]string, error)
	getUsers(channel string, topics []string, application Application) (map[string][]string, error)
	setTime(user, channel, topic string, application Application) error
	containsChannel(channel string, application Application) (bool, error)
	addDelayedMessage(messages Message) error
	getDelayedMessages(user string) ([]Message, error)
	pauseUser(user string) error
	unpauseUser(user string) error
	isPaused(user string) (bool, error)
	getID(user string) (int64, error)
	getVKPublicNameByID(groupID string) (string, error)
	addVKPublic(groupName, groupId string, postID int) error
	getVKPublic() ([]string, error)
	updateVKLastPostID(groupID string, postID int) error
	getVKLastPostID(groupID string) (int, error)
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
	VKPostID string
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

func (d *DataBase) addTopic(user, channel, topic string, application Application) error {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE nickname=$1 AND channel=$2 AND topic=$3 AND application=$4", d.Names.Channels)
	row := d.DB.QueryRow(query,
		user, channel, topic, application)

	var count int64
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	if count != 0 {
		return nil
	}

	query = fmt.Sprintf("INSERT INTO %s (nickname, channel, topic, last_time, application) VALUES ($1,$2,$3,$4,$5)", d.Names.Channels)
	_, err = d.DB.Exec(
		query,
		user,
		channel,
		topic,
		time.Now().Add(-Delay),
		application,
	)

	return err
}

func (d *DataBase) removeTopic(user, channel, topic string, application Application) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE nickname = $1 AND channel = $2 AND topic = $3 AND application = $4", d.Names.Channels)
	_, err := d.DB.Exec(
		query,
		user,
		channel,
		topic,
		application,
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *DataBase) removeChannel(user, channel string, application Application) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE nickname = $1 AND channel = $2 AND application = $3", d.Names.Channels)
	_, err := d.DB.Exec(
		query,
		user,
		channel,
		application,
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *DataBase) getTopics(channel string, application Application) ([]string, error) {
	query := fmt.Sprintf("SELECT topic FROM %s WHERE channel = $1 AND application = $2", d.Names.Channels)
	rows, err := d.DB.Query(
		query,
		channel,
		application,
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

func (d *DataBase) getUserInfo(user string) (map[Application]map[string][]string, error) {
	queryVK := fmt.Sprintf("SELECT groupid, public_name FROM %s", d.Names.VKPostID)
	rows, err := d.DB.Query(queryVK)
	if err != nil {
		return nil, err
	}
	vkMap := make(map[string]string)
	for rows.Next() {
		var id int
		var publicName string
		err = rows.Scan(&id, &publicName)
		sID := fmt.Sprintf("%d", id)
		vkMap[sID] = publicName
		log.Println(publicName, id)
	}

	query := fmt.Sprintf("SELECT channel, topic, application FROM %s WHERE nickname = $1", d.Names.Channels)
	rows, err = d.DB.Query(
		query,
		user,
	)
	if err != nil {
		return nil, err
	}

	answer := make(map[Application]map[string][]string)
	for _, application := range getUsingApplications() {
		curApplicationMap := make(map[string][]string)
		answer[application] = curApplicationMap
	}

	for rows.Next() {
		var channel, topic string
		var application Application
		err = rows.Scan(&channel, &topic, &application)
		if err != nil {
			return nil, err
		}
		switch application {
		case VK:
			channel = vkMap[channel]
			curTopics := answer[VK][channel]
			curTopics = append(curTopics, topic)
			answer[VK][channel] = curTopics
		case Telegram:
			curTopics := answer[Telegram][channel]
			curTopics = append(curTopics, topic)
			answer[Telegram][channel] = curTopics
		}
	}

	return answer, nil
}

func (d *DataBase) getUsers(channel string, topics []string, application Application) (map[string][]string, error) {
	answer := make(map[string][]string)
	for _, topic := range topics {
		query := fmt.Sprintf("SELECT nickname FROM %s WHERE channel = $1 AND topic = $2 AND last_time < $3 AND application = $4 ", d.Names.Channels)
		rows, err := d.DB.Query(
			query,
			channel,
			topic,
			time.Now().Add(-Delay),
			application,
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

func (d *DataBase) setTime(user, channel, topic string, application Application) error {
	query := fmt.Sprintf("UPDATE %s SET last_time = $1 WHERE  nickname = $2 AND channel = $3 AND topic = $4 AND application = $5", d.Names.Channels)
	_, err := d.DB.Exec(
		query,
		time.Now(),
		user,
		channel,
		topic,
		application,
	)
	return err
}

func (d *DataBase) containsChannel(channel string, application Application) (bool, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE channel=$1 AND application = $2", d.Names.Channels)
	row := d.DB.QueryRow(
		query,
		channel,
		application,
	)
	var count int64
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	return count != 0, nil
}

func (d *DataBase) addDelayedMessage(message Message) error {
	query := fmt.Sprintf("INSERT INTO %s (nickname, link, channel, topic, summary, application) VALUES ($1,$2,$3,$4,$5,$6)", d.Names.Messages)
	_, err := d.DB.Exec(
		query,
		message.User,
		message.Link,
		message.Channel,
		message.Topic,
		message.Summary,
		message.Application,
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
		err = rows.Scan(&message.User, &message.Link, &message.Channel, &message.Topic, &message.Summary, &message.Application)
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

func (d *DataBase) getVKPublic() ([]string, error) {
	query := fmt.Sprintf("SELECT channel FROM %s WHERE application=$1 GROUP BY channel", d.Names.Channels)
	rows, err := d.DB.Query(
		query,
		VK,
	)

	if err != nil {
		return nil, err
	}

	var ansRows []string

	for rows.Next() {
		var row string
		err = rows.Scan(&row)
		if err != nil {
			return nil, err
		}
		ansRows = append(ansRows, row)
	}

	return ansRows, nil
}

func (d *DataBase) updateVKLastPostID(groupID string, postID int) error {
	query := fmt.Sprintf(
		`INSERT INTO %s (groupid, last_post) VALUES ($1, $2) 
				ON CONFLICT (groupid) DO UPDATE SET last_post =  $3`,
		d.Names.VKPostID)
	_, err := d.DB.Exec(
		query,
		groupID,
		postID,
		postID,
	)
	return err
}

func (d *DataBase) getVKLastPostID(groupID string) (int, error) {
	query := fmt.Sprintf("SELECT last_post FROM %s WHERE groupID=$1", d.Names.VKPostID)
	row := d.DB.QueryRow(
		query,
		groupID,
	)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (d *DataBase) addVKPublic(groupName, groupID string, postID int) error {
	query := fmt.Sprintf(
		`INSERT INTO %s (groupid, last_post, public_name) VALUES ($1, $2, $3) 
				ON CONFLICT (groupid) DO UPDATE SET public_name =  $4`,
		d.Names.VKPostID)
	_, err := d.DB.Exec(
		query,
		groupID,
		postID,
		groupName,
		groupName,
	)
	return err
}

func (d *DataBase) getVKPublicNameByID(groupID string) (string, error) {
	query := fmt.Sprintf("SELECT public_name FROM %s WHERE groupid = $1", d.Names.VKPostID)
	row := d.DB.QueryRow(query, groupID)

	var name string
	err := row.Scan(&name)
	if err != nil {
		return "", err
	}

	return name, nil

}
