package main

import (
	_ "runtime/debug"
	"testing"
)

func TestPositive(t *testing.T) {

	for _, tc := range []struct {
		given  string
		answer string
	}{
		{"@channel", "channel"},
		{"https://t.me/channel", "channel"},
		{"@двач", "двач"},
		{"https://t.me/двач", "двач"},
		{"https://t.me/12345", "12345"},
		{"@123423", "123423"},
		{"@", ""},
		{"https://t.me/", ""},
	} {
		res, err := parseChannelName(tc.given)
		if err != nil {
			t.Errorf("Error (%s)in parsing: %s", err.Error(), tc.given)
		}
		if res != tc.answer {
			t.Errorf("Got %s but answer is %s", res, tc.answer)
		}
	}
}

func TestNegative(t *testing.T) {

	for _, tc := range []struct {
		given string
	}{
		{"dfssd@dfsd"},
		{"dfssd@"},
		{"1233423"},
		{"https://t.me@dfsd"},
		{"htps://t.me/"},
		{"https:/t.me/"},
		{"https:/t.m/"},
		{"https:/tme/"},
		{"https:/t.mr/"},
		{"https:/T.mr/"},
		{"https:|T.mr|"},
	} {
		_, err := parseChannelName(tc.given)
		if err == nil {
			t.Errorf("Error (%s)in parsing: %s", err.Error(), tc.given)
		}
	}
}
