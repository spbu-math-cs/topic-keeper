package main

import (
	"errors"
	"fmt"
	_ "runtime/debug"
	"testing"
	"time"
)

type operationType string

const (
	Add    operationType = "add"
	Remove operationType = "remove"
	Get    operationType = "get"
	Set    operationType = "set"
)

type testOperation struct {
	operation operationType
	p1        string
	p2        string
	p3        string
	property  Property
}

type expected struct {
	something any
	err       error
}

func showError(t *testing.T, args testOperation) {
	t.Errorf("error in %s oparation with args: %s, %s, %s, %s", args.operation, args.p1, args.p2, args.p3, args.property)
}

func TestUsersChannelsTopicsAdd(t *testing.T) {
	var users SafeStorage
	users = &UsersChannelsTopics{}
	users.create()
	for _, tc := range []struct {
		function testOperation
		exp      expected
	}{
		{testOperation{Add, "u1", "c1", "t1", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "u1", "c1", "t2", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "u1", "c2", "t1", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "u2", "c1", "t1", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "u2", "c2", "t2", Topic},
			expected{nil, nil},
		},
	} {
		err := users.add(tc.function.p1, tc.function.p2, tc.function.p3, tc.function.property)
		if !errors.Is(err, tc.exp.err) {
			showError(t, tc.function)
		}
	}
}

