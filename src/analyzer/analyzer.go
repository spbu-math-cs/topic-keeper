package main

import "strings"

type BasicTextAnalyzer interface {
	contains(text string, keyWord string) float64
}

type Analyzer struct{}

func (a *Analyzer) contains(text string, keyWord string) float64 {
	text = strings.ToLower(text)
	keyWord = strings.ToLower(keyWord)
	if strings.Contains(text, keyWord) {
		return 1.0
	}
	return 0.0
}

func (a *Analyzer) analyze(topics map[string]struct{}, message string) ([]string, error) {
	var answer []string
	for topic, _ := range topics {
		if a.contains(message, topic) != 0.0 {
			answer = append(answer, topic)
		}
	}
	return answer, nil
}
