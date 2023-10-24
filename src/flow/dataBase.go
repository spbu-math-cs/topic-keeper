package main

import (
	"database/sql"
	"errors"
	"time"
)

var (
	invalidArguments = errors.New("invalid argument type")
	invalidCondition = errors.New("invalid condition type")
)

type Condition string

const (
	User    Condition = "user"
	Channel Condition = "channel"
	Topic   Condition = "topic"
)

type Request struct {
	User      string
	Channel   string
	Topic     string
	Time      time.Time
	Condition Condition
}

type Storage interface {
	add(parameters any) error
	remove(parameters any) error
	get(parameters any) (any, error)
	set(parameters any) error
}

type DataBase struct {
	dataBase  *sql.DB
	TableName string
}

func (d *DataBase) add(parameters any) error {
	arg, ok := parameters.(Request)
	if !ok {
		return invalidArguments
	}
	_, err := d.dataBase.Exec(
		"INSERT INTO $1 (user, channel, topic, time) VALUES ($2, $3, $4, $5, $6)",
		d.TableName,
		arg.User,
		arg.Channel,
		arg.Topic,
		arg.Time,
		arg.Condition,
	)
	if err != nil {
		return err
	}
	return nil
}

func (d *DataBase) remove(parameters any) error {
	arg, ok := parameters.(Request)
	if !ok {
		return invalidArguments
	}
	switch arg.Condition {
	case User:
		_, err := d.dataBase.Exec(
			"DELETE FROM $1 WHERE user = $2",
			d.TableName,
			arg.User,
		)
		return err
	case Channel:
		_, err := d.dataBase.Exec(
			"DELETE FROM $1 WHERE user = $2 AND channel = $3",
			d.TableName,
			arg.User,
			arg.User,
		)
		return err
	case Topic:
		_, err := d.dataBase.Exec(
			"DELETE FROM $1 WHERE user = $2 AND channel = $3 AND topic = $4",
			d.TableName,
			arg.User,
			arg.User,
			arg.Topic,
		)
		return err
	default:
		return invalidCondition
	}
}

func (d *DataBase) get(parameters any) (any, error) {
	//TODO implement me
	panic("implement me")
}

func (d *DataBase) set(parameters any) error {
	//TODO implement me
	panic("implement me")
}
