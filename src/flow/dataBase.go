package main

import (
	"database/sql"
	"errors"
)

var (
	invalidArguments = errors.New("invalid argument type")
	invalidCondition = errors.New("invalid condition type")
)

type Message struct {
	User    string `json:"user"`
	Channel string `json:"channel"`
	Topic   string `json:"topic"`
	Summary string `json:"summary"`
}

type LocalStorage interface {
	add(user, channel, topic string) error
	removeTopic(user, channel, topic string) error
	removeChannel(user, channel string) error
	getTopics(channel string) ([]string, error)
	getUsers(channel string, topics []string) map[string][]string
	setTime(user, channel, topic string) error
	containsChannel(channel string) (bool, error)
	getDelayedMessages(user string) ([]Message, error)
}

type Table struct {
	DataBase *sql.DB
	Name     string
}

type DataBase struct {
	Channels        Table
	Users           Table
	DelayedMessages Table
}

func (d DataBase) add(user, channel, topic string) error {
	//TODO implement me
	panic("implement me")
}

func (d DataBase) removeTopic(user, channel, topic string) error {
	//TODO implement me
	panic("implement me")
}

func (d DataBase) removeChannel(user, channel string) error {
	//TODO implement me
	panic("implement me")
}

func (d DataBase) getTopics(channel string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (d DataBase) getUsers(channel string, topics []string) map[string][]string {
	//TODO implement me
	panic("implement me")
}

func (d DataBase) setTime(user, channel, topic string) error {
	//TODO implement me
	panic("implement me")
}

func (d DataBase) containsChannel(channel string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (d DataBase) getDelayedMessages(user string) ([]Message, error) {
	//TODO implement me
	panic("implement me")
}