func TestChannelTopicCountAdd(t *testing.T) {
	var users SafeStorage
	users = &ChannelTopicCount{}
	users.create()
	for _, tc := range []struct {
		function testOperation
		exp      expected
	}{
		{testOperation{Add, "c1", "t1", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c1", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c1", "t1", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c2", "t2", "", Topic},
			expected{nil, nil},
		},
	} {
		err := users.add(tc.function.p1, tc.function.p2, tc.function.p3, tc.function.property)
		if !errors.Is(err, tc.exp.err) {
			showError(t, tc.function)
		}
	}

}

func TestChannelTopicTimeAdd(t *testing.T) {
	var users SafeStorage
	users = &ChannelTopicTime{interval: 60 * time.Second}
	users.create()
	for _, tc := range []struct {
		function testOperation
		exp      expected
	}{
		{testOperation{Add, "c1", "t1", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c1", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c1", "t1", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c2", "t2", "", Topic},
			expected{nil, nil},
		},
	} {
		err := users.add(tc.function.p1, tc.function.p2, tc.function.p3, tc.function.property)
		if !errors.Is(err, tc.exp.err) {
			showError(t, tc.function)
		}
	}

}

func TestUsersChannelsTopicsRemove(t *testing.T) {
	var users SafeStorage
	users = &UsersChannelsTopics{}
	users.create()
	for _, tc := range []struct {
		function testOperation
		exp      expected
	}{
		{testOperation{Add, "u1", "c1", "t1", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "u1", "c1", "t2", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "u1", "c2", "t1", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "u2", "c1", "t1", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "u2", "c2", "t2", Topic},
			expected{nil, nil},
		},

		{testOperation{Remove, "u2", "c2", "t2", Topic},
			expected{nil, nil},
		},

		{testOperation{Remove, "u2", "c2", "t2", Topic},
			expected{nil, invalidChannelError},
		},

		{testOperation{Remove, "u1", "c1", "t2", Topic},
			expected{nil, nil},
		},

		{testOperation{Remove, "u1", "c1", "t2", Topic},
			expected{nil, invalidTopicError},
		},

		{testOperation{Remove, "u1", "c1", "t1", Topic},
			expected{nil, nil},
		},

		{testOperation{Remove, "u1", "c1", "t1", Topic},
			expected{nil, invalidChannelError},
		},

		{testOperation{Remove, "u3", "c1", "t1", Topic},
			expected{nil, invalidUserError},
		},
	} {
		switch tc.function.operation {
		case Add:
			err := users.add(tc.function.p1, tc.function.p2, tc.function.p3, tc.function.property)
			if !errors.Is(err, tc.exp.err) {
				showError(t, tc.function)
			}
		case Remove:
			err := users.remove(tc.function.p1, tc.function.p2, tc.function.p3, tc.function.property)
			if !errors.Is(err, tc.exp.err) {
				showError(t, tc.function)
			}
		}

	}
}

func TestChannelTopicCountRemove(t *testing.T) {
	var users SafeStorage
	users = &ChannelTopicCount{}
	users.create()

	for _, tc := range []struct {
		function testOperation
		exp      expected
	}{
		{testOperation{Add, "c1", "t1", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c1", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c1", "t1", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Remove, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Remove, "c2", "t1", "", Topic},
			expected{nil, invalidTopicError},
		},

		{testOperation{Remove, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Remove, "c2", "t2", "", Topic},
			expected{nil, invalidChannelError},
		},
	} {
		switch tc.function.operation {
		case Add:
			err := users.add(tc.function.p1, tc.function.p2, tc.function.p3, tc.function.property)
			if !errors.Is(err, tc.exp.err) {
				showError(t, tc.function)
			}
		case Remove:
			err := users.remove(tc.function.p1, tc.function.p2, tc.function.p3, tc.function.property)
			if !errors.Is(err, tc.exp.err) {
				showError(t, tc.function)
			}
		}
	}

}

func TestTopicTimeRemove(t *testing.T) {
	var users SafeStorage
	users = &ChannelTopicTime{interval: 60 * time.Second}
	users.create()
	for _, tc := range []struct {
		function testOperation
		exp      expected
	}{
		{testOperation{Add, "c1", "t1", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c1", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c1", "t1", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Remove, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Remove, "c2", "t2", "", Topic},
			expected{nil, invalidChannelError},
		},

		{testOperation{Remove, "c1", "t3", "", Topic},
			expected{nil, invalidTopicError},
		},
	} {
		switch tc.function.operation {
		case Add:
			err := users.add(tc.function.p1, tc.function.p2, tc.function.p3, tc.function.property)
			if !errors.Is(err, tc.exp.err) {
				showError(t, tc.function)
			}
		case Remove:
			err := users.remove(tc.function.p1, tc.function.p2, tc.function.p3, tc.function.property)
			if !errors.Is(err, tc.exp.err) {
				showError(t, tc.function)
			}
		}
	}
}

func TestUsersChannelsTopicsGet(t *testing.T) {
	var users SafeStorage
	users = &UsersChannelsTopics{}
	users.create()

	var m1 = make(map[string]map[string]struct{})
	var m1u1 = make(map[string]struct{})
	m1u1["t1"] = struct{}{}
	var m1u2 = make(map[string]struct{})
	m1u1["t2"] = struct{}{}
	m1["u1"] = m1u1
	m1["u2"] = m1u2

	for _, tc := range []struct {
		function testOperation
		exp      expected
	}{
		{testOperation{Add, "u1", "c1", "t1", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "u1", "c1", "t2", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "u1", "c2", "t1", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "u2", "c1", "t1", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "u2", "c2", "t2", Topic},
			expected{nil, nil},
		},

		{testOperation{Get, "c2", "", "", UsersTopicsByChannel},
			expected{m1, nil},
		},

		{testOperation{Get, "c1", "", "", UsersTopicsByChannel},
			expected{m1, nil},
		},
	} {
		switch tc.function.operation {
		case Add:
			err := users.add(tc.function.p1, tc.function.p2, tc.function.p3, tc.function.property)
			if !errors.Is(err, tc.exp.err) {
				showError(t, tc.function)
			}
		case Get:
			gotAny, err := users.get(tc.function.p1, tc.function.p2, tc.function.property)
			_, ok := (gotAny).(map[string]map[string]struct{})
			if !errors.Is(err, tc.exp.err) || !ok {
				showError(t, tc.function)
			}
		}
	}
}

func TestChannelTopicCountGet(t *testing.T) {
	var users SafeStorage
	users = &ChannelTopicCount{}
	users.create()
	for _, tc := range []struct {
		function testOperation
		exp      expected
	}{
		{testOperation{Add, "c1", "t1", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c1", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c1", "t1", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Get, "c1", "t1", "", Count},
			expected{uint64(2), nil},
		},

		{testOperation{Get, "c1", "t3", "", Count},
			expected{nil, invalidTopicError},
		},

		{testOperation{Get, "c3", "t3", "", Count},
			expected{nil, invalidChannelError},
		},

		{testOperation{Get, "c1", "t2", "", Count},
			expected{uint64(1), nil},
		},

		{testOperation{Get, "c2", "t2", "", Count},
			expected{uint64(2), nil},
		},
	} {
		switch tc.function.operation {
		case Add:
			err := users.add(tc.function.p1, tc.function.p2, tc.function.p3, tc.function.property)
			if !errors.Is(err, tc.exp.err) {
				showError(t, tc.function)
			}
		case Get:
			gotAny, err := users.get(tc.function.p1, tc.function.p2, tc.function.property)
			if !errors.Is(err, tc.exp.err) {
				showError(t, tc.function)
			}
			if err == nil {
				got, ok := (gotAny).(uint64)
				if !ok {
					showError(t, tc.function)
				}
				ex, ok := (tc.exp.something).(uint64)
				if !ok {
					showError(t, tc.function)
				}
				if got != ex {
					showError(t, tc.function)
				}
			}
		}
	}
}

func TestTopicTimeGet(t *testing.T) {
	var users SafeStorage
	users = &ChannelTopicTime{interval: time.Hour}
	users.create()
	for _, tc := range []struct {
		function testOperation
		exp      expected
	}{
		{testOperation{Add, "c1", "t1", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c1", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c1", "t1", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Get, "c3", "", "", Topics},
			expected{nil, invalidChannelError},
		},

		{testOperation{Get, "c1", "", "", Topics},
			expected{nil, nil},
		},

		{testOperation{Get, "c1", "", "", Topics},
			expected{2, nil},
		},

		{testOperation{Get, "c2", "", "", Topics},
			expected{1, nil},
		},
	} {
		switch tc.function.operation {
		case Add:
			err := users.add(tc.function.p1, tc.function.p2, tc.function.p3, tc.function.property)
			if !errors.Is(err, tc.exp.err) {
				showError(t, tc.function)
			}
		case Get:
			gotAny, err := users.get(tc.function.p1, "", tc.function.property)
			if !errors.Is(err, tc.exp.err) {
				showError(t, tc.function)
			}
			if err == nil {
				_, ok := (gotAny).([]string)
				if !ok {
					showError(t, tc.function)
				}
			}
		}

	}
}

func TestTopicTimeSet(t *testing.T) {
	var users SafeStorage
	users = &ChannelTopicTime{interval: time.Hour}
	users.create()
	for _, tc := range []struct {
		function testOperation
		exp      expected
	}{
		{testOperation{Add, "c1", "t1", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c1", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c1", "t1", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Add, "c2", "t2", "", Topic},
			expected{nil, nil},
		},

		{testOperation{Set, "c3", "t1", "", Time},
			expected{nil, invalidChannelError},
		},

		{testOperation{Set, "c1", "t3", "", Time},
			expected{nil, invalidTopicError},
		},

		{testOperation{Set, "c1", "t1", "", Time},
			expected{nil, nil},
		},
	} {
		switch tc.function.operation {
		case Add:
			err := users.add(tc.function.p1, tc.function.p2, tc.function.p3, tc.function.property)
			if !errors.Is(err, tc.exp.err) {
				showError(t, tc.function)
			}
		case Set:
			time.Sleep(time.Second)
			err := users.set(tc.function.p1, tc.function.p2, tc.function.property)
			if !errors.Is(err, tc.exp.err) {
				showError(t, tc.function)
			}
		}

	}

	fmt.Print("")
}
