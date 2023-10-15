package main

import "strings"

type BasicTextAnalyzer interface {
	analyze(topics map[string]struct{}, message string) ([]string, error)
	//contains(text string, keyword string) float64
}

type Analyzer struct{}

func (a *Analyzer) contains(text string, keyword string) float64 {
	text = strings.ToLower(text)
	keyword = strings.ToLower(keyword)
	if strings.Contains(text, keyword) {
		return 1.0
	}
	return 0.0
}

func (a *Analyzer) analyze(topics map[string]struct{}, message string) ([]string, error) {
	var answer []string
	for topic := range topics {
		if a.contains(message, topic) != 0.0 {
			answer = append(answer, topic)
		}
	}
	return answer, nil
}
