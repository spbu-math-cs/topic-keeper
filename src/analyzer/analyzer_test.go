package main

import (
	_ "runtime/debug"
	"testing"
)

func TestContainsKeyWord(t *testing.T) {

	var analyzer Analyzer

	for _, tc := range []struct {
		input string
		sub   string
	}{
		{"(x))!!№;%:?:%;№;%:?", "x"},
		{"Хей, ребятки! Не подскажете, когда у нас дедлайн по ПИ? А то я люблю все до последней ночи откладывать.", "дедлайн"},
		{"Хей, ребятки! Не подскажете, когда у нас дедлайн по ПИ? А то я люблю все до последней ночи откладывать.", "ПИ"},
		{"Дедлайны эти уже надоели, честно говоря", "дедлайн"},
		{"Елизавета, добрый вечер", "Елизавета"},
	} {
		res := analyzer.contains(tc.input, tc.sub)
		if res != 1.0 {
			t.Errorf("Didn't find %s in %s", tc.sub, tc.input)
		}
	}
}

func TestDoesntContainKeyWord(t *testing.T) {
	var analyzer Analyzer

	for _, tc := range []struct {
		input string
		sub   string
	}{
		{"(x))", "xy"},
		{"Хей, ребятки! Не подскажете, когда у нас дедлайн по ПИ? А то я люблю все до последней ночи откладывать.", "деадлайн"},
		{"Хей, ребятки! Не подскажете, когда у нас дедлайн по ПИ? А то я люблю все до последней ночи откладывать.", "ТИ"},
		{"Дедлайны эти уже надоели, честно говоря", "Дедлайнч"},
		{"Елизавета, добрый вечер", "Елизавета Сергеевна"},
	} {
		res := analyzer.contains(tc.input, tc.sub)
		if res != 0.0 {
			t.Errorf("Didn't find %s in %s", tc.sub, tc.input)
		}
	}
}
