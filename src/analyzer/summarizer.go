package main

import "math"

type Summarizer interface {
	summarize(text string) (string, error)
}

type MessageSummarizer struct {
	textLength int
}

func (m MessageSummarizer) summarize(text string) (string, error) {

	length := math.Min(float64(m.textLength), float64(len(text)))

	return text[:int(length)-1], nil
}
