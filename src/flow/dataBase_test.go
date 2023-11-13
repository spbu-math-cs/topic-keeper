package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v2"
	_ "runtime/debug"
	"testing"
)

var names = TablesNames{
	Channels: "channels_test",
	Users:    "users_test",
	Messages: "message_test",
}

//go:embed migrations/init_test.sql
var initScriptTest string

func NewTestDatabase(cfg DBConfig, names TablesNames) (*DataBase, error) {
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port,
		cfg.DB)
	db, err := sql.Open("pgx", url)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(initScriptTest); err != nil {
		return nil, err
	}
	return &DataBase{db, names}, nil
}

func TestAddUser(t *testing.T) {

	var dbConfig DBConfig
	if err := yaml.Unmarshal(rawDBConfig, &dbConfig); err != nil {
		panic(err)
	}
	var err error
	testBase, err := NewTestDatabase(dbConfig, names)
	if err != nil {
		t.Errorf("error in creating data base with error :[%s] \n", err.Error())
	}

	_, err = testBase.DB.Exec(fmt.Sprintf("DELETE FROM %s;", testBase.Names.Users))
	if err != nil {
		t.Errorf("error in delete from data base with :[%s] \n", err.Error())
	}

	/// addUser Test
	for i, tc := range []struct {
		user string
		id   int64
		sub  string
	}{
		{"user1", 0, fmt.Sprintf("SELECT COUNT(*) FROM %s", names.Users)},
		{"user2", 1, fmt.Sprintf("SELECT COUNT(*) FROM %s", names.Users)},
		{"user3", 2, fmt.Sprintf("SELECT COUNT(*) FROM %s", names.Users)},
		{"user4", 3, fmt.Sprintf("SELECT COUNT(*) FROM %s", names.Users)},
		{"user5", 4, fmt.Sprintf("SELECT COUNT(*) FROM %s", names.Users)},
		{"user6", 5, fmt.Sprintf("SELECT COUNT(*) FROM %s", names.Users)},
	} {
		err = testBase.addUser(tc.user, tc.id)
		if err != nil {
			t.Errorf("error in adding User :[%s] \n", err.Error())
		}
		row := testBase.DB.QueryRow(tc.sub)
		var count int
		if err := row.Scan(&count); err != nil {
			t.Errorf("error in scan with error :[%s] \n", err.Error())
		}
		if count != (i + 1) {
			t.Errorf("wrong user count in data base :[%s] \n", err.Error())
		}
	}
}
